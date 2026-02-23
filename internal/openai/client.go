package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yourusername/ai-agent-team/internal/llm"
)

const defaultTimeout = 120 * time.Second

// Client implements llm.Client for OpenAI-compatible APIs.
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
		client:  llm.NewHTTPClient(defaultTimeout),
	}
}

// chatRequest is the OpenAI chat completions request body.
type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []chatMsg `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature *float64  `json:"temperature,omitempty"`
	User        string    `json:"user,omitempty"`
}

type chatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse is the OpenAI chat completions response body.
type chatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (c *Client) SendMessage(messages []llm.Message, systemPrompt string, temperature float64) (string, error) {
	return c.SendMessageWithTokens(messages, systemPrompt, temperature, 4096)
}

func (c *Client) SendMessageWithTokens(messages []llm.Message, systemPrompt string, temperature float64, maxTokens int) (string, error) {
	var msgs []chatMsg

	// OpenAI uses a system message in the messages array
	if systemPrompt != "" {
		msgs = append(msgs, chatMsg{Role: "system", Content: systemPrompt})
	}
	for _, m := range messages {
		msgs = append(msgs, chatMsg{Role: m.Role, Content: m.Content})
	}

	req := chatRequest{
		Model:     c.Model,
		Messages:  msgs,
		MaxTokens: maxTokens,
		User:      c.User,
	}
	if temperature > 0 {
		req.Temperature = &temperature
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

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

	var apiResp chatResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return apiResp.Choices[0].Message.Content, nil
}

func (c *Client) SimpleQuery(query string, systemPrompt string) (string, error) {
	messages := []llm.Message{
		{Role: "user", Content: query},
	}
	return c.SendMessage(messages, systemPrompt, 0.7)
}
