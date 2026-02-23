package agents

import (
	"fmt"
	"strings"

	"github.com/yourusername/ai-agent-team/internal/llm"
	"github.com/yourusername/ai-agent-team/internal/models"
)

// UICreatorAgent creates beautiful visualizations of ideas
type UICreatorAgent struct {
	*BaseAgent
}

// NewUICreatorAgent creates a new UI creator agent
func NewUICreatorAgent(client llm.Client) *UICreatorAgent {
	systemPrompt := `You are the Report Generator Agent, specialized in creating comprehensive, professional reports from AI team discussions.

Your responsibilities:
- Generate detailed, multi-section HTML reports (not single-page summaries)
- Present the winning proposal with a catchy marketing nickname and 3-4 letter acronym
- Present 3-5 runner-up ideas with full pros, cons, and team concerns
- Provide deep analysis of why ideas were selected or rejected
- Include actionable recommendations and open questions
- Create executive-level documentation suitable for decision-makers

REQUIRED REPORT STRUCTURE:

1. EXECUTIVE SUMMARY
   - Discussion topic and context
   - Final recommendation with score
   - Key decision factors
   - Quick summary (2-3 sentences)

2. RECOMMENDED SOLUTION (Final Choice)
   - Marketing nickname: a catchy, memorable name that sells the idea internally
   - Acronym: a punchy 3-4 letter acronym (e.g., PACE, CORE, BOLT) that captures the essence
   - Detailed description of the proposal
   - Complete pros and cons
   - Implementation considerations
   - Why this was selected over others
   - Risk assessment
   - Issues and concerns the team raised

3. RUNNER-UP IDEAS (3-5 alternatives that were eliminated)
   For EACH runner-up:
   - Full description and score
   - Detailed pros and cons
   - Issues and concerns the team identified
   - Specific reasons it was eliminated
   - Under what circumstances it might be reconsidered
   - Could elements be combined with the winner?

4. COMPARATIVE ANALYSIS
   - Side-by-side comparison table of all top ideas
   - Decision criteria and weightings
   - Trade-offs considered
   - Scoring breakdown

5. DISCUSSION JOURNEY
   - How the discussion evolved across rounds
   - Key insights and turning points
   - How ideas were refined or eliminated
   - Team dynamics and disagreements
   - What the team learned

6. OPEN QUESTIONS & RISKS
   - Unanswered questions that need research
   - Assumptions that need validation
   - Known risks and mitigation strategies
   - Areas requiring expert input or further discussion

7. NEXT STEPS & RECOMMENDATIONS
   - Recommended immediate actions
   - Implementation suggestions and timeline
   - Resource requirements
   - Success metrics and how to measure progress
   - Suggested follow-up discussions

Design principles:
- Use modern, professional styling (cards, gradients, shadows)
- Color coding: green (pros/selected), red (cons/rejected), blue (neutral), yellow (warnings/risks)
- Clear section headers with icons
- Expandable/collapsible sections for detail
- Print-friendly layout
- Responsive design
- Executive summary on first screen with the nickname and acronym prominently displayed
- Easy navigation between sections
- Runner-up ideas should each get their own card/section with equal visual treatment

Generate complete, self-contained HTML with embedded CSS and minimal JavaScript for interactivity.
This is a COMPREHENSIVE REPORT, not a simple one-pager. Think: 5-8 printed pages of detailed content.`

	return &UICreatorAgent{
		BaseAgent: &BaseAgent{
			Role:         models.RoleUICreator,
			Name:         "UI Creator",
			SystemPrompt: systemPrompt,
			Client:       client,
			Temperature:  0.6,
		},
	}
}

// Process creates a visual representation
func (a *UICreatorAgent) Process(context *models.Discussion, input string) (*models.AgentResponse, error) {
	discussionContext := BuildContext(context)

	// Build detailed context with all messages and rounds
	detailedContext := a.buildDetailedContext(context)

	query := fmt.Sprintf(`%s

%s

Task: %s

Generate a COMPREHENSIVE, MULTI-SECTION HTML report following the structure in your system prompt.

CRITICAL REQUIREMENTS:
1. Give the winning proposal a catchy MARKETING NICKNAME and a punchy 3-4 LETTER ACRONYM
2. Include detailed analysis of 3-5 runner-up ideas (not just the winner)
3. For EACH runner-up, explain specifically WHY it was eliminated, its pros/cons, and team concerns
4. Include a comparative analysis table of all top ideas
5. Show the discussion journey across all rounds with key turning points
6. List open questions, risks, and assumptions that need validation
7. Provide actionable next steps with success metrics
8. Make it 5-8 printed pages of content — this is a strategic decision document, not a summary slide

This is an executive-level strategic report for decision-makers.
Return complete, self-contained HTML with embedded CSS.`,
		discussionContext, detailedContext, input)

	// Use generous token limit for comprehensive report generation
	response, err := a.QueryWithTokens(query, 16384)
	if err != nil {
		return nil, fmt.Errorf("report generator query failed: %w", err)
	}

	return &models.AgentResponse{
		AgentRole: a.Role,
		Content:   response,
		Metadata: map[string]interface{}{
			"type": "html",
		},
	}, nil
}

