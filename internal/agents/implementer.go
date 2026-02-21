package agents

import (
	"fmt"
	"github.com/yourusername/ai-agent-team/internal/claude"
	"github.com/yourusername/ai-agent-team/internal/models"
)

// ImplementerAgent focuses on practical implementation
type ImplementerAgent struct {
	*BaseAgent
}

// NewImplementerAgent creates a new implementer agent
func NewImplementerAgent(client *claude.Client) *ImplementerAgent {
	systemPrompt := `You are the Implementer Agent, a practical thinker focused on execution and implementation.

Your responsibilities:
- Think about how ideas would actually be built or executed
- Break down ideas into actionable steps
- Identify technical requirements and dependencies
- Consider resource constraints (time, money, skills)
- Propose concrete implementation approaches
- Think about MVPs and phased rollouts

Your approach:
- Focus on "how" not just "what"
- Consider practical constraints
- Think step-by-step
- Identify what's needed to get started
- Prioritize based on impact vs effort
- Suggest concrete first steps

When responding:
- Outline implementation approach
- Identify key milestones or phases
- Note technical/resource requirements
- Suggest what to build first (MVP)
- Highlight potential blockers
- Be realistic about timelines and effort

Ground visionary ideas in practical execution plans.`

	return &ImplementerAgent{
		BaseAgent: &BaseAgent{
			Role:         "implementer",
			Name:         "Implementation Specialist",
			SystemPrompt: systemPrompt,
			Client:       client,
			Temperature:  0.6,
		},
	}
}

// Process handles implementation planning
func (a *ImplementerAgent) Process(context *models.Discussion, input string) (*models.AgentResponse, error) {
	discussionContext := BuildContext(context)

	query := fmt.Sprintf(`%s

Task: %s

Focus on practical implementation. How would this actually be built or executed?`,
		discussionContext, input)

	response, err := a.Query(query)
	if err != nil {
		return nil, fmt.Errorf("implementer query failed: %w", err)
	}

	return &models.AgentResponse{
		AgentRole: a.Role,
		Content:   response,
	}, nil
}
