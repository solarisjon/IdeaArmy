package orchestrator

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/ai-agent-team/internal/agents"
	"github.com/yourusername/ai-agent-team/internal/llm"
	"github.com/yourusername/ai-agent-team/internal/models"
)

// ConfigurableOrchestrator coordinates a configurable team of agents
type ConfigurableOrchestrator struct {
	Config     *models.TeamConfig
	Agents     map[models.AgentRole]agents.Agent
	Discussion *models.Discussion
	OnProgress func(message string)
}

// NewConfigurableOrchestrator creates a new orchestrator with custom team config
func NewConfigurableOrchestrator(client llm.Client, config *models.TeamConfig) *ConfigurableOrchestrator {
	if config == nil {
		config = models.DefaultTeamConfig()
	}

	agentMap := make(map[models.AgentRole]agents.Agent)

	// Initialize agents based on config
	if config.IncludeTeamLeader {
		agentMap[models.RoleTeamLeader] = agents.NewTeamLeaderAgent(client)
	}
	if config.IncludeIdeation {
		agentMap[models.RoleIdeation] = agents.NewIdeationAgent(client)
	}
	if config.IncludeModerator {
		agentMap[models.RoleModerator] = agents.NewModeratorAgent(client)
	}
	if config.IncludeResearcher {
		agentMap[models.RoleResearcher] = agents.NewResearcherAgent(client)
	}
	if config.IncludeCritic {
		agentMap[models.RoleCritic] = agents.NewCriticAgent(client)
	}
	if config.IncludeImplementer {
		agentMap[models.RoleImplementer] = agents.NewImplementerAgent(client)
	}
	if config.IncludeUICreator {
		agentMap[models.RoleUICreator] = agents.NewUICreatorAgent(client)
	}

	return &ConfigurableOrchestrator{
		Config: config,
		Agents: agentMap,
	}
}

// StartDiscussion initiates a multi-round discussion
func (o *ConfigurableOrchestrator) StartDiscussion(topic string) error {
	o.Discussion = &models.Discussion{
		ID:        uuid.New().String(),
		Topic:     topic,
		StartTime: time.Now(),
		Messages:  []models.Message{},
		Ideas:     []models.Idea{},
		Status:    "running",
		Round:     0,
		MaxRounds: o.Config.MaxRounds,
	}

	teamSize := o.Config.TeamSize()
	o.notify(fmt.Sprintf("🎯 Starting discussion with %d agents on: %s", teamSize, topic))
	o.notify(fmt.Sprintf("📊 Configuration: %d rounds, deep dive: %v", o.Config.MaxRounds, o.Config.DeepDive))

	// Phase 1: Kickoff
	if err := o.runKickoff(); err != nil {
		return fmt.Errorf("kickoff failed: %w", err)
	}

	// Phase 2: Multi-round exploration
	for round := 1; round <= o.Config.MaxRounds; round++ {
		o.Discussion.Round = round
		o.notify(fmt.Sprintf("\n🔄 Round %d of %d", round, o.Config.MaxRounds))

		if err := o.runExplorationRound(round); err != nil {
			return fmt.Errorf("round %d failed: %w", round, err)
		}

		// Leader synthesis after each round
		if err := o.runLeaderSynthesis(round); err != nil {
			return fmt.Errorf("synthesis in round %d failed: %w", round, err)
		}
	}

	// Phase 3: Final validation and selection
	if err := o.runFinalValidation(); err != nil {
		return fmt.Errorf("final validation failed: %w", err)
	}

	// Phase 4: Visualization
	if err := o.runVisualization(); err != nil {
		return fmt.Errorf("visualization failed: %w", err)
	}

	o.Discussion.EndTime = time.Now()
	o.Discussion.Status = "completed"
	o.notify("\n✅ Discussion completed successfully!")

	return nil
}

