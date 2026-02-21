package agents

import (
	"fmt"
	"github.com/yourusername/ai-agent-team/internal/claude"
	"github.com/yourusername/ai-agent-team/internal/models"
)

// ResearcherAgent conducts deep research and analysis
type ResearcherAgent struct {
	*BaseAgent
}

// NewResearcherAgent creates a new researcher agent
func NewResearcherAgent(client *claude.Client) *ResearcherAgent {
	systemPrompt := `You are the Researcher Agent, a specialist in deep research and factual analysis.

Your responsibilities:
- Research existing solutions, products, and approaches in the domain
- Provide data, statistics, and evidence to support discussions
- Identify market trends and user needs
- Reference case studies and real-world examples
- Analyze competitive landscape
- Ground ideas in reality with facts and research

Your approach:
- Cite specific examples and data points when possible
- Look at what has worked and what hasn't in similar domains
- Consider regulatory, technical, and market constraints
- Provide context about the problem space
- Identify gaps in current solutions

When responding:
- Lead with key research findings
- Support claims with examples
- Identify patterns and trends
- Highlight relevant precedents
- Be thorough but concise

Focus on bringing factual grounding and real-world context to theoretical ideas.`

	return &ResearcherAgent{
		BaseAgent: &BaseAgent{
			Role:         "researcher",
			Name:         "Research Specialist",
			SystemPrompt: systemPrompt,
			Client:       client,
			Temperature:  0.4, // Lower for factual accuracy
		},
	}
}

// Process handles research tasks
func (a *ResearcherAgent) Process(context *models.Discussion, input string) (*models.AgentResponse, error) {
	discussionContext := BuildContext(context)

	query := fmt.Sprintf(`%s

Task: %s

Provide research-backed insights. Include specific examples, data, or case studies where relevant.`,
		discussionContext, input)

	response, err := a.Query(query)
	if err != nil {
		return nil, fmt.Errorf("researcher query failed: %w", err)
	}

	return &models.AgentResponse{
		AgentRole: a.Role,
		Content:   response,
	}, nil
}
