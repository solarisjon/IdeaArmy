package llm

// Client is the common interface for all LLM backends.
// Both the Anthropic (Claude) and OpenAI-compatible clients implement this.
type Client interface {
	// SendMessage sends a conversation to the LLM and returns the assistant response text.
	SendMessage(messages []Message, systemPrompt string, temperature float64) (string, error)

	// SendMessageWithTokens sends a conversation with a custom max-token limit.
	SendMessageWithTokens(messages []Message, systemPrompt string, temperature float64, maxTokens int) (string, error)

	// SimpleQuery is a convenience wrapper: single user message with a system prompt.
	SimpleQuery(query string, systemPrompt string) (string, error)
}

// Message is the portable message format shared across backends.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
