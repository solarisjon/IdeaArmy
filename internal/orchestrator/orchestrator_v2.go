package orchestrator

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/ai-agent-team/internal/agents"
	"github.com/yourusername/ai-agent-team/internal/llm"
	"github.com/yourusername/ai-agent-team/internal/llmfactory"
	"github.com/yourusername/ai-agent-team/internal/models"
	"github.com/yourusername/ai-agent-team/internal/report"
)

// ConfigurableOrchestrator coordinates a configurable team of agents
type ConfigurableOrchestrator struct {
	Config        *models.TeamConfig
	BackendConfig *llm.BackendConfig
	Agents        map[models.AgentRole]agents.Agent
	Discussion    *models.Discussion
	OnProgress    func(message string)
	// OnChunk is called for each streaming token: role + chunk text.
	OnChunk func(role string, chunk string)
	// OnEvidence is called when the researcher returns structured search results.
	// role is the agent role string, results is []tools.SearchResult as []interface{}.
	OnEvidence func(role string, results []interface{})
}

// NewConfigurableOrchestrator creates a new orchestrator with custom team config.
// It accepts a BackendConfig so it can create per-agent clients with different models.
func NewConfigurableOrchestrator(cfg *llm.BackendConfig, config *models.TeamConfig) *ConfigurableOrchestrator {
	if config == nil {
		config = models.DefaultTeamConfig()
	}
	if config.AgentModels == nil {
		config.AgentModels = make(map[models.AgentRole]string)
	}

	orch := &ConfigurableOrchestrator{
		Config:        config,
		BackendConfig: cfg,
		Agents:        make(map[models.AgentRole]agents.Agent),
	}

	orch.initAgents()
	return orch
}

// initAgents creates agent instances using per-agent model assignments.
func (o *ConfigurableOrchestrator) initAgents() {
	type agentEntry struct {
		role    models.AgentRole
		include bool
		create  func(llm.Client) agents.Agent
	}

	entries := []agentEntry{
		{models.RoleTeamLeader, o.Config.IncludeTeamLeader, func(c llm.Client) agents.Agent { return agents.NewTeamLeaderAgent(c) }},
		{models.RoleIdeation, o.Config.IncludeIdeation, func(c llm.Client) agents.Agent { return agents.NewIdeationAgent(c) }},
		{models.RoleModerator, o.Config.IncludeModerator, func(c llm.Client) agents.Agent { return agents.NewModeratorAgent(c) }},
		{models.RoleResearcher, o.Config.IncludeResearcher, func(c llm.Client) agents.Agent { return agents.NewResearcherAgent(c) }},
		{models.RoleCritic, o.Config.IncludeCritic, func(c llm.Client) agents.Agent { return agents.NewCriticAgent(c) }},
		{models.RoleImplementer, o.Config.IncludeImplementer, func(c llm.Client) agents.Agent { return agents.NewImplementerAgent(c) }},
		{models.RoleUICreator, o.Config.IncludeUICreator, func(c llm.Client) agents.Agent { return agents.NewUICreatorAgent(c) }},
	}

	for _, e := range entries {
		if !e.include {
			continue
		}
		model := o.Config.AgentModels[e.role]
		if model == "" {
			model = o.BackendConfig.Model
		}

		client, err := llmfactory.NewClientWithModel(o.BackendConfig, model)
		if err != nil {
			log.Printf("Warning: failed to create client for %s with model %s: %v (using default)", e.role, model, err)
			client, _ = llmfactory.NewClient(o.BackendConfig)
			model = o.BackendConfig.Model
		}

		agent := e.create(client)
		// Set the model name on the agent's BaseAgent
		if ba, ok := getBaseAgent(agent); ok {
			ba.Model = model
		}
		o.Agents[e.role] = agent
	}
}

