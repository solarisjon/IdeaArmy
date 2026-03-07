package agents

import (
	"fmt"

	"github.com/yourusername/ai-agent-team/internal/llm"
	"github.com/yourusername/ai-agent-team/internal/models"
	"github.com/yourusername/ai-agent-team/internal/tools"
)

// ResearcherAgent conducts deep research and analysis
type ResearcherAgent struct {
	*BaseAgent
}

// NewResearcherAgent creates a new researcher agent
func NewResearcherAgent(client llm.Client) *ResearcherAgent {
	systemPrompt := `You are the Researcher Agent, a specialist in deep research and factual analysis.

Your responsibilities:
- Research existing solutions, products, and approaches in the domain
- Provide data, statistics, and evidence to support discussions
- Identify market trends and user needs
- Reference case studies and real-world examples
- Analyze competitive landscape
- Ground ideas in reality with facts and research

You have access to a web_search tool. Use it to find current, real-world data about the topic.
Search for: market size, recent news, notable examples, key players, and relevant statistics.
Aim to make 2-3 targeted searches to gather concrete evidence.

When responding:
- Lead with key research findings from your searches
- Cite sources by URL when available
- Support claims with specific data points
- Identify patterns and trends
- Highlight relevant precedents
- Be thorough but concise

Focus on bringing factual grounding and real-world context to theoretical ideas.`

	a := &ResearcherAgent{
		BaseAgent: &BaseAgent{
			Role:         "researcher",
			Name:         "Research Specialist",
			SystemPrompt: systemPrompt,
			Client:       client,
			Temperature:  0.4, // Lower for factual accuracy
		},
	}

	// Register web search tool — executor uses the agent's Notify for status messages
	a.RegisterTool(tools.WebSearchTool(), tools.WebSearchExecutor(func(msg string) {
		if a.Notify != nil {
			a.Notify(fmt.Sprintf("  📣 [researcher] %s", msg))
		}
	}))

	return a
}

// Process handles research tasks
func (a *ResearcherAgent) Process(context *models.Discussion, input string) (*models.AgentResponse, error) {
	discussionContext := BuildContext(context)

	query := fmt.Sprintf(`%s

Task: %s

Use the web_search tool to find current data and real-world examples. Then synthesize your findings into research-backed insights with specific sources.`,
		discussionContext, input)

	response, err := a.QueryWithTools(query)
	if err != nil {
		return nil, fmt.Errorf("researcher query failed: %w", err)
	}

	return &models.AgentResponse{
		AgentRole: a.Role,
		Content:   response,
	}, nil
}

// ResearcherAgent conducts deep research and analysis
