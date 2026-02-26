package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/ai-agent-team/internal/llm"
	"github.com/yourusername/ai-agent-team/internal/models"
	"github.com/yourusername/ai-agent-team/internal/orchestrator"
)

// discussionResult holds the result from the goroutine
var discussionResult struct {
	discussion *models.Discussion
	err        error
}

// Run starts the TUI and runs the discussion
func Run(client llm.Client, config *models.TeamConfig, topic string) (*models.Discussion, error) {
	// Create the TUI model
	m := NewModel(config, topic)

	// Create the bubbletea program
	p := tea.NewProgram(m)

	// Start the discussion in a goroutine
	go runDiscussion(p, client, config, topic)

	// Run the TUI
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("TUI error: %w", err)
	}

	// Extract the final model
	model := finalModel.(Model)

	if model.Status == "error" {
		return nil, fmt.Errorf("discussion failed: %s", model.ErrorMessage)
	}

	// Return the discussion result
	return discussionResult.discussion, discussionResult.err
}

// runDiscussion runs the orchestration and sends updates to the TUI
func runDiscussion(p *tea.Program, client llm.Client, config *models.TeamConfig, topic string) {
	// Create orchestrator
	orch := orchestrator.NewConfigurableOrchestrator(client, config)

	// Set up progress callback to send updates to TUI
	orch.OnProgress = func(message string) {
		// Parse the message to determine what kind of update it is
		p.Send(LogMsg(message))

		// Detect phase changes
		if strings.Contains(message, "Phase") {
			phase := extractPhase(message)
			round := extractRound(message)
			p.Send(ProgressMsg{
				Phase:    phase,
				Round:    round,
				Progress: 0.1,
			})
		}

		// Detect agent speech (📣 [role] content)
		if strings.Contains(message, "📣 [") {
			role, speech := extractSpeech(message)
			if role != "" {
				p.Send(AgentUpdateMsg{
					Role:    role,
					Status:  "working",
					Message: "Just spoke",
					Speech:  speech,
				})
			}
		}

		// Detect agent activity
		if strings.Contains(message, "contributing") || strings.Contains(message, "working") {
			role := extractAgentRole(message)
			if role != "" {
				p.Send(AgentUpdateMsg{
					Role:    role,
					Status:  "working",
					Message: extractAgentMessage(message),
				})
			}
		}

		// Detect leader/moderator/ui_creator phase starts
		if strings.Contains(message, "Team Leader synthesizing") {
			p.Send(AgentUpdateMsg{Role: "team_leader", Status: "working", Message: "Synthesizing round..."})
		}
		if strings.Contains(message, "Final Validation") {
			p.Send(AgentUpdateMsg{Role: "moderator", Status: "working", Message: "Scoring ideas..."})
		}
		if strings.Contains(message, "Final Selection") {
			p.Send(AgentUpdateMsg{Role: "team_leader", Status: "working", Message: "Selecting best idea..."})
		}
		if strings.Contains(message, "Creating Visual Idea Sheet") {
			p.Send(AgentUpdateMsg{Role: "ui_creator", Status: "working", Message: "Painting the vision..."})
		}
		if strings.Contains(message, "Team Leader Kickoff") {
			p.Send(AgentUpdateMsg{Role: "team_leader", Status: "working", Message: "Setting the direction..."})
		}

		// Detect idea generation
		if strings.Contains(message, "New idea:") || strings.Contains(message, "💡") {
			// Ideas will be extracted from the discussion object later
		}

		// Update progress based on message
		if strings.Contains(message, "Round") {
			round := extractRound(message)
			progress := float64(round) / float64(config.MaxRounds)
			p.Send(ProgressMsg{
				Phase:    extractPhase(message),
				Round:    round,
				Progress: progress,
			})
		}
	}

	// Run the discussion
	err := orch.StartDiscussion(topic)

	if err != nil {
		p.Send(ErrorMsg{Err: err})
		discussionResult.err = err
		return
	}

	// Get the discussion results
	discussion := orch.GetDiscussion()
	discussionResult.discussion = discussion

	// Send ideas to TUI
	if discussion != nil {
		for _, idea := range discussion.Ideas {
			ideaCopy := idea
			p.Send(IdeaGeneratedMsg{Idea: &ideaCopy})
		}
	}

	// Mark completion
	p.Send(CompleteMsg{Discussion: discussion})

	// Mark all agents as complete
	for _, role := range config.GetActiveAgentRoles() {
		p.Send(AgentUpdateMsg{
			Role:   string(role),
			Status: "complete",
		})
	}
}

// Helper functions to extract information from progress messages

func extractPhase(message string) string {
	if strings.Contains(message, "Kickoff") {
		return "Team Leader Kickoff"
	}
	if strings.Contains(message, "Exploration") {
		return "Exploration & Ideation"
	}
	if strings.Contains(message, "Validation") {
		return "Validation & Scoring"
	}
	if strings.Contains(message, "Selection") {
		return "Final Selection"
	}
	if strings.Contains(message, "Visualization") {
		return "Creating Idea Sheet"
	}
	if strings.Contains(message, "synthesizing") {
		return "Leader Synthesis"
	}
	return "Processing"
}