// reinitAgent recreates a single agent with a new model.
func (o *ConfigurableOrchestrator) reinitAgent(role models.AgentRole, model string) error {
	creators := map[models.AgentRole]func(llm.Client) agents.Agent{
		models.RoleTeamLeader:  func(c llm.Client) agents.Agent { return agents.NewTeamLeaderAgent(c) },
		models.RoleIdeation:    func(c llm.Client) agents.Agent { return agents.NewIdeationAgent(c) },
		models.RoleModerator:   func(c llm.Client) agents.Agent { return agents.NewModeratorAgent(c) },
		models.RoleResearcher:  func(c llm.Client) agents.Agent { return agents.NewResearcherAgent(c) },
		models.RoleCritic:      func(c llm.Client) agents.Agent { return agents.NewCriticAgent(c) },
		models.RoleImplementer: func(c llm.Client) agents.Agent { return agents.NewImplementerAgent(c) },
		models.RoleUICreator:   func(c llm.Client) agents.Agent { return agents.NewUICreatorAgent(c) },
	}

	creator, ok := creators[role]
	if !ok {
		return fmt.Errorf("unknown role: %s", role)
	}

	client, err := llmfactory.NewClientWithModel(o.BackendConfig, model)
	if err != nil {
		return fmt.Errorf("creating client for %s model %s: %w", role, model, err)
	}

	agent := creator(client)
	if ba, ok := getBaseAgent(agent); ok {
		ba.Model = model
	}
	o.Agents[role] = agent
	o.Config.AgentModels[role] = model
	return nil
}

