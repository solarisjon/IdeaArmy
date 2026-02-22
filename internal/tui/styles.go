package tui

import "github.com/charmbracelet/lipgloss"

// AgentPersona holds the fun identity for each agent role
type AgentPersona struct {
	Name    string
	Icon    string
	Tagline string
	Color   lipgloss.Color
}

var agentPersonas = map[string]AgentPersona{
	"team_leader": {Name: "Captain Rex", Icon: "🎖️", Tagline: "rallying the troops", Color: lipgloss.Color("#FFD700")},
	"ideation":    {Name: "Sparky", Icon: "⚡", Tagline: "igniting ideas", Color: lipgloss.Color("#10B981")},
	"moderator":   {Name: "The Judge", Icon: "⚖️", Tagline: "keeping order", Color: lipgloss.Color("#3B82F6")},
	"researcher":  {Name: "Doc Sage", Icon: "📖", Tagline: "digging deep", Color: lipgloss.Color("#8B5CF6")},
	"critic":      {Name: "Nitpick", Icon: "🧐", Tagline: "poking holes", Color: lipgloss.Color("#F59E0B")},
	"implementer": {Name: "Wrench", Icon: "🔩", Tagline: "making it real", Color: lipgloss.Color("#06B6D4")},
	"ui_creator":  {Name: "Pixel", Icon: "🎨", Tagline: "painting the vision", Color: lipgloss.Color("#EC4899")},
}

func getPersona(role string) AgentPersona {
	if p, ok := agentPersonas[role]; ok {
		return p
	}
	return AgentPersona{Name: role, Icon: "🤖", Tagline: "working", Color: gray}
}

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
	darkGray  = lipgloss.Color("#374151")

	// Base styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(purple).
			MarginTop(1).
			MarginBottom(0)

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

	// Speech bubble style
	speechBubbleStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(gray).
				Padding(0, 1)

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

// Backward compat helpers used elsewhere
func getAgentColor(role string) lipgloss.Color {
	return getPersona(role).Color
}

func getAgentIcon(role string) string {
	return getPersona(role).Icon
}
