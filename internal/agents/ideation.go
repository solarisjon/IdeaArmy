package agents

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/yourusername/ai-agent-team/internal/llm"
	"github.com/yourusername/ai-agent-team/internal/models"
	"strings"
)

// IdeationAgent generates creative ideas
type IdeationAgent struct {
	*BaseAgent
}

// NewIdeationAgent creates a new ideation agent
func NewIdeationAgent(client llm.Client) *IdeationAgent {
	systemPrompt := `You are the Ideation Agent, a creative thinker specialized in generating innovative ideas.

Your responsibilities:
- Generate creative, well-thought-out ideas based on the topic
- Research and reference existing knowledge, trends, and best practices
- Think deeply about concepts from multiple angles
- Explore unconventional approaches and solutions
- Build upon previous ideas in the discussion
- Provide detailed explanations for each idea

Your approach:
- Consider both practical and innovative solutions
- Draw from various domains and disciplines
- Think about user needs, technical feasibility, and market potential
- Generate ideas that are specific and actionable
- Support ideas with reasoning and examples

When generating ideas, structure them as JSON with:
{
  "ideas": [
    {
      "title": "Brief catchy title",
      "description": "Detailed description explaining the concept",
      "category": "Category or domain of the idea"
    }
  ]
}

Be creative, thorough, and insightful. Quality over quantity - each idea should be well-developed.`

	return &IdeationAgent{
		BaseAgent: &BaseAgent{
			Role:         models.RoleIdeation,
			Name:         "Ideation Specialist",
			SystemPrompt: systemPrompt,
			Client:       client,
			Temperature:  0.9, // Higher temperature for creativity
		},
	}
}

// Process generates ideas based on input
func (a *IdeationAgent) Process(context *models.Discussion, input string) (*models.AgentResponse, error) {
	discussionContext := BuildContext(context)

	query := fmt.Sprintf(`%s

Task: %s

Generate 3-5 creative, well-researched ideas. Think deeply about the concepts, their validity, and potential impact. Return your response as JSON following the specified format.`,
		discussionContext, input)

	response, err := a.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ideation query failed: %w", err)
	}

	// Try to extract and parse ideas from the response
	ideas := a.extractIdeas(response, context)

	return &models.AgentResponse{
		AgentRole: a.Role,
		Content:   response,
		Ideas:     ideas,
	}, nil
}

// extractIdeas attempts to parse ideas from the response
func (a *IdeationAgent) extractIdeas(response string, discussion *models.Discussion) []models.Idea {
	var ideas []models.Idea

	// Try to find JSON in the response
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")

	if startIdx == -1 || endIdx == -1 {
		return ideas
	}

	jsonStr := response[startIdx : endIdx+1]

	var parsed struct {
		Ideas []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Category    string `json:"category"`
		} `json:"ideas"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		// If JSON parsing fails, return empty slice
		return ideas
	}

	for _, idea := range parsed.Ideas {
		ideas = append(ideas, models.Idea{
			ID:          uuid.New().String(),
			Title:       idea.Title,
			Description: idea.Description,
			Category:    idea.Category,
			CreatedBy:   string(a.Role),
			Validated:   false,
			Score:       0,
		})
	}

	return ideas
}
