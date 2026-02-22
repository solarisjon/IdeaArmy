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
	Speech    string // Latest contribution text (truncated for bubble)
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
	CurrentPhase    string
	CurrentRound    int
	TotalRounds     int
	PhaseProgress   float64
	OverallProgress float64

	// Progress bars
	ProgressBar progress.Model

	// Ideas generated
	Ideas []*models.Idea

	// Messages
	Messages    []string
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
	Speech  string // Truncated latest contribution text
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

// autoQuitMsg triggers automatic exit after completion
type autoQuitMsg struct{}

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
			Name:    getPersona(string(role)).Name,
			Status:  "idle",
			Spinner: s,
		}
	}

	// Create progress bar
	prog := progress.New(progress.WithDefaultGradient())

	return Model{
		TeamConfig:   config,
		Topic:        topic,
		Agents:       agents,
		CurrentPhase: "Initializing",
		TotalRounds:  config.MaxRounds,
		ProgressBar:  prog,
		Ideas:        []*models.Idea{},
		Messages:     []string{},
		MaxMessages:  10,
		Status:       "initializing",
		StartTime:    time.Now(),
		Width:        80,
		Height:       24,
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
			if msg.Speech != "" {
				agent.Speech = msg.Speech
			}
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
		// Auto-exit after a brief pause so user can see final state
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return autoQuitMsg{}
		})

	case autoQuitMsg:
		return m, tea.Quit

	case ErrorMsg:
		m.Status = "error"
		m.ErrorMessage = msg.Err.Error()
		return m, nil
	}

	return m, nil
}

