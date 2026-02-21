package orchestrator

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/ai-agent-team/internal/agents"
	"github.com/yourusername/ai-agent-team/internal/claude"
	"github.com/yourusername/ai-agent-team/internal/models"
)

// Orchestrator coordinates the agent team
type Orchestrator struct {
	TeamLeader *agents.TeamLeaderAgent
	Ideation   *agents.IdeationAgent
	Moderator  *agents.ModeratorAgent
	UICreator  *agents.UICreatorAgent
	Discussion *models.Discussion
	OnProgress func(message string) // Callback for progress updates
}

// NewOrchestrator creates a new orchestrator with all agents
func NewOrchestrator(apiKey string) *Orchestrator {
	client := claude.NewClient(apiKey)

	return &Orchestrator{
		TeamLeader: agents.NewTeamLeaderAgent(client),
		Ideation:   agents.NewIdeationAgent(client),
		Moderator:  agents.NewModeratorAgent(client),
		UICreator:  agents.NewUICreatorAgent(client),
	}
}

// StartDiscussion initiates a new discussion on a topic
func (o *Orchestrator) StartDiscussion(topic string) error {
	o.Discussion = &models.Discussion{
		ID:        uuid.New().String(),
		Topic:     topic,
		StartTime: time.Now(),
		Messages:  []models.Message{},
		Ideas:     []models.Idea{},
		Status:    "running",
	}

	o.notify(fmt.Sprintf("🎯 Starting discussion on: %s", topic))

	// Phase 1: Team Leader kicks off the discussion
	if err := o.runPhase1_Kickoff(); err != nil {
		return fmt.Errorf("kickoff phase failed: %w", err)
	}

	// Phase 2: Ideation - Generate ideas
	if err := o.runPhase2_Ideation(); err != nil {
		return fmt.Errorf("ideation phase failed: %w", err)
	}

	// Phase 3: Validation - Moderator evaluates ideas
	if err := o.runPhase3_Validation(); err != nil {
		return fmt.Errorf("validation phase failed: %w", err)
	}

	// Phase 4: Selection - Team Leader chooses the best idea
	if err := o.runPhase4_Selection(); err != nil {
		return fmt.Errorf("selection phase failed: %w", err)
	}

	// Phase 5: Visualization - UI Creator generates the idea sheet
	if err := o.runPhase5_Visualization(); err != nil {
		return fmt.Errorf("visualization phase failed: %w", err)
	}

	o.Discussion.EndTime = time.Now()
	o.Discussion.Status = "completed"
	o.notify("✅ Discussion completed successfully!")

	return nil
}

// runPhase1_Kickoff - Team Leader introduces the topic
func (o *Orchestrator) runPhase1_Kickoff() error {
	o.notify("📋 Phase 1: Team Leader Kickoff")

	input := fmt.Sprintf("We need to explore and develop ideas around: %s. Please outline how we should approach this and direct the team.", o.Discussion.Topic)

	response, err := o.TeamLeader.Process(o.Discussion, input)
	if err != nil {
		return err
	}

	o.addMessage("system", string(models.RoleTeamLeader), response.Content, "kickoff")
	o.notify(fmt.Sprintf("Team Leader: %s", o.truncate(response.Content, 150)))

	return nil
}

// runPhase2_Ideation - Generate creative ideas
func (o *Orchestrator) runPhase2_Ideation() error {
	o.notify("💡 Phase 2: Ideation")

	input := "Generate creative, well-researched ideas for this topic. Think deeply about feasibility, innovation, and impact."

	response, err := o.Ideation.Process(o.Discussion, input)
	if err != nil {
		return err
	}

	// Add ideas to discussion
	for _, idea := range response.Ideas {
		o.Discussion.Ideas = append(o.Discussion.Ideas, idea)
		o.notify(fmt.Sprintf("  💡 Idea: %s", idea.Title))
	}

	o.addMessage(string(models.RoleTeamLeader), string(models.RoleIdeation), input, "request")
	o.addMessage(string(models.RoleIdeation), "team", response.Content, "idea")

	return nil
}

