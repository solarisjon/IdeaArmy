package agents

import (
	"encoding/json"
	"fmt"
	"github.com/yourusername/ai-agent-team/internal/claude"
	"github.com/yourusername/ai-agent-team/internal/models"
	"strings"
)

// ModeratorAgent validates and evaluates ideas
type ModeratorAgent struct {
	*BaseAgent
}

// NewModeratorAgent creates a new moderator agent
func NewModeratorAgent(client *claude.Client) *ModeratorAgent {
	systemPrompt := `You are the Moderator/Facilitator Agent, responsible for ensuring idea quality and validity.

Your responsibilities:
- Critically evaluate all proposed ideas
- Check ideas for feasibility, coherence, and value
- Identify potential issues, risks, or limitations
- Provide constructive feedback and suggestions for improvement
- Assign quality scores to ideas (0-10 scale)
- Ensure ideas are well-researched and thought-through
- Challenge assumptions and ask probing questions

Evaluation criteria:
- Feasibility: Can this be realistically implemented?
- Innovation: Is this creative and differentiated?
- Impact: What value does this provide?
- Clarity: Is the idea well-defined and understandable?
- Completeness: Is the idea fully thought through?

When evaluating ideas, structure your response as JSON:
{
  "evaluations": [
    {
      "idea_id": "id of the idea",
      "score": 8.5,
      "pros": ["strength 1", "strength 2"],
      "cons": ["weakness 1", "weakness 2"],
      "feedback": "Detailed feedback and suggestions"
    }
  ],
  "overall_assessment": "Summary of the evaluation"
}

Be thorough, fair, and constructive. Your goal is to ensure only high-quality ideas move forward.`

	return &ModeratorAgent{
		BaseAgent: &BaseAgent{
			Role:         models.RoleModerator,
			Name:         "Moderator/Facilitator",
			SystemPrompt: systemPrompt,
			Client:       client,
			Temperature:  0.5, // Lower temperature for analytical thinking
		},
	}
}

// Process evaluates ideas
func (a *ModeratorAgent) Process(context *models.Discussion, input string) (*models.AgentResponse, error) {
	discussionContext := BuildContext(context)

	query := fmt.Sprintf(`%s

Task: %s

Evaluate the ideas presented. Provide scores, identify pros and cons, and give detailed feedback. Return your response as JSON following the specified format.`,
		discussionContext, input)

	response, err := a.Query(query)
	if err != nil {
		return nil, fmt.Errorf("moderator query failed: %w", err)
	}

	// Update ideas with evaluation data
	a.updateIdeasWithEvaluation(response, context)

	return &models.AgentResponse{
		AgentRole: a.Role,
		Content:   response,
	}, nil
}

// updateIdeasWithEvaluation parses evaluation and updates ideas
func (a *ModeratorAgent) updateIdeasWithEvaluation(response string, discussion *models.Discussion) {
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")

	if startIdx == -1 || endIdx == -1 || discussion == nil {
		return
	}

	jsonStr := response[startIdx : endIdx+1]

	var parsed struct {
		Evaluations []struct {
			IdeaID   string   `json:"idea_id"`
			Score    float64  `json:"score"`
			Pros     []string `json:"pros"`
			Cons     []string `json:"cons"`
			Feedback string   `json:"feedback"`
		} `json:"evaluations"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return
	}

	// Update ideas in the discussion
	for i := range discussion.Ideas {
		for _, eval := range parsed.Evaluations {
			if discussion.Ideas[i].ID == eval.IdeaID ||
			   discussion.Ideas[i].Title == eval.IdeaID { // Allow matching by title too
				discussion.Ideas[i].Score = eval.Score
				discussion.Ideas[i].Pros = eval.Pros
				discussion.Ideas[i].Cons = eval.Cons
				discussion.Ideas[i].Validated = true
			}
		}
	}
}