// View renders the war room UI
func (m Model) View() string {
	if m.Width == 0 {
		return "Initializing..."
	}

	var sections []string

	// War room header
	sections = append(sections, m.renderHeader())

	// Progress bar
	sections = append(sections, m.renderProgress())

	// Agent grid with speech bubbles
	sections = append(sections, m.renderWarRoom())

	// Ideas board
	if len(m.Ideas) > 0 {
		sections = append(sections, m.renderIdeas())
	}

	// Status bar
	sections = append(sections, m.renderStatus())

	// Footer
	if m.Status == "running" || m.Status == "initializing" {
		sections = append(sections, systemMessageStyle.Render("  Press 'q' to quit"))
	} else if m.Status == "complete" {
		sections = append(sections, systemMessageStyle.Render("  Press 'q' to exit"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderHeader() string {
	title := titleStyle.Render("⚔️  The War Room")
	topic := subtitleStyle.Render("  Mission: ") + m.Topic
	return lipgloss.JoinVertical(lipgloss.Left, title, topic)
}

func (m Model) renderTeam() string {
	// Not used in war room layout — kept for interface compat
	return ""
}

func (m Model) renderProgress() string {
	phaseText := phaseStyle.Render(fmt.Sprintf("  ⚙ %s", m.CurrentPhase))
	roundText := fmt.Sprintf("Round %d/%d", m.CurrentRound, m.TotalRounds)

	progressBar := m.ProgressBar.ViewAs(m.OverallProgress)
	progressPercent := fmt.Sprintf("%.0f%%", m.OverallProgress*100)

	header := lipgloss.JoinHorizontal(lipgloss.Left, phaseText, "  ", roundText)
	prog := lipgloss.JoinHorizontal(lipgloss.Left, "  ", progressBar, " ", progressPercent)

	return lipgloss.JoinVertical(lipgloss.Left, header, prog)
}

func (m Model) renderWarRoom() string {
	roles := m.TeamConfig.GetActiveAgentRoles()

	// Calculate card width — fit 2 per row with some padding
	cardWidth := (m.Width - 8) / 2
	if cardWidth < 30 {
		cardWidth = 30
	}
	if cardWidth > 50 {
		cardWidth = 50
	}

	var cards []string
	for _, role := range roles {
		agent, ok := m.Agents[string(role)]
		if !ok {
			continue
		}
		cards = append(cards, m.renderAgentCard(agent, cardWidth))
	}

	// Lay out in 2-column grid
	var rows []string
	for i := 0; i < len(cards); i += 2 {
		if i+1 < len(cards) {
			row := lipgloss.JoinHorizontal(lipgloss.Top, "  ", cards[i], "  ", cards[i+1])
			rows = append(rows, row)
		} else {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, "  ", cards[i]))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, "", strings.Join(rows, "\n"))
}

func (m Model) renderAgentCard(agent *AgentState, width int) string {
	persona := getPersona(agent.Role)
	nameStyle := lipgloss.NewStyle().Foreground(persona.Color).Bold(true)

	// Header line: icon + name + status indicator
	var statusIndicator string
	switch agent.Status {
	case "working":
		statusIndicator = agent.Spinner.View()
	case "complete":
		statusIndicator = checkmarkStyle.Render("✓")
	default:
		statusIndicator = lipgloss.NewStyle().Foreground(gray).Render("○")
	}

	header := fmt.Sprintf("%s %s %s", persona.Icon, nameStyle.Render(persona.Name), statusIndicator)

	// Speech bubble content
	bubbleWidth := width - 4
	if bubbleWidth < 20 {
		bubbleWidth = 20
	}

	var speechContent string
	switch agent.Status {
	case "working":
		if agent.Speech != "" {
			speechContent = truncateText(agent.Speech, bubbleWidth, 3)
		} else {
			speechContent = lipgloss.NewStyle().Foreground(yellow).Italic(true).Render("🗣️ " + agent.Message)
		}
	case "complete":
		if agent.Speech != "" {
			speechContent = truncateText(agent.Speech, bubbleWidth, 3)
		} else {
			speechContent = lipgloss.NewStyle().Foreground(green).Render("✅ Done")
		}
	default:
		speechContent = lipgloss.NewStyle().Foreground(gray).Italic(true).Render("💤 " + persona.Tagline + "...")
	}

	bubble := speechBubbleStyle.
		Copy().
		Width(bubbleWidth).
		BorderForeground(persona.Color).
		Render(speechContent)

	return lipgloss.JoinVertical(lipgloss.Left, header, bubble)
}

// renderAgents kept as alias for backward compat
func (m Model) renderAgents() string {
	return m.renderWarRoom()
}

func (m Model) renderIdeas() string {
	header := lipgloss.NewStyle().Foreground(green).Bold(true).
		Render(fmt.Sprintf("  📋 Ideas on the Board (%d)", len(m.Ideas)))

	var ideaLines []string
	start := 0
	if len(m.Ideas) > 5 {
		start = len(m.Ideas) - 5
	}

	for i := start; i < len(m.Ideas); i++ {
		idea := m.Ideas[i]
		num := fmt.Sprintf("%d.", i+1)

		scoreStr := ""
		if idea.Validated && idea.Score > 0 {
			if idea.Score >= 8.0 {
				scoreStr = ideaScoreStyle.Render(fmt.Sprintf(" ⭐ %.1f/10", idea.Score))
			} else {
				scoreStr = ideaScoreStyle.Render(fmt.Sprintf(" [%.1f/10]", idea.Score))
			}
		}

		title := ideaTitleStyle.Render(idea.Title)
		ideaLines = append(ideaLines, fmt.Sprintf("    %s %s%s", num, title, scoreStr))
	}

	if start > 0 {
		ideaLines = append([]string{fmt.Sprintf("    ... and %d more", start)}, ideaLines...)
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
		statusLine = statusRunningStyle.Render("  ⏳ Assembling the team...")
	case "running":
		elapsed := time.Since(m.StartTime)
		statusLine = statusRunningStyle.Render(fmt.Sprintf("  ⚡ Discussion in progress... (%s)", formatDuration(elapsed)))
	case "complete":
		duration := m.EndTime.Sub(m.StartTime)
		statusLine = statusCompleteStyle.Render(fmt.Sprintf("  ✅ Mission accomplished! (%s)", formatDuration(duration)))
	case "error":
		statusLine = statusErrorStyle.Render(fmt.Sprintf("  ❌ Error: %s", m.ErrorMessage))
	}

	stats := fmt.Sprintf("  💡 %d ideas | 📨 %d messages", m.TotalIdeas, m.TotalMessages)

	return lipgloss.JoinVertical(lipgloss.Left, "", statusLine, systemMessageStyle.Render(stats))
}

func getAgentName(role string) string {
	return getPersona(role).Name
}

// truncateText wraps/truncates text to fit inside a speech bubble
func truncateText(text string, width int, maxLines int) string {
	if width < 10 {
		width = 10
	}
	// Word-wrap into lines
	var lines []string
	words := strings.Fields(text)
	current := ""
	for _, w := range words {
		if current == "" {
			current = w
		} else if len(current)+1+len(w) <= width {
			current += " " + w
		} else {
			lines = append(lines, current)
			current = w
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines[maxLines-1] += "..."
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
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