// getBaseAgent extracts the embedded *BaseAgent from any agent via the common
// concrete types. Returns false if the agent type is not recognized.
func getBaseAgent(a agents.Agent) (*agents.BaseAgent, bool) {
	switch v := a.(type) {
	case *agents.TeamLeaderAgent:
		return v.BaseAgent, true
	case *agents.IdeationAgent:
		return v.BaseAgent, true
	case *agents.ModeratorAgent:
		return v.BaseAgent, true
	case *agents.ResearcherAgent:
		return v.BaseAgent, true
	case *agents.CriticAgent:
		return v.BaseAgent, true
	case *agents.ImplementerAgent:
		return v.BaseAgent, true
	case *agents.UICreatorAgent:
		return v.BaseAgent, true
	}
	return nil, false
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

	// Phase 0: Model Assignment (team leader selects models for agents)
	if err := o.runModelAssignment(); err != nil {
		// Non-fatal: fall back to default model for all agents
		log.Printf("Model assignment skipped: %v", err)
	}

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

	// Phase 5: Concept map — inject into the idea sheet (non-fatal)
	o.appendConceptMap()

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

	// Wire streaming and notify callbacks on the agent's BaseAgent before Process()
	if ba, ok := getBaseAgent(agent); ok {
		if o.OnChunk != nil {
			roleStr := string(role)
			ba.OnChunk = func(chunk string) { o.OnChunk(roleStr, chunk) }
		}
		ba.Notify = func(msg string) { o.notify(msg) }
		defer func() {
			ba.OnChunk = nil
			ba.Notify = nil
		}()
	}

	response, err := agent.Process(o.Discussion, prompt)
	if err != nil {
		log.Printf("Warning: %s contribution failed: %v", agent.GetName(), err)
		return nil // Don't fail the whole discussion
	}

	// Fire evidence callback if the agent returned structured search results
	if len(response.SearchResults) > 0 && o.OnEvidence != nil {
		o.OnEvidence(string(role), response.SearchResults)
	}

	// Add ideas if any were generated
	if len(response.Ideas) > 0 {
		var ideaTitles []string
		for _, idea := range response.Ideas {
			o.Discussion.Ideas = append(o.Discussion.Ideas, idea)
			o.notify(fmt.Sprintf("    💡 New idea: %s", idea.Title))
			ideaTitles = append(ideaTitles, idea.Title)
		}
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

// runModelAssignment asks the team leader to assign models to each agent.
func (o *ConfigurableOrchestrator) runModelAssignment() error {
	o.notify("🧠 Phase 0: Model Assignment")

	leader, ok := o.Agents[models.RoleTeamLeader]
	if !ok {
		return fmt.Errorf("team leader is required for model assignment")
	}

	// Discover available models
	availableModels, err := llm.ListModels(o.BackendConfig)
	if err != nil {
		o.notify(fmt.Sprintf("  ⚠️  Could not list models: %s (using default for all agents)", err))
		return err
	}

	if len(availableModels) == 0 {
		o.notify("  ⚠️  No models returned from API (using default for all agents)")
		return fmt.Errorf("no models available")
	}

	// Build a model list string
	var modelList []string
	for _, m := range availableModels {
		modelList = append(modelList, m.ID)
	}

	// Build the agent roster
	var agentRoster []string
	for role := range o.Agents {
		agentRoster = append(agentRoster, string(role))
	}

	prompt := fmt.Sprintf(`You are assigning LLM models to each agent on our team.

Available models:
%s

Team agents that need a model assigned:
%s

The current default model is: %s

Consider each agent's role when choosing:
- Creative roles (ideation) may benefit from more creative/capable models
- Analytical roles (critic, moderator) may benefit from strong reasoning models
- Research roles need broad knowledge
- UI/visualization roles need good instruction following
- The team leader (you) should use a strong general model

Respond with ONLY a JSON object mapping agent role to model ID. Example:
{"team_leader": "gpt-4o", "ideation": "gpt-4o", "moderator": "gpt-4o-mini"}

JSON response:`, strings.Join(modelList, "\n"), strings.Join(agentRoster, ", "), o.BackendConfig.Model)

	response, err := leader.Process(o.Discussion, prompt)
	if err != nil {
		o.notify(fmt.Sprintf("  ⚠️  Model assignment failed: %s (using default)", err))
		return err
	}

	// Parse the JSON assignments
	assignments := o.parseModelAssignments(response.Content, modelList)
	if len(assignments) == 0 {
		o.notify("  ⚠️  Could not parse model assignments (using default for all agents)")
		return fmt.Errorf("no valid assignments parsed")
	}

	// Apply assignments: reinitialize agents with assigned models
	for role, model := range assignments {
		agentRole := models.AgentRole(role)
		if _, exists := o.Agents[agentRole]; !exists {
			continue
		}
		if err := o.reinitAgent(agentRole, model); err != nil {
			log.Printf("Warning: failed to reassign %s to model %s: %v", role, model, err)
			continue
		}
		o.notify(fmt.Sprintf("  🔧 [%s] → %s", role, model))
	}

	o.notify("  ✅ Model assignments complete")
	return nil
}

// parseModelAssignments extracts a role→model map from the LLM response.
// It validates model IDs against the available list.
func (o *ConfigurableOrchestrator) parseModelAssignments(content string, availableModels []string) map[string]string {
	// Build a set of valid model IDs
	validModels := make(map[string]bool, len(availableModels))
	for _, m := range availableModels {
		validModels[m] = true
	}

	// Try to extract JSON from the response (may be wrapped in markdown)
	jsonStr := extractJSON(content)
	if jsonStr == "" {
		return nil
	}

	var raw map[string]string
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		log.Printf("Warning: could not parse model assignment JSON: %v", err)
		return nil
	}

	// Validate and filter
	result := make(map[string]string)
	for role, model := range raw {
		if validModels[model] {
			result[role] = model
		} else {
			log.Printf("Warning: model %q assigned to %s is not in available list, skipping", model, role)
		}
	}
	return result
}

// extractJSON finds the first JSON object in a string (handles markdown fences).
func extractJSON(s string) string {
	// Strip markdown code fences
	s = strings.ReplaceAll(s, "```json", "")
	s = strings.ReplaceAll(s, "```", "")
	s = strings.TrimSpace(s)

	// Find first { ... }
	start := strings.Index(s, "{")
	if start < 0 {
		return ""
	}
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}

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

// appendConceptMap builds an interactive D3.js concept map from the completed
// Discussion and injects it into the "visualization" message HTML.
// This is non-fatal: any error is logged and the idea sheet is left unchanged.
func (o *ConfigurableOrchestrator) appendConceptMap() {
	o.notify("  🗺️  Generating concept map...")

	data := report.BuildConceptMap(o.Discussion)
	if len(data.Nodes) == 0 {
		return
	}

	mapHTML := report.RenderConceptMapHTML(data)

	// Find the visualization message and inject the concept map before </body>
	for i := range o.Discussion.Messages {
		if o.Discussion.Messages[i].Type == "visualization" {
			o.Discussion.Messages[i].Content = report.InjectIntoHTML(
				o.Discussion.Messages[i].Content, mapHTML,
			)
			o.notify("  ✅ Concept map injected into idea sheet")
			return
		}
	}

	// No visualization message yet — store the map on its own
	o.addMessage(string(models.RoleUICreator), "team", mapHTML, "concept_map")
	o.notify("  ✅ Concept map saved as standalone section")
}
