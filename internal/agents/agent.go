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
}

// GetRole returns the agent's role
func (a *BaseAgent) GetRole() models.AgentRole {
	return a.Role
}

// GetName returns the agent's name
func (a *BaseAgent) GetName() string {
	return a.Name
}

// GetSystemPrompt returns the agent's system prompt
func (a *BaseAgent) GetSystemPrompt() string {
	return a.SystemPrompt
}

// Query sends a query to Claude with the agent's context
func (a *BaseAgent) Query(query string) (string, error) {
	return a.Client.SimpleQuery(query, a.SystemPrompt)
}

// QueryWithTokens sends a query with custom max tokens
func (a *BaseAgent) QueryWithTokens(query string, maxTokens int) (string, error) {
	messages := []llm.Message{
		{
			Role:    "user",
			Content: query,
		},
	}
	return a.Client.SendMessageWithTokens(messages, a.SystemPrompt, a.Temperature, maxTokens)
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
