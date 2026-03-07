package llm

import "encoding/json"

// ToolDefinition describes a callable tool exposed to the LLM.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"` // JSON Schema object
}

// ToolCall represents a tool invocation requested by the LLM.
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON-encoded arguments string
}

// StreamingClient is an optional interface for backends that support token streaming.
// Detect with: sc, ok := client.(llm.StreamingClient)
type StreamingClient interface {
	SendMessageStream(messages []Message, systemPrompt string, temperature float64, onChunk func(string)) (string, error)
}

// ToolCallingClient is an optional interface for backends that support tool/function calling.
// Detect with: tc, ok := client.(llm.ToolCallingClient)
type ToolCallingClient interface {
	SendMessageWithTools(messages []Message, systemPrompt string, temperature float64, tools []ToolDefinition, executeTool func(name, arguments string) (string, error)) (string, error)
}
