# AI Agent Team v2 - Configurable Multi-Agent Ideation System

A sophisticated multi-agent AI system built in Go that orchestrates **configurable teams of specialized AI agents** to collaboratively explore complex ideas through **multi-round discussions**, validate concepts, and produce beautiful visualizations.

## 🆕 What's New in v2

### Configurable Team Composition
- **Choose your team size**: 4, 6, or 7 agents based on your needs
- **3 preset configurations**: Standard (quick), Extended (deep), Full (maximum depth)
- **Custom teams**: Mix and match agents for specific use cases

### New Specialized Agents
- **Researcher** 📚 - Provides factual grounding and real-world context
- **Critic** 🤔 - Challenges assumptions and identifies risks constructively
- **Implementer** 🔧 - Focuses on practical execution and implementation planning

### Multi-Round Discussions
- **Iterative refinement**: Ideas evolve through multiple discussion rounds
- **Leader facilitation**: Team Leader synthesizes each round and guides the team
- **Agent-to-agent dialogue**: Agents build on each other's contributions
- **Deeper exploration**: 1-3 rounds depending on configuration

### Smarter Orchestration
- Agents contribute in logical sequence (research → ideation → criticism → implementation)
- Leader synthesizes after each round
- More comprehensive final evaluation
- Better context building across rounds

## The Team

### Core Agents (Always Included)

**🎯 Team Leader**
- Orchestrates multi-round discussions
- Synthesizes contributions after each round
- Makes final decisions on idea selection
- Ensures all perspectives are heard

**🎨 UI Creator**
- Creates beautiful HTML visualizations
- Designs comprehensive idea sheets
- Presents discussion journey and outcomes

### Standard Team Agents

**💡 Ideation Agent**
- Generates creative, well-researched ideas
- Builds on previous round feedback
- Explores concepts from multiple angles

**🔍 Moderator/Facilitator**
- Validates idea quality and feasibility
- Provides objective scoring (0-10 scale)
- Identifies pros, cons, and potential issues

### Extended Team Agents

**📚 Researcher** (New in v2!)
- Conducts deep research and analysis
- Provides data, statistics, and evidence
- References case studies and real-world examples
- Grounds ideas in factual reality

**🤔 Critic** (New in v2!)
- Challenges underlying assumptions
- Identifies potential failure modes and risks
- Asks difficult questions constructively
- Ensures ideas are robust

### Full Team Agents

**🔧 Implementer** (New in v2!)
- Focuses on practical implementation
- Breaks ideas into actionable steps
- Identifies technical requirements
- Proposes concrete execution approaches

## Team Configurations

### ⚡ Standard (4 agents, 1 round)
**Perfect for:** Quick ideation, focused exploration, time-sensitive projects

**Team:** Leader, Ideation, Moderator, UI Creator
**Duration:** ~1-2 minutes
**Depth:** Good balance of speed and quality

### 🔬 Extended (6 agents, 2 rounds)
**Perfect for:** Important decisions, complex topics, deeper analysis

**Team:** Standard + Researcher, Critic
**Rounds:** 2 with iterative refinement
**Duration:** ~3-5 minutes
**Depth:** Thorough exploration with critical analysis

### 🚀 Full (7 agents, 3 rounds)
**Perfect for:** Strategic initiatives, product planning, comprehensive exploration

**Team:** Extended + Implementer
**Rounds:** 3 with extensive discussion
**Duration:** ~5-10 minutes
**Depth:** Maximum depth with implementation planning

## Quick Start

### Option 1: Beautiful TUI (Terminal UI) ✨ NEW!

Experience the discussion in real-time with dynamic spinners, progress bars, and live updates!

```bash
# Set your API key
export ANTHROPIC_API_KEY="your-key-here"
# or
export ANTHROPIC_KEY="your-key-here"

# Run the beautiful TUI
./bin/cli-tui

# Or with go run
go run cmd/cli/main_tui.go
```

**Features:**
- 🔄 Live agent spinners showing who's working
- 📊 Real-time progress bars
- 💡 Ideas appear as they're generated
- 🎨 Color-coded agents with unique icons
- ⏱️ Live timer and statistics
- 📄 Comprehensive multi-section reports with runner-ups

[See TUI_GUIDE.md for screenshots | REPORTS.md for report structure]

### Option 2: Standard CLI

```bash
# Run the v2 CLI
./bin/cli-v2

# Or with go run
go run cmd/cli/main_v2.go
```

