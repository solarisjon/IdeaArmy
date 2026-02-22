package agents

import (
	"fmt"
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
- Present ALL ideas explored, with focus on top 3-4 candidates
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
   - Detailed description
   - Complete pros and cons
   - Implementation considerations
   - Why this was selected over others
   - Risk assessment

3. RUNNER-UP IDEAS (Top 3-4 alternatives)
   For EACH runner-up:
   - Full description and score
   - Detailed pros and cons
   - Why it wasn't selected (specific reasons)
   - Under what circumstances it might be better
   - Could it be combined with the winner?

4. ALL IDEAS EXPLORED
   - Complete list with scores
   - Brief description of each
   - Quick assessment

5. DISCUSSION JOURNEY
   - How the discussion evolved
   - Key insights from each round
   - How ideas were refined
   - Team dynamics and perspectives
   - What we learned

6. COMPARATIVE ANALYSIS
   - Side-by-side comparison of top ideas
   - Decision criteria and weightings
   - Trade-offs considered

7. OPEN QUESTIONS & NEXT STEPS
   - Unanswered questions that need research
   - Assumptions that need validation
   - Recommended next actions
   - Suggested follow-up discussions
   - Areas requiring expert input

8. RECOMMENDATIONS & CONSIDERATIONS
   - Implementation suggestions
   - Timeline considerations
   - Resource requirements
   - Risk mitigation strategies
   - Success metrics

Design principles:
- Use modern, professional styling (cards, gradients, shadows)
- Color coding: green (pros/selected), red (cons/rejected), blue (neutral), yellow (warnings)
- Clear section headers with icons
- Expandable/collapsible sections for detail
- Print-friendly layout
- Responsive design
- Executive summary on first screen
- Easy navigation between sections

Generate complete, self-contained HTML with embedded CSS and minimal JavaScript for interactivity.
This is a COMPREHENSIVE REPORT, not a simple one-pager. Think: 3-5 screens of detailed content.`

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
1. Include detailed analysis of the TOP 3-4 runner-up ideas (not just the winner)
2. For EACH runner-up, explain specifically WHY it wasn't selected
3. Include open questions and recommendations
4. Show the discussion journey across all rounds
5. Provide actionable next steps
6. Make it 3-5 screens of content, not a single page

This is an executive-level strategic report, not a summary slide.
Return complete, self-contained HTML with embedded CSS.`,
		discussionContext, detailedContext, input)

	// Use more tokens for comprehensive report generation
	response, err := a.QueryWithTokens(query, 4096)
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

	// Sort ideas by score to identify runners-up
	topIdeas := getTopIdeas(discussion, 4)

	input := fmt.Sprintf(`Generate a comprehensive strategic report.

Focus on:
- Executive summary with final recommendation
- Top %d ideas with detailed analysis
- Specific reasons why runner-ups weren't selected
- What circumstances might favor each alternative
- Open questions for further exploration
- Actionable next steps and recommendations

Remember: This is a detailed report for decision-makers, not a brief summary.`, len(topIdeas))

	response, err := a.Process(discussion, input)
	if err != nil {
		return "", err
	}

	return response.Content, nil
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
