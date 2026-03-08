package agents

import (
	"fmt"
	"os"
	"strings"

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

When web search is available, use the web_search tool to find current, real-world data.
Search for: market size, recent news, notable examples, key players, and relevant statistics.
Aim to make 2-3 targeted searches to gather concrete evidence.

When web search is not available, draw deeply on your training knowledge to produce a
comprehensive, structured research brief. Be specific: name real companies, cite known figures,
describe actual products and case studies. Do not be vague — concrete detail is more valuable
than generic observations.

When responding:
- Lead with key research findings
- Cite sources or known references where possible
- Support claims with specific data points and named examples
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
			Temperature:  0.4,
		},
	}
}

// hasWebSearch returns true if a Firecrawl API key is configured.
func (a *ResearcherAgent) hasWebSearch() bool {
	return a.FirecrawlKey != "" || os.Getenv("FIRECRAWL_API_KEY") != ""
}

// Process handles research tasks, using live web search when available and
// falling back to a structured deep-knowledge synthesis when it is not.
func (a *ResearcherAgent) Process(context *models.Discussion, input string) (*models.AgentResponse, error) {
	discussionContext := BuildContext(context)

	if a.hasWebSearch() {
		return a.processWithWebSearch(discussionContext, input)
	}
	return a.processFromKnowledge(discussionContext, input)
}

// processWithWebSearch runs the researcher with live Firecrawl web search.
func (a *ResearcherAgent) processWithWebSearch(discussionContext, input string) (*models.AgentResponse, error) {
	var capturedResults []tools.SearchResult
	a.RegisterTool(tools.WebSearchTool(), tools.WebSearchExecutor(
		a.FirecrawlKey,
		func(msg string) {
			if a.Notify != nil {
				a.Notify(fmt.Sprintf("  📣 [researcher] %s", msg))
			}
		},
		func(results []tools.SearchResult) {
			capturedResults = append(capturedResults, results...)
		},
	))

	query := fmt.Sprintf(`%s

Task: %s

Use the web_search tool to find current data and real-world examples. Then synthesize your findings into research-backed insights with specific sources.`,
		discussionContext, input)

	response, err := a.QueryWithTools(query)
	if err != nil {
		return nil, fmt.Errorf("researcher query failed: %w", err)
	}

	var srIface []interface{}
	for _, r := range capturedResults {
		srIface = append(srIface, r)
	}

	return &models.AgentResponse{
		AgentRole:     a.Role,
		Content:       response,
		SearchResults: srIface,
	}, nil
}

// processFromKnowledge runs the researcher using structured LLM knowledge synthesis
// when no web search API is available. It requests a response in labelled sections
// and converts each section into a rich evidence card.
func (a *ResearcherAgent) processFromKnowledge(discussionContext, input string) (*models.AgentResponse, error) {
	query := fmt.Sprintf(`%s

Task: %s

Web search is not available. Produce a comprehensive research brief from your training knowledge,
structured with EXACTLY these section headers (use ## prefix):

## Market Overview
Describe the market size, growth trajectory, and overall landscape. Include specific figures and
timeframes where known.

## Key Players & Solutions
Name real companies, products, and open-source projects in this space. Describe what makes each
notable and their approximate market position.

## Real-World Examples & Case Studies
Give 3-5 concrete real-world implementations or deployments. Be specific about who, what, and
what outcomes were achieved.

## Data & Statistics
List specific numeric data points: adoption rates, cost figures, performance benchmarks, survey
results. Attribute each figure to its known source.

## Trends & Emerging Patterns
Describe what is changing in this space right now. What technologies, behaviours, or business
models are gaining momentum?

## Challenges & Gaps
What problems remain unsolved? What do practitioners consistently struggle with?

Be specific and name real entities throughout. Vague generalisations are not useful here.`,
		discussionContext, input)

	if a.Notify != nil {
		a.Notify("  📣 [researcher] 🧠 Synthesising from training knowledge (no web search available)")
	}

	response, err := a.Query(query)
	if err != nil {
		return nil, fmt.Errorf("researcher knowledge query failed: %w", err)
	}

	evidence := extractKnowledgeCards(response)

	return &models.AgentResponse{
		AgentRole:     a.Role,
		Content:       response,
		SearchResults: evidence,
	}, nil
}

// knowledgeSections are the expected section headers in the structured research brief.
var knowledgeSections = []string{
	"Market Overview",
	"Key Players & Solutions",
	"Real-World Examples & Case Studies",
	"Data & Statistics",
	"Trends & Emerging Patterns",
	"Challenges & Gaps",
}

// extractKnowledgeCards parses the structured research response into evidence cards,
// one per section, so the UI can display them as rich research findings.
func extractKnowledgeCards(response string) []interface{} {
	// Split on ## headings.
	sections := make(map[string]string)
	current := ""
	for _, line := range strings.Split(response, "\n") {
		if strings.HasPrefix(line, "## ") {
			current = strings.TrimPrefix(line, "## ")
			current = strings.TrimSpace(current)
			sections[current] = ""
		} else if current != "" {
			sections[current] += line + "\n"
		}
	}

	var cards []interface{}
	for _, heading := range knowledgeSections {
		body, ok := sections[heading]
		if !ok {
			// Fuzzy match: accept if the response heading contains the key word.
			for k, v := range sections {
				if strings.Contains(k, strings.Split(heading, " ")[0]) {
					body = v
					ok = true
					break
				}
			}
		}
		if !ok || strings.TrimSpace(body) == "" {
			continue
		}

		// First non-empty line becomes the snippet/description.
		desc := ""
		for _, l := range strings.Split(strings.TrimSpace(body), "\n") {
			l = strings.TrimSpace(l)
			if l != "" {
				desc = l
				break
			}
		}
		if len(desc) > 160 {
			desc = desc[:157] + "…"
		}

		cards = append(cards, map[string]interface{}{
			"title":       heading,
			"description": desc,
			"snippet":     strings.TrimSpace(body),
			"query":       heading,
		})
	}

	// Fallback: if parsing found nothing, surface a single "Training data" card.
	if len(cards) == 0 {
		cards = append(cards, map[string]interface{}{
			"title":       "Research from training data",
			"description": "Structured knowledge synthesis — add a Firecrawl API key for live web search.",
		})
	}

	return cards
}