Clean text output, same powerful team collaboration.

### Option 3: Web Interface

```bash
# Start the v2 server
./bin/server-v2

# Or with go run
go run cmd/server/main_v2.go
```

Then open `http://localhost:8080`

The web interface lets you:
- Choose team configuration with one click
- See real-time progress with round indicators
- View the beautiful HTML output inline

## How Multi-Round Discussions Work

### Round Structure

Each round follows this pattern:

1. **Research** (if Researcher is included)
   - Provides factual context
   - Shares relevant examples and data

2. **Ideation**
   - Generates or refines ideas
   - Builds on previous feedback

3. **Critical Analysis** (if Critic is included)
   - Challenges assumptions
   - Identifies potential risks

4. **Implementation Planning** (if Implementer is included)
   - Proposes execution approach
   - Identifies requirements

5. **Leader Synthesis**
   - Summarizes the round
   - Guides the next round

### Example: Extended Team (2 Rounds)

**Round 1:**
- Researcher provides market context
- Ideation generates 4-5 initial ideas
- Critic challenges key assumptions
- Leader synthesizes: "Focus on ideas 2 and 4, address the scalability concerns"

**Round 2:**
- Researcher provides implementation examples
- Ideation refines ideas 2 and 4 based on feedback
- Critic validates the refinements
- Leader selects the best idea

**Final:**
- Moderator provides detailed scores
- Leader makes final selection
- UI Creator generates visualization

## Installation & Setup

