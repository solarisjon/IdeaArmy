package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Color palette
	purple    = lipgloss.Color("#7C3AED")
	blue      = lipgloss.Color("#3B82F6")
	green     = lipgloss.Color("#10B981")
	yellow    = lipgloss.Color("#F59E0B")
	red       = lipgloss.Color("#EF4444")
	gray      = lipgloss.Color("#6B7280")
	lightGray = lipgloss.Color("#9CA3AF")
	white     = lipgloss.Color("#FFFFFF")

	// Base styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(purple).
			MarginTop(1).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(blue).
			Bold(true)

	// Agent styles
	agentActiveStyle = lipgloss.NewStyle().
				Foreground(green).
				Bold(true)

	agentIdleStyle = lipgloss.NewStyle().
			Foreground(gray)

	agentCompleteStyle = lipgloss.NewStyle().
				Foreground(blue)

	// Message styles
	systemMessageStyle = lipgloss.NewStyle().
				Foreground(lightGray).
				Italic(true)

	agentMessageStyle = lipgloss.NewStyle().
				Foreground(white).
				PaddingLeft(2)

	// Progress styles
	progressBarStyle = lipgloss.NewStyle().
				Foreground(purple)

	phaseStyle = lipgloss.NewStyle().
			Foreground(yellow).
			Bold(true)

	// Idea styles
	ideaTitleStyle = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)

	ideaScoreStyle = lipgloss.NewStyle().
			Foreground(yellow)

	// Box styles
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purple).
			Padding(0, 1)

	summaryBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(green).
			Padding(1, 2)

	// Status styles
	statusRunningStyle = lipgloss.NewStyle().
				Foreground(yellow).
				Bold(true)

	statusCompleteStyle = lipgloss.NewStyle().
				Foreground(green).
				Bold(true)

	statusErrorStyle = lipgloss.NewStyle().
				Foreground(red).
				Bold(true)

	// Spinner and indicator styles
	spinnerStyle = lipgloss.NewStyle().
			Foreground(purple)

	checkmarkStyle = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)

	errorMarkStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)
)

// Agent role colors
func getAgentColor(role string) lipgloss.Color {
	switch role {
	case "team_leader":
		return lipgloss.Color("#FFD700") // Gold
	case "ideation":
		return lipgloss.Color("#10B981") // Green
	case "moderator":
		return lipgloss.Color("#3B82F6") // Blue
	case "researcher":
		return lipgloss.Color("#8B5CF6") // Purple
	case "critic":
		return lipgloss.Color("#F59E0B") // Orange
	case "implementer":
		return lipgloss.Color("#06B6D4") // Cyan
	case "ui_creator":
		return lipgloss.Color("#EC4899") // Pink
	default:
		return gray
	}
}

// Get agent icon
func getAgentIcon(role string) string {
	switch role {
	case "team_leader":
		return "🎯"
	case "ideation":
		return "💡"
	case "moderator":
		return "🔍"
	case "researcher":
		return "📚"
	case "critic":
		return "🤔"
	case "implementer":
		return "🔧"
	case "ui_creator":
		return "🎨"
	default:
		return "🤖"
	}
}
