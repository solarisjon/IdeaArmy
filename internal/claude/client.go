package claude

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/yourusername/ai-agent-team/internal/llm"
)

const (
	DefaultAPIURL = "https://api.anthropic.com/v1/messages"
	DefaultModel  = "claude-sonnet-4-20250514"
)

// Client represents an Anthropic Claude API client.
// It implements llm.Client and llm.StreamingClient.
type Client struct {
	APIKey  string
	Model   string
	BaseURL string
	client  *http.Client
}

// NewClient creates a new Claude API client
func NewClient(apiKey string) *Client {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("ANTHROPIC_KEY")
		}
	}

	return &Client{
		APIKey:  apiKey,
		Model:   DefaultModel,
		BaseURL: DefaultAPIURL,
		client:  llm.NewHTTPClient(llm.DefaultTimeout),
	}
}

// Message is an alias for the portable llm.Message type.
type Message = llm.Message

// apiRequest represents an API request to the Anthropic API.
type apiRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	System      string    `json:"system,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// Response represents an API response
type Response struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// claudeStreamEvent is a parsed Anthropic SSE event.
type claudeStreamEvent struct {
	Type  string `json:"type"`
	Delta *struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
}

// SendMessage sends a message to Claude and returns the response
func (c *Client) SendMessage(messages []Message, systemPrompt string, temperature float64) (string, error) {
	return c.SendMessageWithTokens(messages, systemPrompt, temperature, 4096)
}

// SendMessageWithTokens sends a message with custom max tokens
func (c *Client) SendMessageWithTokens(messages []Message, systemPrompt string, temperature float64, maxTokens int) (string, error) {
	req := apiRequest{
		Model:       c.Model,
		MaxTokens:   maxTokens,
		Messages:    messages,
		Temperature: temperature,
		System:      systemPrompt,
	}
	return c.doRequest(req)
}

// SendMessageStream streams token-by-token via onChunk and returns the full response.
// Implements llm.StreamingClient.
func (c *Client) SendMessageStream(messages []Message, systemPrompt string, temperature float64, onChunk func(string)) (string, error) {
	req := apiRequest{
		Model:       c.Model,
		MaxTokens:   4096,
		Messages:    messages,
		Temperature: temperature,
		System:      systemPrompt,
		Stream:      true,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send streaming request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var sb strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		var event claudeStreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		if event.Type == "content_block_delta" && event.Delta != nil && event.Delta.Type == "text_delta" {
			text := event.Delta.Text
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

// SimpleQuery sends a simple query and returns the response
func (c *Client) SimpleQuery(query string, systemPrompt string) (string, error) {
	messages := []Message{{Role: "user", Content: query}}
	return c.SendMessage(messages, systemPrompt, 0.7)
}

func (c *Client) doRequest(req apiRequest) (string, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp Response
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}
	return apiResp.Content[0].Text, nil
}
