package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/ai-agent-team/internal/models"
)

// AgentState represents the current state of an agent
type AgentState struct {
	Role      string
	Name      string
	Status    string // "idle", "working", "complete"
	Message   string
	Spinner   spinner.Model
	StartTime time.Time
}

// Model represents the TUI state
type Model struct {
	// Configuration
	TeamConfig *models.TeamConfig
	Topic      string

	// Agent states
	Agents map[string]*AgentState

	// Discussion progress
	CurrentPhase   string
	CurrentRound   int
	TotalRounds    int
	PhaseProgress  float64
	OverallProgress float64

	// Progress bars
	ProgressBar progress.Model

	// Ideas generated
	Ideas []*models.Idea

	// Messages
	Messages []string
	MaxMessages int

	// Status
	Status       string // "initializing", "running", "complete", "error"
	ErrorMessage string
	StartTime    time.Time
	EndTime      time.Time

	// Terminal dimensions
	Width  int
	Height int

	// Discussion stats
	TotalIdeas    int
	TotalMessages int
}

// ProgressMsg is sent to update progress
type ProgressMsg struct {
	Phase    string
	Round    int
	Progress float64
}

// AgentUpdateMsg is sent when an agent status changes
type AgentUpdateMsg struct {
	Role    string
	Status  string
	Message string
}

// IdeaGeneratedMsg is sent when a new idea is created
type IdeaGeneratedMsg struct {
	Idea *models.Idea
}

// LogMsg is sent to add a log message
type LogMsg string

// CompleteMsg is sent when discussion completes
type CompleteMsg struct {
	Discussion *models.Discussion
}

// ErrorMsg is sent when an error occurs
type ErrorMsg struct {
	Err error
}

