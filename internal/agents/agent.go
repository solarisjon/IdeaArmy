package agents

import (
	"fmt"

	"github.com/yourusername/ai-agent-team/internal/llm"
	"github.com/yourusername/ai-agent-team/internal/models"
)

// Agent interface defines the common behavior for all agents
type Agent interface {
	GetRole() models.AgentRole
	GetName() string
	GetModel() string
	Process(context *models.Discussion, input string) (*models.AgentResponse, error)
	GetSystemPrompt() string
}

// BaseAgent provides common functionality for all agents
type BaseAgent struct {
	Role         models.AgentRole
	Name         string
	SystemPrompt string
	Client       llm.Client
	Temperature  float64
	Model        string // LLM model identifier used by this agent

	// FirecrawlKey is the Firecrawl API key for web search. If empty, falls back to FIRECRAWL_API_KEY env var.
	FirecrawlKey string

	// OnChunk is called for each streaming token. Set by the orchestrator.
	OnChunk func(string)

	// Notify is called to send status messages (e.g. tool use). Set by the orchestrator.
	Notify func(string)

	tools         []llm.ToolDefinition
	toolExecutors map[string]func(args string) (string, error)
}

// GetRole returns the agent's role
func (a *BaseAgent) GetRole() models.AgentRole { return a.Role }

// GetName returns the agent's name
func (a *BaseAgent) GetName() string { return a.Name }

// GetSystemPrompt returns the agent's system prompt
func (a *BaseAgent) GetSystemPrompt() string { return a.SystemPrompt }

// GetModel returns the LLM model identifier this agent is using
func (a *BaseAgent) GetModel() string { return a.Model }

// RegisterTool registers a tool and its executor for this agent.
func (a *BaseAgent) RegisterTool(def llm.ToolDefinition, executor func(args string) (string, error)) {
	// Replace existing tool with same name to prevent duplicate declarations
	for i, t := range a.tools {
		if t.Name == def.Name {
			a.tools[i] = def
			if a.toolExecutors == nil {
				a.toolExecutors = make(map[string]func(args string) (string, error))
			}
			a.toolExecutors[def.Name] = executor
			return
		}
	}
	a.tools = append(a.tools, def)
	if a.toolExecutors == nil {
		a.toolExecutors = make(map[string]func(args string) (string, error))
	}
	a.toolExecutors[def.Name] = executor
}

// Query sends a query using the agent's system prompt (blocking).
func (a *BaseAgent) Query(query string) (string, error) {
	return a.Client.SimpleQuery(query, a.SystemPrompt)
}

// QueryWithTokens sends a query with custom max tokens (blocking).
func (a *BaseAgent) QueryWithTokens(query string, maxTokens int) (string, error) {
	messages := []llm.Message{{Role: "user", Content: query}}
	return a.Client.SendMessageWithTokens(messages, a.SystemPrompt, a.Temperature, maxTokens)
}

// QueryStream sends a query and streams tokens via the OnChunk callback.
// Falls back to a blocking Query if the client doesn't support streaming or OnChunk is nil.
func (a *BaseAgent) QueryStream(query string) (string, error) {
	messages := []llm.Message{{Role: "user", Content: query}}
	if a.OnChunk != nil {
		if sc, ok := a.Client.(llm.StreamingClient); ok {
			return sc.SendMessageStream(messages, a.SystemPrompt, a.Temperature, a.OnChunk)
		}
		// Client doesn't support streaming — run blocking and emit result as a single chunk
		result, err := a.Client.SendMessage(messages, a.SystemPrompt, a.Temperature)
		if err == nil {
			a.OnChunk(result)
		}
		return result, err
	}
	return a.Client.SendMessage(messages, a.SystemPrompt, a.Temperature)
}

// QueryWithTools sends a query using registered tools.
// If no tools are registered or the client doesn't support tool calling, falls back to QueryStream.
func (a *BaseAgent) QueryWithTools(query string) (string, error) {
	if len(a.tools) == 0 {
		return a.QueryStream(query)
	}
	if tc, ok := a.Client.(llm.ToolCallingClient); ok {
		messages := []llm.Message{{Role: "user", Content: query}}
		executor := func(name, args string) (string, error) {
			fn, exists := a.toolExecutors[name]
			if !exists {
				return "", fmt.Errorf("unknown tool: %s", name)
			}
			return fn(args)
		}
		return tc.SendMessageWithTools(messages, a.SystemPrompt, a.Temperature, a.tools, executor)
	}
	// Fall back to streaming query if tool calling is not supported
	return a.QueryStream(query)
}

// BuildContext creates a context string from the discussion history
func BuildContext(discussion *models.Discussion) string {
	if discussion == nil {
		return ""
	}

	context := fmt.Sprintf("Topic: %s\n\n", discussion.Topic)

	if len(discussion.Messages) > 0 {
		context += "Previous Discussion:\n"
		for _, msg := range discussion.Messages {
			context += fmt.Sprintf("[%s -> %s]: %s\n", msg.From, msg.To, msg.Content)
		}
		context += "\n"
	}

	if len(discussion.Ideas) > 0 {
		context += "Current Ideas:\n"
		for i, idea := range discussion.Ideas {
			context += fmt.Sprintf("%d. %s - %s\n", i+1, idea.Title, idea.Description)
			if idea.Validated {
				context += fmt.Sprintf("   Score: %.1f/10\n", idea.Score)
			}
		}
		context += "\n"
	}

	return context
}

// Agent interface defines the common behavior for all agents