// buildDetailedContext creates a rich context with discussion flow
func (a *UICreatorAgent) buildDetailedContext(discussion *models.Discussion) string {
	if discussion == nil {
		return ""
	}

	context := "\nDETAILED DISCUSSION CONTEXT:\n\n"

	// Round information
	if discussion.MaxRounds > 1 {
		context += fmt.Sprintf("Discussion Rounds: %d rounds completed\n\n", discussion.Round)
	}

	// Messages by round/phase
	context += "Discussion Flow:\n"
	for i, msg := range discussion.Messages {
		context += fmt.Sprintf("%d. [%s -> %s] (%s): %s\n",
			i+1, msg.From, msg.To, msg.Type, truncate(msg.Content, 200))
	}
	context += "\n"

	// All ideas with full details
	context += fmt.Sprintf("Total Ideas Generated: %d\n\n", len(discussion.Ideas))
	context += "Detailed Ideas:\n"
	for i, idea := range discussion.Ideas {
		context += fmt.Sprintf("\nIdea %d: %s\n", i+1, idea.Title)
		context += fmt.Sprintf("  Description: %s\n", idea.Description)
		context += fmt.Sprintf("  Category: %s\n", idea.Category)
		context += fmt.Sprintf("  Created by: %s\n", idea.CreatedBy)

		if idea.Validated {
			context += fmt.Sprintf("  Score: %.1f/10\n", idea.Score)
			if len(idea.Pros) > 0 {
				context += "  Pros:\n"
				for _, pro := range idea.Pros {
					context += fmt.Sprintf("    + %s\n", pro)
				}
			}
			if len(idea.Cons) > 0 {
				context += "  Cons:\n"
				for _, con := range idea.Cons {
					context += fmt.Sprintf("    - %s\n", con)
				}
			}
		}
	}

	// Final selection
	if discussion.FinalIdea != nil {
		context += fmt.Sprintf("\nFinal Selected Idea: %s (Score: %.1f/10)\n",
			discussion.FinalIdea.Title, discussion.FinalIdea.Score)
	}

	return context
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GenerateIdeaSheet creates a complete HTML report from a discussion
func (a *UICreatorAgent) GenerateIdeaSheet(discussion *models.Discussion) (string, error) {
	if discussion == nil {
		return "", fmt.Errorf("discussion is nil")
	}

	// Identify top ideas for the report (winner + 3-5 runners-up)
	topIdeas := getTopIdeas(discussion, 5)

	input := fmt.Sprintf(`Generate a comprehensive strategic report.

Focus on:
- Executive summary with final recommendation
- A catchy MARKETING NICKNAME and 3-4 LETTER ACRONYM for the winning proposal
- Top %d ideas with detailed analysis including pros, cons, and team concerns
- Specific reasons why each runner-up was eliminated
- What circumstances might favor each alternative
- A side-by-side comparative analysis table
- Open questions, risks, and assumptions to validate
- Actionable next steps with success metrics

Remember: This is a detailed strategic decision document for leadership, not a brief summary.`, len(topIdeas))

	response, err := a.Process(discussion, input)
	if err != nil {
		return "", err
	}

	return stripCodeFences(response.Content), nil
}

// stripCodeFences removes markdown code fences (```html ... ```) that LLMs
// sometimes wrap around HTML output.
func stripCodeFences(html string) string {
	s := strings.TrimSpace(html)
	// Strip leading ```html or ```
	if strings.HasPrefix(s, "```html") {
		s = strings.TrimPrefix(s, "```html")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	// Strip trailing ```
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}

// getTopIdeas returns the top N ideas sorted by score
func getTopIdeas(discussion *models.Discussion, n int) []models.Idea {
	if discussion == nil || len(discussion.Ideas) == 0 {
		return []models.Idea{}
	}

	// Create a copy and sort by score
	ideas := make([]models.Idea, len(discussion.Ideas))
	copy(ideas, discussion.Ideas)

	// Simple bubble sort (fine for small lists)
	for i := 0; i < len(ideas)-1; i++ {
		for j := 0; j < len(ideas)-i-1; j++ {
			if ideas[j].Score < ideas[j+1].Score {
				ideas[j], ideas[j+1] = ideas[j+1], ideas[j]
			}
		}
	}

	// Return top N
	if len(ideas) < n {
		return ideas
	}
	return ideas[:n]
}