// NewModel creates a new TUI model
func NewModel(config *models.TeamConfig, topic string) Model {
	// Create spinners for each agent
	agents := make(map[string]*AgentState)

	roles := config.GetActiveAgentRoles()
	for _, role := range roles {
		s := spinner.New()
		s.Spinner = spinner.Dot
		s.Style = spinnerStyle

		agents[string(role)] = &AgentState{
			Role:    string(role),
			Name:    getAgentName(string(role)),
			Status:  "idle",
			Spinner: s,
		}
	}

	// Create progress bar
	prog := progress.New(progress.WithDefaultGradient())

	return Model{
		TeamConfig:      config,
		Topic:          topic,
		Agents:         agents,
		CurrentPhase:   "Initializing",
		TotalRounds:    config.MaxRounds,
		ProgressBar:    prog,
		Ideas:          []*models.Idea{},
		Messages:       []string{},
		MaxMessages:    10,
		Status:         "initializing",
		StartTime:      time.Now(),
		Width:          80,
		Height:         24,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// Start all spinners
	var cmds []tea.Cmd
	for _, agent := range m.Agents {
		cmds = append(cmds, agent.Spinner.Tick)
	}
	return tea.Batch(cmds...)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmds []tea.Cmd
		for _, agent := range m.Agents {
			if agent.Status == "working" {
				var cmd tea.Cmd
				agent.Spinner, cmd = agent.Spinner.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)

	case ProgressMsg:
		m.CurrentPhase = msg.Phase
		m.CurrentRound = msg.Round
		m.PhaseProgress = msg.Progress
		m.OverallProgress = (float64(msg.Round-1) + msg.Progress) / float64(m.TotalRounds)
		return m, nil

	case AgentUpdateMsg:
		if agent, ok := m.Agents[msg.Role]; ok {
			agent.Status = msg.Status
			agent.Message = msg.Message
			if msg.Status == "working" {
				agent.StartTime = time.Now()
			}
		}
		return m, nil

	case IdeaGeneratedMsg:
		m.Ideas = append(m.Ideas, msg.Idea)
		m.TotalIdeas++
		return m, nil

	case LogMsg:
		m.Messages = append(m.Messages, string(msg))
		if len(m.Messages) > m.MaxMessages {
			m.Messages = m.Messages[1:]
		}
		m.TotalMessages++
		return m, nil

	case CompleteMsg:
		m.Status = "complete"
		m.EndTime = time.Now()
		return m, nil

	case ErrorMsg:
		m.Status = "error"
		m.ErrorMessage = msg.Err.Error()
		return m, nil
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.Width == 0 {
		return "Initializing..."
	}

	var sections []string

	// Header
	sections = append(sections, m.renderHeader())

	// Team composition
	sections = append(sections, m.renderTeam())

	// Current phase and progress
	sections = append(sections, m.renderProgress())

	// Active agents
	sections = append(sections, m.renderAgents())

	// Ideas generated
	if len(m.Ideas) > 0 {
		sections = append(sections, m.renderIdeas())
	}

	// Recent messages
	if len(m.Messages) > 0 {
		sections = append(sections, m.renderMessages())
	}

	// Status and stats
	sections = append(sections, m.renderStatus())

	// Footer
	if m.Status == "running" {
		sections = append(sections, systemMessageStyle.Render("\nPress 'q' to quit"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderHeader() string {
	title := titleStyle.Render("🤖 AI Agent Team - Collaborative Ideation")
	topic := subtitleStyle.Render("Topic: ") + m.Topic
	return lipgloss.JoinVertical(lipgloss.Left, title, topic, "")
}

func (m Model) renderTeam() string {
	var agents []string
	for _, role := range m.TeamConfig.GetActiveAgentRoles() {
		if agent, ok := m.Agents[string(role)]; ok {
			icon := getAgentIcon(agent.Role)
			name := agent.Name
			agents = append(agents, fmt.Sprintf("%s %s", icon, name))
		}
	}

	header := lipgloss.NewStyle().Foreground(blue).Bold(true).Render("Team Composition:")
	teamList := lipgloss.NewStyle().Foreground(lightGray).Render(strings.Join(agents, " • "))

	return lipgloss.JoinVertical(lipgloss.Left, header, teamList, "")
}

func (m Model) renderProgress() string {
	phaseText := phaseStyle.Render(fmt.Sprintf("Phase: %s", m.CurrentPhase))
	roundText := fmt.Sprintf("Round %d/%d", m.CurrentRound, m.TotalRounds)

	progressBar := m.ProgressBar.ViewAs(m.OverallProgress)
	progressPercent := fmt.Sprintf("%.0f%%", m.OverallProgress*100)

	header := lipgloss.JoinHorizontal(lipgloss.Left, phaseText, "  ", roundText)
	progress := lipgloss.JoinHorizontal(lipgloss.Left, progressBar, " ", progressPercent)

	return lipgloss.JoinVertical(lipgloss.Left, "", header, progress, "")
}

func (m Model) renderAgents() string {
	var agentLines []string

	for _, role := range m.TeamConfig.GetActiveAgentRoles() {
		agent, ok := m.Agents[string(role)]
		if !ok {
			continue
		}

		icon := getAgentIcon(agent.Role)
		name := agent.Name

		var statusStr string
		switch agent.Status {
		case "working":
			statusStr = agent.Spinner.View() + " " + agent.Message
		case "complete":
			statusStr = checkmarkStyle.Render("✓") + " Complete"
		case "idle":
			statusStr = agentIdleStyle.Render("Ready")
		}

		color := getAgentColor(agent.Role)
		nameStyled := lipgloss.NewStyle().Foreground(color).Bold(true).Render(fmt.Sprintf("%s %s", icon, name))

		line := fmt.Sprintf("  %s: %s", nameStyled, statusStr)
		agentLines = append(agentLines, line)
	}

	header := lipgloss.NewStyle().Foreground(blue).Bold(true).Render("Agent Status:")
	return lipgloss.JoinVertical(lipgloss.Left, "", header, strings.Join(agentLines, "\n"), "")
}

func (m Model) renderIdeas() string {
	header := lipgloss.NewStyle().Foreground(green).Bold(true).Render(fmt.Sprintf("💡 Ideas Generated (%d):", len(m.Ideas)))

	var ideaLines []string
	// Show last 3 ideas
	start := 0
	if len(m.Ideas) > 3 {
		start = len(m.Ideas) - 3
	}

	for i := start; i < len(m.Ideas); i++ {
		idea := m.Ideas[i]
		title := ideaTitleStyle.Render(idea.Title)

		scoreStr := ""
		if idea.Validated && idea.Score > 0 {
			scoreStr = ideaScoreStyle.Render(fmt.Sprintf(" [%.1f/10]", idea.Score))
		}

		ideaLines = append(ideaLines, fmt.Sprintf("  • %s%s", title, scoreStr))
	}

	return lipgloss.JoinVertical(lipgloss.Left, "", header, strings.Join(ideaLines, "\n"))
}

func (m Model) renderMessages() string {
	header := lipgloss.NewStyle().Foreground(lightGray).Italic(true).Render("Recent Activity:")

	var msgLines []string
	for _, msg := range m.Messages {
		msgLines = append(msgLines, systemMessageStyle.Render("  → "+msg))
	}

	return lipgloss.JoinVertical(lipgloss.Left, "", header, strings.Join(msgLines, "\n"))
}

func (m Model) renderStatus() string {
	var statusLine string

	switch m.Status {
	case "initializing":
		statusLine = statusRunningStyle.Render("⏳ Initializing...")
	case "running":
		elapsed := time.Since(m.StartTime)
		statusLine = statusRunningStyle.Render(fmt.Sprintf("⚡ Running... (%s)", formatDuration(elapsed)))
	case "complete":
		duration := m.EndTime.Sub(m.StartTime)
		statusLine = statusCompleteStyle.Render(fmt.Sprintf("✅ Complete! (%s)", formatDuration(duration)))
	case "error":
		statusLine = statusErrorStyle.Render(fmt.Sprintf("❌ Error: %s", m.ErrorMessage))
	}

	stats := fmt.Sprintf("  Ideas: %d | Messages: %d", m.TotalIdeas, m.TotalMessages)

	return lipgloss.JoinVertical(lipgloss.Left, "", statusLine, systemMessageStyle.Render(stats))
}

func getAgentName(role string) string {
	switch role {
	case "team_leader":
		return "Team Leader"
	case "ideation":
		return "Ideation Specialist"
	case "moderator":
		return "Moderator"
	case "researcher":
		return "Researcher"
	case "critic":
		return "Critical Analyst"
	case "implementer":
		return "Implementation Specialist"
	case "ui_creator":
		return "UI Creator"
	default:
		return role
	}
}

func formatDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	seconds = seconds % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}
