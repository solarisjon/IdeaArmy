package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/yourusername/ai-agent-team/internal/llm"
)

// Client implements llm.Client for OpenAI-compatible APIs.
// It also implements llm.StreamingClient and llm.ToolCallingClient.
type Client struct {
	APIKey  string
	Model   string
	BaseURL string // e.g. "https://llm-proxy-api.ai.eng.netapp.com/v1"
	User    string // optional "user" field sent in request body (required by some proxies)
	client  *http.Client
}

// NewClient creates a new OpenAI-compatible client.
func NewClient(apiKey, baseURL, model string) *Client {
	return &Client{
		APIKey:  apiKey,
		Model:   model,
		BaseURL: baseURL,
		client:  llm.NewHTTPClient(llm.DefaultTimeout),
	}
}

// chatRequest is the unified OpenAI chat completions request body.
type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []chatMsg `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature *float64  `json:"temperature,omitempty"`
	User        string    `json:"user,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	Tools       []apiTool `json:"tools,omitempty"`
	ToolChoice  string    `json:"tool_choice,omitempty"`
}

// chatMsg handles all OpenAI message roles: user, assistant, system, tool.
type chatMsg struct {
	Role       string        `json:"role"`
	Content    string        `json:"content,omitempty"`
	ToolCalls  []apiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
}

// apiTool is the OpenAI function tool definition format.
type apiTool struct {
	Type     string `json:"type"` // "function"
	Function struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Parameters  json.RawMessage `json:"parameters"`
	} `json:"function"`
}

// apiToolCall is the tool call reference in an assistant response.
type apiToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // "function"
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// chatResponse is the OpenAI chat completions response body.
type chatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Role      string        `json:"role"`
			Content   string        `json:"content"`
			ToolCalls []apiToolCall `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// streamChunk is a single SSE chunk from the streaming API.
type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content   string        `json:"content"`
			ToolCalls []apiToolCall `json:"tool_calls"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

func (c *Client) SendMessage(messages []llm.Message, systemPrompt string, temperature float64) (string, error) {
	return c.SendMessageWithTokens(messages, systemPrompt, temperature, 4096)
}

func (c *Client) SendMessageWithTokens(messages []llm.Message, systemPrompt string, temperature float64, maxTokens int) (string, error) {
	msgs := c.buildMsgs(systemPrompt, messages)
	req := chatRequest{
		Model:     c.Model,
		Messages:  msgs,
		MaxTokens: maxTokens,
		User:      c.User,
	}
	if temperature > 0 {
		req.Temperature = &temperature
	}
	return c.doRequest(req)
}

// SendMessageStream streams token-by-token via onChunk and returns the full response.
// Implements llm.StreamingClient.
func (c *Client) SendMessageStream(messages []llm.Message, systemPrompt string, temperature float64, onChunk func(string)) (string, error) {
	msgs := c.buildMsgs(systemPrompt, messages)
	req := chatRequest{
		Model:     c.Model,
		Messages:  msgs,
		MaxTokens: 4096,
		User:      c.User,
		Stream:    true,
	}
	if temperature > 0 {
		req.Temperature = &temperature
	}
	return c.doStream(req, onChunk)
}