// runPhase3_Validation - Evaluate and validate ideas
func (o *Orchestrator) runPhase3_Validation() error {
	o.notify("🔍 Phase 3: Validation & Evaluation")

	if len(o.Discussion.Ideas) == 0 {
		return fmt.Errorf("no ideas to validate")
	}

	input := "Review and evaluate all the proposed ideas. Provide scores, identify pros and cons, and give constructive feedback."

	response, err := o.Moderator.Process(o.Discussion, input)
	if err != nil {
		return err
	}

	o.addMessage(string(models.RoleTeamLeader), string(models.RoleModerator), input, "request")
	o.addMessage(string(models.RoleModerator), "team", response.Content, "validation")

	// Show scores
	for _, idea := range o.Discussion.Ideas {
		if idea.Validated {
			o.notify(fmt.Sprintf("  📊 %s - Score: %.1f/10", idea.Title, idea.Score))
		}
	}

	return nil
}

// runPhase4_Selection - Choose the best idea
func (o *Orchestrator) runPhase4_Selection() error {
	o.notify("🎯 Phase 4: Selection")

	input := "Based on the evaluations, select the best idea to move forward with. Explain your reasoning."

	response, err := o.TeamLeader.Process(o.Discussion, input)
	if err != nil {
		return err
	}

	o.addMessage("system", string(models.RoleTeamLeader), response.Content, "selection")

	// Find the highest-scored idea as the final idea
	var bestIdea *models.Idea
	bestScore := 0.0

	for i := range o.Discussion.Ideas {
		if o.Discussion.Ideas[i].Score > bestScore {
			bestScore = o.Discussion.Ideas[i].Score
			bestIdea = &o.Discussion.Ideas[i]
		}
	}

	if bestIdea != nil {
		o.Discussion.FinalIdea = bestIdea
		o.notify(fmt.Sprintf("  ⭐ Final Idea: %s (Score: %.1f/10)", bestIdea.Title, bestIdea.Score))
	}

	return nil
}

// runPhase5_Visualization - Create the idea sheet
func (o *Orchestrator) runPhase5_Visualization() error {
	o.notify("🎨 Phase 5: Creating Visual Idea Sheet")

	html, err := o.UICreator.GenerateIdeaSheet(o.Discussion)
	if err != nil {
		return err
	}

	o.addMessage(string(models.RoleTeamLeader), string(models.RoleUICreator), "Create the final idea sheet", "request")
	o.addMessage(string(models.RoleUICreator), "team", html, "visualization")

	o.notify("  ✨ Idea sheet generated successfully")

	return nil
}

// GetIdeaSheetHTML returns the HTML idea sheet from the discussion
func (o *Orchestrator) GetIdeaSheetHTML() string {
	if o.Discussion == nil {
		return ""
	}

	for _, msg := range o.Discussion.Messages {
		if msg.Type == "visualization" {
			return msg.Content
		}
	}

	return ""
}

// addMessage adds a message to the discussion
func (o *Orchestrator) addMessage(from, to, content, msgType string) {
	msg := models.Message{
		ID:        uuid.New().String(),
		From:      from,
		To:        to,
		Content:   content,
		Timestamp: time.Now(),
		Type:      msgType,
	}
	o.Discussion.Messages = append(o.Discussion.Messages, msg)
}

// notify sends a progress update
func (o *Orchestrator) notify(message string) {
	if o.OnProgress != nil {
		o.OnProgress(message)
	} else {
		log.Println(message)
	}
}

// truncate truncates a string to maxLen characters
func (o *Orchestrator) truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GetDiscussion returns the current discussion
func (o *Orchestrator) GetDiscussion() *models.Discussion {
	return o.Discussion
}
