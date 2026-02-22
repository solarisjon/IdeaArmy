package claude

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/yourusername/ai-agent-team/internal/llm"
)

const (
	DefaultAPIURL = "https://api.anthropic.com/v1/messages"
	DefaultModel  = "claude-sonnet-4-20250514"
)

// Client represents an Anthropic Claude API client
type Client struct {
	APIKey  string
	Model   string
	BaseURL string
	client  *http.Client
}

// NewClient creates a new Claude API client
func NewClient(apiKey string) *Client {
	if apiKey == "" {
		// Check both ANTHROPIC_API_KEY and ANTHROPIC_KEY
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("ANTHROPIC_KEY")
		}
	}

	return &Client{
		APIKey:  apiKey,
		Model:   DefaultModel,
		BaseURL: DefaultAPIURL,
		client:  &http.Client{},
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

// SimpleQuery sends a simple query and returns the response
func (c *Client) SimpleQuery(query string, systemPrompt string) (string, error) {
	messages := []Message{
		{
			Role:    "user",
			Content: query,
		},
	}
	return c.SendMessage(messages, systemPrompt, 0.7)
}
