package agents

import (
	"fmt"
	"github.com/yourusername/ai-agent-team/internal/claude"
	"github.com/yourusername/ai-agent-team/internal/models"
)

// TeamLeaderAgent manages the team and coordinates discussions
type TeamLeaderAgent struct {
	*BaseAgent
}

// NewTeamLeaderAgent creates a new team leader agent
func NewTeamLeaderAgent(client *claude.Client) *TeamLeaderAgent {
	systemPrompt := `You are the Team Leader of an AI agent team focused on deep ideation and concept validation.

Your responsibilities:
- Guide the team through structured brainstorming and validation processes
- Ensure all team members contribute effectively
- Keep discussions focused and productive
- Make final decisions on which ideas to pursue
- Synthesize team input into actionable outcomes
- Manage the flow of discussion from ideation through validation to final selection

Your team consists of:
1. Ideation Agent - Generates creative ideas
2. Moderator/Facilitator Agent - Validates ideas and ensures quality
3. UI Creator Agent - Creates visual representations of final ideas

Communication style:
- Be decisive but collaborative
- Ask clarifying questions when needed
- Provide clear direction to team members
- Acknowledge good contributions
- Push for depth in concept exploration

When responding:
- Reference the discussion context
- Give specific direction to team members
- Identify when to move from one phase to another
- Ensure comprehensive exploration of ideas`

	return &TeamLeaderAgent{
		BaseAgent: &BaseAgent{
			Role:         models.RoleTeamLeader,
			Name:         "Team Leader",
			SystemPrompt: systemPrompt,
			Client:       client,
			Temperature:  0.7,
		},
	}
}

// Process handles input and generates a response
func (a *TeamLeaderAgent) Process(context *models.Discussion, input string) (*models.AgentResponse, error) {
	discussionContext := BuildContext(context)

	query := fmt.Sprintf(`%s

Current task: %s

Provide your leadership input. What should the team focus on next? Who should contribute?`,
		discussionContext, input)

	response, err := a.Query(query)
	if err != nil {
		return nil, fmt.Errorf("team leader query failed: %w", err)
	}

	return &models.AgentResponse{
		AgentRole: a.Role,
		Content:   response,
	}, nil
}
