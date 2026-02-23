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
	"team_leader": {Name: "Commander Bleep", Icon: "🤖", Tagline: "beep-boop, let's rally!", Color: lipgloss.Color("#FF6B6B")},
	"ideation":    {Name: "Sparx", Icon: "💡", Tagline: "zapping up wild ideas", Color: lipgloss.Color("#51E898")},
	"moderator":   {Name: "Balancebot", Icon: "🔮", Tagline: "keeping the circuits aligned", Color: lipgloss.Color("#7B68EE")},
	"researcher":  {Name: "Digger-3000", Icon: "🔍", Tagline: "scanning all known databases", Color: lipgloss.Color("#00D4FF")},
	"critic":      {Name: "Glitchy", Icon: "👾", Tagline: "poking the logic boards", Color: lipgloss.Color("#FFD93D")},
	"implementer": {Name: "Bolt", Icon: "🔧", Tagline: "tightening the bolts", Color: lipgloss.Color("#FF8C42")},
	"ui_creator":  {Name: "Doodlebot", Icon: "🎨", Tagline: "painting pixels with love", Color: lipgloss.Color("#FF6BC1")},
}

func getPersona(role string) AgentPersona {
	if p, ok := agentPersonas[role]; ok {
		return p
	}
	return AgentPersona{Name: role, Icon: "🤖", Tagline: "booting up", Color: robotGray}
}

var (
	// Candy color palette
	hotPink      = lipgloss.Color("#FF6BC1")
	neonMint     = lipgloss.Color("#51E898")
	electricCyan = lipgloss.Color("#00D4FF")
	brightYellow = lipgloss.Color("#FFD93D")
	coral        = lipgloss.Color("#FF6B6B")
	tangerine    = lipgloss.Color("#FF8C42")
	slatePurple  = lipgloss.Color("#7B68EE")
	robotGray    = lipgloss.Color("#8892A0")
	white        = lipgloss.Color("#FFFFFF")
	darkGray     = lipgloss.Color("#374151")

	// Base styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(hotPink).
			MarginTop(1).
			MarginBottom(0)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(electricCyan).
			Bold(true)

	// Agent styles
	agentActiveStyle = lipgloss.NewStyle().
				Foreground(neonMint).
				Bold(true)

	agentIdleStyle = lipgloss.NewStyle().
			Foreground(robotGray)

	agentCompleteStyle = lipgloss.NewStyle().
				Foreground(electricCyan)

	// Message styles
	systemMessageStyle = lipgloss.NewStyle().
				Foreground(robotGray).
				Italic(true)

	agentMessageStyle = lipgloss.NewStyle().
				Foreground(white).
				PaddingLeft(2)

	// Progress styles
	progressBarStyle = lipgloss.NewStyle().
				Foreground(hotPink)

	phaseStyle = lipgloss.NewStyle().
			Foreground(brightYellow).
			Bold(true)

	// Idea styles
	ideaTitleStyle = lipgloss.NewStyle().
			Foreground(neonMint).
			Bold(true)

	ideaScoreStyle = lipgloss.NewStyle().
			Foreground(brightYellow)

	// Box styles
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(slatePurple).
			Padding(0, 1)

	summaryBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(neonMint).
			Padding(1, 2)

	// Speech bubble style
	speechBubbleStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(robotGray).
				Padding(0, 1)

	// Status styles
	statusRunningStyle = lipgloss.NewStyle().
				Foreground(brightYellow).
				Bold(true)

	statusCompleteStyle = lipgloss.NewStyle().
				Foreground(neonMint).
				Bold(true)

	statusErrorStyle = lipgloss.NewStyle().
				Foreground(coral).
				Bold(true)

	// Spinner and indicator styles
	spinnerStyle = lipgloss.NewStyle().
			Foreground(hotPink)

	checkmarkStyle = lipgloss.NewStyle().
			Foreground(neonMint).
			Bold(true)

	errorMarkStyle = lipgloss.NewStyle().
			Foreground(coral).
			Bold(true)
)

// Backward compat helpers used elsewhere
func getAgentColor(role string) lipgloss.Color {
	return getPersona(role).Color
}

func getAgentIcon(role string) string {
	return getPersona(role).Icon
}