// runKickoff - Team leader introduces the topic
func (o *ConfigurableOrchestrator) runKickoff() error {
	o.notify("📋 Phase 1: Team Leader Kickoff")

	leader, ok := o.Agents[models.RoleTeamLeader]
	if !ok {
		return fmt.Errorf("team leader is required")
	}

	teamMembers := o.getTeamMembersList()
	input := fmt.Sprintf(`We have a team of %d agents to explore: %s

Team members: %s

Please set the direction for this discussion. What should each team member focus on?`,
		o.Config.TeamSize(), o.Discussion.Topic, teamMembers)

	response, err := leader.Process(o.Discussion, input)
	if err != nil {
		return err
	}

	o.addMessage("system", string(models.RoleTeamLeader), response.Content, "kickoff")
	o.notify(fmt.Sprintf("  📣 [team_leader] %s", o.truncate(response.Content, 200)))

	return nil
}

// runExplorationRound - Agents contribute in sequence, building on each other
func (o *ConfigurableOrchestrator) runExplorationRound(round int) error {
	o.notify(fmt.Sprintf("💡 Exploration Round %d", round))

	// Research phase (if researcher is available)
	if _, hasResearcher := o.Agents[models.RoleResearcher]; hasResearcher {
		if err := o.runAgentContribution(models.RoleResearcher, "Provide research and context for this topic"); err != nil {
			return err
		}
	}

	// Ideation phase — run 1..IdeationCount passes
	if _, hasIdeation := o.Agents[models.RoleIdeation]; hasIdeation {
		count := o.Config.IdeationCount
		if count < 1 {
			count = 1
		}
		for pass := 0; pass < count; pass++ {
			prompt := "Generate creative ideas based on the discussion so far"
			if round > 1 || pass > 0 {
				prompt = "Building on previous ideas and feedback, generate refined or new creative ideas"
			}
			if err := o.runAgentContribution(models.RoleIdeation, prompt); err != nil {
				return err
			}
		}
	}

	// Critical analysis (if critic is available)
	if _, hasCritic := o.Agents[models.RoleCritic]; hasCritic && len(o.Discussion.Ideas) > 0 {
		if err := o.runAgentContribution(models.RoleCritic, "Challenge the assumptions in these ideas. What could go wrong?"); err != nil {
			return err
		}
	}

	// Implementation thinking (if implementer is available)
	if _, hasImplementer := o.Agents[models.RoleImplementer]; hasImplementer && len(o.Discussion.Ideas) > 0 {
		if err := o.runAgentContribution(models.RoleImplementer, "How would we actually implement these ideas? What's the practical approach?"); err != nil {
			return err
		}
	}

	return nil
}

// runAgentContribution - Single agent contributes, results go back to team leader
func (o *ConfigurableOrchestrator) runAgentContribution(role models.AgentRole, prompt string) error {
	agent, ok := o.Agents[role]
	if !ok {
		return nil // Agent not in team
	}

	o.notify(fmt.Sprintf("  🗣️  %s contributing...", agent.GetName()))

	response, err := agent.Process(o.Discussion, prompt)
	if err != nil {
		log.Printf("Warning: %s contribution failed: %v", agent.GetName(), err)
		return nil // Don't fail the whole discussion
	}

	// Add ideas if any were generated
	if len(response.Ideas) > 0 {
		var ideaTitles []string
		for _, idea := range response.Ideas {
			o.Discussion.Ideas = append(o.Discussion.Ideas, idea)
			o.notify(fmt.Sprintf("    💡 New idea: %s", idea.Title))
			ideaTitles = append(ideaTitles, idea.Title)
		}
		// Use idea titles for speech bubble instead of raw JSON
		speechText := fmt.Sprintf("💡 Proposed: %s", strings.Join(ideaTitles, " | "))
		o.notify(fmt.Sprintf("  📣 [%s] %s", string(role), o.truncate(speechText, 200)))
	} else {
		o.notify(fmt.Sprintf("  📣 [%s] %s", string(role), o.truncate(response.Content, 200)))
	}

	o.addMessage(string(role), "team", response.Content, string(role))

	return nil
}

// runLeaderSynthesis - Leader synthesizes the round and directs next steps
func (o *ConfigurableOrchestrator) runLeaderSynthesis(round int) error {
	leader, ok := o.Agents[models.RoleTeamLeader]
	if !ok {
		return nil
	}

	o.notify(fmt.Sprintf("  🎯 Team Leader synthesizing round %d...", round))

	prompt := fmt.Sprintf(`Synthesize the contributions from round %d.

What are the key insights? What should the team focus on in the next round?
If this is the final round, identify which ideas are strongest.`, round)

	response, err := leader.Process(o.Discussion, prompt)
	if err != nil {
		return err
	}

	o.addMessage("system", string(models.RoleTeamLeader), response.Content, "synthesis")
	o.notify(fmt.Sprintf("  📣 [team_leader] %s", o.truncate(response.Content, 200)))

	return nil
}