// SendMessageWithTools sends a conversation with tool definitions and executes a
// tool-call loop (max 5 iterations) until the LLM returns a final text response.
// Implements llm.ToolCallingClient.
func (c *Client) SendMessageWithTools(messages []llm.Message, systemPrompt string, temperature float64, tools []llm.ToolDefinition, executeTool func(name, arguments string) (string, error)) (string, error) {
	msgs := c.buildMsgs(systemPrompt, messages)

	// Convert tools to OpenAI format
	apiTools := make([]apiTool, len(tools))
	for i, t := range tools {
		apiTools[i].Type = "function"
		apiTools[i].Function.Name = t.Name
		apiTools[i].Function.Description = t.Description
		apiTools[i].Function.Parameters = t.Parameters
	}

	for iter := 0; iter < 5; iter++ {
		req := chatRequest{
			Model:      c.Model,
			Messages:   msgs,
			MaxTokens:  4096,
			User:       c.User,
			Tools:      apiTools,
			ToolChoice: "auto",
		}
		if temperature > 0 {
			req.Temperature = &temperature
		}

		content, toolCalls, finishReason, err := c.doRequestFull(req)
		if err != nil {
			return "", err
		}

		if finishReason != "tool_calls" || len(toolCalls) == 0 {
			return content, nil
		}

		// Add assistant message (with its tool_calls) back to context
		msgs = append(msgs, chatMsg{Role: "assistant", ToolCalls: toolCalls})

		// Execute each tool call and add results
		for _, tc := range toolCalls {
			result, execErr := executeTool(tc.Function.Name, tc.Function.Arguments)
			if execErr != nil {
				result = fmt.Sprintf("tool error: %v", execErr)
			}
			msgs = append(msgs, chatMsg{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
	}
	return "", fmt.Errorf("tool call loop exceeded maximum iterations")
}

func (c *Client) SimpleQuery(query string, systemPrompt string) (string, error) {
	messages := []llm.Message{{Role: "user", Content: query}}
	return c.SendMessage(messages, systemPrompt, 0.7)
}

// buildMsgs converts llm.Messages to chatMsgs, prepending the system prompt.
func (c *Client) buildMsgs(systemPrompt string, messages []llm.Message) []chatMsg {
	var msgs []chatMsg
	if systemPrompt != "" {
		msgs = append(msgs, chatMsg{Role: "system", Content: systemPrompt})
	}
	for _, m := range messages {
		msgs = append(msgs, chatMsg{Role: m.Role, Content: m.Content})
	}
	return msgs
}

// doRequest performs a blocking (non-streaming) API call and returns the response text.
func (c *Client) doRequest(req chatRequest) (string, error) {
	content, _, _, err := c.doRequestFull(req)
	return content, err
}

// doRequestFull performs a blocking API call and returns content, tool calls, and finish reason.
func (c *Client) doRequestFull(req chatRequest) (content string, toolCalls []apiToolCall, finishReason string, err error) {
	body, statusCode, err := c.httpPost(req)
	if err != nil {
		return "", nil, "", err
	}

	if statusCode != http.StatusOK {
		bodyStr := string(body)
		// Retry without temperature for models that only support the default value
		if statusCode == 400 && req.Temperature != nil && strings.Contains(bodyStr, "temperature") {
			log.Printf("Model %s rejected temperature, retrying with default", c.Model)
			req.Temperature = nil
			body, statusCode, err = c.httpPost(req)
			if err != nil {
				return "", nil, "", err
			}
			if statusCode != http.StatusOK {
				return "", nil, "", fmt.Errorf("API error (status %d): %s", statusCode, string(body))
			}
		} else {
			return "", nil, "", fmt.Errorf("API error (status %d): %s", statusCode, bodyStr)
		}
	}

	var apiResp chatResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", nil, "", fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if len(apiResp.Choices) == 0 {
		return "", nil, "", fmt.Errorf("no choices in response")
	}

	choice := apiResp.Choices[0]
	return choice.Message.Content, choice.Message.ToolCalls, choice.FinishReason, nil
}

// doStream performs a streaming API call, calling onChunk for each token.
// Returns the full accumulated response text.
func (c *Client) doStream(req chatRequest, onChunk func(string)) (string, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal error: %w", err)
	}

	url := c.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("request create error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("streaming request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		// Retry without temperature
		if resp.StatusCode == 400 && req.Temperature != nil && strings.Contains(bodyStr, "temperature") {
			log.Printf("Model %s rejected temperature in stream, retrying", c.Model)
			req.Temperature = nil
			return c.doStream(req, onChunk)
		}
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, bodyStr)
	}

	var sb strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) > 0 {
			text := chunk.Choices[0].Delta.Content
			if text != "" {
				sb.WriteString(text)
				if onChunk != nil {
					onChunk(text)
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return sb.String(), fmt.Errorf("stream read error: %w", err)
	}
	return sb.String(), nil
}

// httpPost marshals the request and executes the HTTP POST, returning raw body and status code.
func (c *Client) httpPost(req chatRequest) ([]byte, int, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal error: %w", err)
	}

	url := c.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, 0, fmt.Errorf("request create error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, 0, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("read response failed: %w", err)
	}
	return body, resp.StatusCode, nil
}

// Client implements llm.Client for OpenAI-compatible APIs.
