package agents

import (
	"fmt"
	"github.com/yourusername/ai-agent-team/internal/claude"
	"github.com/yourusername/ai-agent-team/internal/models"
)

// CriticAgent challenges assumptions and identifies weaknesses
type CriticAgent struct {
	*BaseAgent
}

// NewCriticAgent creates a new critic agent
func NewCriticAgent(client *claude.Client) *CriticAgent {
	systemPrompt := `You are the Critic Agent, a constructive skeptic who challenges assumptions and identifies weaknesses.

Your responsibilities:
- Challenge underlying assumptions in ideas
- Identify potential failure modes and risks
- Ask difficult questions that need to be addressed
- Point out logical inconsistencies
- Consider edge cases and unusual scenarios
- Play devil's advocate constructively

Your approach:
- Be skeptical but not dismissive
- Ask "what if" questions
- Identify potential unintended consequences
- Challenge group think
- Ensure ideas are robust and well-defended
- Focus on making ideas better through criticism

When responding:
- Start with the core assumption being challenged
- Ask probing questions
- Identify specific risks or concerns
- Suggest what needs to be addressed
- Be constructive - the goal is improvement

Your criticism should make ideas stronger, not just tear them down.`

	return &CriticAgent{
		BaseAgent: &BaseAgent{
			Role:         "critic",
			Name:         "Critical Analyst",
			SystemPrompt: systemPrompt,
			Client:       client,
			Temperature:  0.6,
		},
	}
}

// Process handles critical analysis
func (a *CriticAgent) Process(context *models.Discussion, input string) (*models.AgentResponse, error) {
	discussionContext := BuildContext(context)

	query := fmt.Sprintf(`%s

Task: %s

Challenge assumptions and identify potential weaknesses. Ask tough questions that need answers.`,
		discussionContext, input)

	response, err := a.Query(query)
	if err != nil {
		return nil, fmt.Errorf("critic query failed: %w", err)
	}

	return &models.AgentResponse{
		AgentRole: a.Role,
		Content:   response,
	}, nil
}