func extractRound(message string) int {
	// Try to extract round number from messages like "Round 1 of 3"
	if strings.Contains(message, "Round") {
		parts := strings.Split(message, "Round")
		if len(parts) > 1 {
			// Try to parse the number
			var round int
			fmt.Sscanf(parts[1], "%d", &round)
			if round > 0 {
				return round
			}
		}
	}
	return 1
}

func extractAgentRole(message string) string {
	lowerMsg := strings.ToLower(message)

	if strings.Contains(lowerMsg, "team leader") {
		return "team_leader"
	}
	if strings.Contains(lowerMsg, "ideation") {
		return "ideation"
	}
	if strings.Contains(lowerMsg, "moderator") {
		return "moderator"
	}
	if strings.Contains(lowerMsg, "researcher") {
		return "researcher"
	}
	if strings.Contains(lowerMsg, "critic") {
		return "critic"
	}
	if strings.Contains(lowerMsg, "implementer") {
		return "implementer"
	}
	if strings.Contains(lowerMsg, "ui creator") {
		return "ui_creator"
	}

	return ""
}

func extractAgentMessage(message string) string {
	// Extract the meaningful part of the message
	if strings.Contains(message, "contributing") {
		return "Contributing ideas..."
	}
	if strings.Contains(message, "synthesizing") {
		return "Synthesizing discussion..."
	}
	if strings.Contains(message, "evaluating") {
		return "Evaluating ideas..."
	}
	return "Working..."
}

// extractSpeech parses "📣 [role] speech content" messages
func extractSpeech(message string) (string, string) {
	idx := strings.Index(message, "📣 [")
	if idx < 0 {
		return "", ""
	}
	rest := message[idx+len("📣 ["):]
	endBracket := strings.Index(rest, "] ")
	if endBracket < 0 {
		return "", ""
	}
	role := rest[:endBracket]
	speech := strings.TrimSpace(rest[endBracket+2:])
	return role, cleanSpeechContent(speech)
}

// cleanSpeechContent distills raw LLM output into a short conversational soundbite.
func cleanSpeechContent(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	// JSON content (ideation agent mostly) — extract idea titles
	if strings.HasPrefix(text, "{") || strings.HasPrefix(text, "[") {
		var titles []string
		for _, line := range strings.Split(text, "\n") {
			line = strings.TrimSpace(line)
			if strings.Contains(line, `"title"`) {
				val := extractJSONValue(line)
				if val != "" {
					titles = append(titles, val)
				}
			}
		}
		if len(titles) == 1 {
			return "💡 " + titles[0]
		}
		if len(titles) > 1 {
			more := ""
			if len(titles) > 2 {
				more = fmt.Sprintf(" (+%d more)", len(titles)-2)
			}
			return "💡 " + titles[0] + " · " + titles[1] + more
		}
		return "Thinking through ideas..."
	}

	// Strip common LLM preambles that aren't actual dialog
	for _, prefix := range []string{
		"Sure, here", "Certainly,", "Of course,", "Here is", "Here are",
		"As a ", "In my ", "I will ", "I'll ",
	} {
		if strings.HasPrefix(text, prefix) {
			if nl := strings.IndexAny(text, ".\n"); nl > 0 && nl < 80 {
				text = strings.TrimSpace(text[nl+1:])
			}
		}
	}

	// Strip markdown prefix chars and inline formatting
	text = strings.TrimLeft(text, "*#-> \t")
	text = stripInlineMarkdown(text)

	// Extract first complete sentence
	for i, ch := range text {
		if ch == '.' || ch == '!' || ch == '?' {
			sentence := strings.TrimSpace(text[:i+1])
			if len(sentence) > 8 {
				if len(sentence) > 160 {
					sentence = sentence[:157] + "..."
				}
				return sentence
			}
		}
	}

	// No sentence terminator — take first line or clip
	if nl := strings.Index(text, "\n"); nl > 0 {
		text = strings.TrimSpace(text[:nl])
	}
	if len(text) > 160 {
		return text[:157] + "..."
	}
	return text
}

// stripInlineMarkdown removes **bold**, *italic*, __underline__, `code` markers.
func stripInlineMarkdown(s string) string {
	result := strings.Builder{}
	i := 0
	for i < len(s) {
		// ** or __
		if i+1 < len(s) && ((s[i] == '*' && s[i+1] == '*') || (s[i] == '_' && s[i+1] == '_')) {
			i += 2
			continue
		}
		// single * or _
		if s[i] == '*' || s[i] == '_' {
			i++
			continue
		}
		// backtick
		if s[i] == '`' {
			i++
			continue
		}
		result.WriteByte(s[i])
		i++
	}
	return result.String()
}

// extractJSONValue pulls the value from a "key": "value" line
func extractJSONValue(line string) string {
	colonIdx := strings.Index(line, ":")
	if colonIdx < 0 {
		return ""
	}
	val := strings.TrimSpace(line[colonIdx+1:])
	val = strings.TrimSuffix(val, ",")
	val = strings.Trim(val, `"`)
	return val
}