// runFinalValidation - Moderator does final evaluation
func (o *ConfigurableOrchestrator) runFinalValidation() error {
	o.notify("\n🔍 Phase: Final Validation")

	moderator, ok := o.Agents[models.RoleModerator]
	if !ok {
		// If no moderator, skip validation
		return o.runLeaderSelection()
	}

	if len(o.Discussion.Ideas) == 0 {
		return fmt.Errorf("no ideas to validate")
	}

	response, err := moderator.Process(o.Discussion,
		"Provide final scores and comprehensive evaluation of all ideas discussed")
	if err != nil {
		return err
	}

	o.addMessage("system", string(models.RoleModerator), response.Content, "validation")
	o.notify(fmt.Sprintf("  📣 [moderator] Evaluating and scoring all ideas..."))

	// Show scores
	for _, idea := range o.Discussion.Ideas {
		if idea.Validated {
			o.notify(fmt.Sprintf("  📊 %s - Score: %.1f/10", idea.Title, idea.Score))
		}
	}

	return o.runLeaderSelection()
}

// runLeaderSelection - Leader selects the best idea
func (o *ConfigurableOrchestrator) runLeaderSelection() error {
	leader, ok := o.Agents[models.RoleTeamLeader]
	if !ok {
		// Auto-select highest scored idea
		return o.autoSelectBestIdea()
	}

	o.notify("\n🎯 Phase: Final Selection")

	response, err := leader.Process(o.Discussion,
		"Based on all the discussion, evaluation, and team input, select the best idea and explain your decision")
	if err != nil {
		return err
	}

	o.addMessage("system", string(models.RoleTeamLeader), response.Content, "selection")
	o.notify(fmt.Sprintf("  📣 [team_leader] %s", o.truncate(response.Content, 200)))

	// Select highest scored idea
	return o.autoSelectBestIdea()
}

// autoSelectBestIdea selects the highest-scored idea
func (o *ConfigurableOrchestrator) autoSelectBestIdea() error {
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

// runVisualization - Create the idea sheet
func (o *ConfigurableOrchestrator) runVisualization() error {
	uiCreator, ok := o.Agents[models.RoleUICreator]
	if !ok {
		return nil // Optional
	}

	o.notify("\n🎨 Phase: Creating Visual Idea Sheet")

	html, err := uiCreator.(*agents.UICreatorAgent).GenerateIdeaSheet(o.Discussion)
	if err != nil {
		o.notify(fmt.Sprintf("  ⚠️ Report generation failed: %s", err.Error()))
		o.notify("  📣 [ui_creator] Sorry, couldn't generate the report this time!")
		// Non-fatal — don't fail the whole discussion over a visualization error
		return nil
	}

	o.addMessage(string(models.RoleTeamLeader), string(models.RoleUICreator), "Create the final idea sheet", "request")
	o.addMessage(string(models.RoleUICreator), "team", html, "visualization")

	o.notify("  ✨ Idea sheet generated successfully")
	o.notify("  📣 [ui_creator] Idea sheet created — painting the final vision!")

	return nil
}

// Helper methods

func (o *ConfigurableOrchestrator) getTeamMembersList() string {
	var members []string
	for role, agent := range o.Agents {
		if role != models.RoleTeamLeader && role != models.RoleUICreator {
			members = append(members, agent.GetName())
		}
	}
	result := ""
	for i, member := range members {
		if i > 0 {
			result += ", "
		}
		result += member
	}
	return result
}

func (o *ConfigurableOrchestrator) addMessage(from, to, content, msgType string) {
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

func (o *ConfigurableOrchestrator) notify(message string) {
	if o.OnProgress != nil {
		o.OnProgress(message)
	} else {
		log.Println(message)
	}
}

func (o *ConfigurableOrchestrator) truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (o *ConfigurableOrchestrator) GetDiscussion() *models.Discussion {
	return o.Discussion
}

func (o *ConfigurableOrchestrator) GetIdeaSheetHTML() string {
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