### Prerequisites
- Go 1.21 or higher
- Anthropic API key ([get one here](https://console.anthropic.com/))

### Installation

```bash
# Clone or navigate to the project
cd ai-agent-team

# Install dependencies
go mod download

# Build both versions
go build -o bin/cli ./cmd/cli/main.go         # v1
go build -o bin/cli-v2 ./cmd/cli/main_v2.go   # v2
go build -o bin/server ./cmd/server/main.go   # v1
go build -o bin/server-v2 ./cmd/server/main_v2.go  # v2
```

## Project Structure

```
ai-agent-team/
├── cmd/
│   ├── cli/
│   │   ├── main.go         # v1 CLI
│   │   ├── main_v2.go      # v2 CLI with team config
│   │   └── main_tui.go     # v2 CLI with beautiful TUI
│   └── server/
│       ├── main.go         # v1 server
│       └── main_v2.go      # v2 server with team config
├── internal/
│   ├── agents/
│   │   ├── agent.go        # Base agent interface
│   │   ├── team_leader.go  # Team Leader
│   │   ├── ideation.go     # Ideation Agent
│   │   ├── moderator.go    # Moderator
│   │   ├── researcher.go   # Researcher (NEW)
│   │   ├── critic.go       # Critic (NEW)
│   │   ├── implementer.go  # Implementer (NEW)
│   │   └── ui_creator.go   # UI Creator
│   ├── orchestrator/
│   │   ├── orchestrator.go    # v1 orchestrator
│   │   └── orchestrator_v2.go # v2 with multi-round support
│   ├── models/
│   │   ├── types.go        # Data structures
│   │   └── config.go       # Team configurations (NEW)
│   ├── tui/                # Terminal UI (NEW)
│   │   ├── model.go        # Bubbletea model
│   │   ├── styles.go       # Lipgloss styles
│   │   └── runner.go       # TUI runner
│   └── claude/
│       └── client.go       # Anthropic API client
├── bin/
│   ├── cli, cli-v2, cli-tui
│   └── server, server-v2
├── README.md               # v1 documentation
├── README_V2.md           # This file - v2 documentation
├── TUI_GUIDE.md           # Beautiful TUI guide (NEW)
├── QUICKSTART.md          # Quick start guide
└── USAGE.md               # Detailed usage guide
```

## Configuration Options

### Programmatic Configuration

```go
import "github.com/yourusername/ai-agent-team/internal/models"

// Use a preset
config := models.ExtendedTeamConfig()

// Or customize
config := &models.TeamConfig{
    IncludeTeamLeader:  true,   // Required
    IncludeUICreator:   true,   // Required
    IncludeIdeation:    true,
    IncludeModerator:   true,
    IncludeResearcher:  true,   // Enable researcher
    IncludeCritic:      true,   // Enable critic
    IncludeImplementer: false,  // Disable implementer
    MaxRounds:          2,      // 2 discussion rounds
    MinIdeas:           4,
    DeepDive:           true,
    MinScoreThreshold:  7.0,
}

orch := orchestrator.NewConfigurableOrchestrator(apiKey, config)
```

## Example Topics by Configuration

### Standard Team Topics
- "Mobile app features for improving daily habit tracking"
- "Content ideas for a tech startup blog"
- "UX improvements for an e-commerce checkout flow"

### Extended Team Topics
- "Strategic approach to entering the sustainable packaging market"
- "Comprehensive employee retention program for remote teams"
- "Product roadmap for an AI-powered writing assistant"

### Full Team Topics
- "Complete go-to-market strategy for a new SaaS product"
- "End-to-end solution for reducing food waste in supply chains"
- "Platform architecture for a decentralized social network"

## Performance & Costs

### Typical Discussion Times
- **Standard**: 1-2 minutes
- **Extended**: 3-5 minutes
- **Full**: 5-10 minutes

### API Usage Per Discussion
- **Standard**: 5-8 API calls, ~10,000-20,000 tokens ($0.15-$0.30)
- **Extended**: 12-18 API calls, ~25,000-40,000 tokens ($0.40-$0.70)
- **Full**: 20-30 API calls, ~40,000-60,000 tokens ($0.70-$1.20)

*Costs based on Claude Sonnet 4 pricing as of Feb 2025*

## When to Use Each Configuration

### Use Standard When:
- ✅ You need quick results
- ✅ The topic is straightforward
- ✅ You want a focused approach
- ✅ Time is limited

### Use Extended When:
- ✅ The topic is complex or multifaceted
- ✅ You need deeper analysis
- ✅ Critical evaluation is important
- ✅ You want research-backed ideas

### Use Full When:
- ✅ Strategic or high-impact decisions
- ✅ You need implementation planning
- ✅ Multiple perspectives are crucial
- ✅ Time and thoroughness > speed

## Advanced Usage

### Custom Team Example

```go
// Build a specialized team for technical architecture decisions
config := &models.TeamConfig{
    IncludeTeamLeader:  true,
    IncludeIdeation:    true,
    IncludeModerator:   true,
    IncludeResearcher:  true,  // For technology research
    IncludeCritic:      true,  // For technical risk analysis
    IncludeImplementer: true,  // For architecture planning
    IncludeUICreator:   false, // Skip visualization
    MaxRounds:          3,     // Deep technical exploration
    DeepDive:           true,
}
```

## Comparison: v1 vs v2

| Feature | v1 | v2 |
|---------|----|----|
| Team Size | Fixed (4 agents) | Configurable (4-7 agents) |
| Rounds | 1 | 1-3 (configurable) |
| Agent Types | 4 | 7 |
| Discussion Flow | Linear | Iterative with synthesis |
| Customization | None | Full team customization |
| Use Case | Quick ideation | Quick to comprehensive |

## Troubleshooting

### "Team Leader synthesis failed"
- This is usually due to API rate limits. The system will continue.
- Consider reducing MaxRounds or using a smaller team temporarily.

### Long execution times
- This is normal for Full configuration (5-10 minutes)
- Each agent makes API calls, and multi-round discussions take time
- Use Standard config if speed is critical

### High API costs
- Use Standard configuration for routine tasks
- Reserve Extended/Full for important decisions
- Monitor your Anthropic console for usage

## Migration from v1

v1 continues to work! You can use both:

```bash
# v1 - Original 4-agent team
./bin/cli

# v2 - Configurable teams
./bin/cli-v2
```

To migrate code:
```go
// v1
orch := orchestrator.NewOrchestrator(apiKey)

// v2 - with same behavior as v1
config := models.StandardTeamConfig()
orch := orchestrator.NewConfigurableOrchestrator(apiKey, config)
```

## Contributing

This is an open demonstration project. Feel free to:
- Add new agent types
- Create new configurations
- Enhance the orchestration logic
- Improve the UI

## License

MIT License - free to use and modify

## Acknowledgments

Built with:
- Go programming language
- Anthropic Claude API (Sonnet 4)
- Modern web standards

---

**Ready to explore ideas with your AI team?**

```bash
# Quick start
export ANTHROPIC_API_KEY="your-key"
# or
export ANTHROPIC_KEY="your-key"
./bin/cli-v2

# Or try the web interface
./bin/server-v2
# Open http://localhost:8080
```

**Need help?** Check USAGE.md for detailed examples and best practices.
