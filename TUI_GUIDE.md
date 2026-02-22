## 🎨 Terminal UI (TUI) Guide

The AI Agent Team now includes a beautiful, dynamic Terminal User Interface built with [Charm Bracelet](https://github.com/charmbracelet) libraries!

### Features

**Real-time Agent Visualization — "War Room" Theme**
- 🎖️ Persona-named agents with distinct personalities
- 💬 Speech bubbles showing each agent's contributions
- 🔄 Live spinners showing which agents are actively working
- ✅ Checkmarks when agents complete their tasks
- 🎨 Color-coded agents with unique icons

**Dynamic Progress Tracking**
- Progress bar showing overall completion
- Phase indicators (Kickoff, Exploration, Validation, etc.)
- Round tracking for multi-round discussions
- Live timer showing elapsed time

**Live Updates**
- Ideas appear as they're generated
- Recent activity log with latest agent actions
- Real-time statistics (ideas count, messages count)
- Smooth animations and transitions

**Beautiful Styling**
- Gradient progress bars
- Color-coded agent roles
- Professional typography with lipgloss
- Responsive layout

### Quick Start

```bash
# Set at least one LLM API key:
export ANTHROPIC_API_KEY="your-key"       # Anthropic Claude
# or: export OPENAI_API_KEY="your-key"    # OpenAI-compatible
# or: export LLMPROXY_KEY="user=me&key=sk_xxx"  # NetApp LLM Proxy

./bin/cli-tui
```

### What You'll See

```
╔════════════════════════════════════════════════════════╗
║            🎖️  WAR ROOM  — AI Agent Team              ║
╚════════════════════════════════════════════════════════╝

🤖 AI Agent Team - Collaborative Ideation
Topic: Your topic here

Team Composition:
🎖️ Captain Rex (Leader) • ⚡ Sparky (Ideation) • ⚖️ The Judge (Moderator) • 📚 Doc Sage (Researcher) • 🧐 Nitpick (Critic) • 🔧 Wrench (Implementer) • 🎨 Pixel (UI Creator)

Phase: Exploration & Ideation  Round 1/2
████████████████████░░░░ 75%

Agent Status:
  🎖️ Captain Rex: ✓ Complete
  ⚡ Sparky: ⠋ Generating creative ideas...
  ⚖️ The Judge: Ready
  📚 Doc Sage: ⠙ Researching context...
  🧐 Nitpick: Ready
  🔧 Wrench: Ready
  🎨 Pixel: Ready

💬 Speech Bubbles:
  ┌─ Captain Rex ──────────────────────────────┐
  │ "Let's focus on sustainability angles..."  │
  └────────────────────────────────────────────┘
  ┌─ Sparky ───────────────────────────────────┐
  │ "What if we combine vertical farming with  │
  │  AI-driven crop rotation?"                 │
  └────────────────────────────────────────────┘

💡 Ideas Generated (3):
  • Smart vertical farming system [8.5/10]
  • Community-integrated food hubs [7.2/10]
  • AI-optimized crop rotation [9.1/10]

Recent Activity:
  → Doc Sage providing market analysis
  → Sparky generating new concepts
  → Captain Rex synthesizing Round 1

⚡ Running... (2m 34s)
  Ideas: 3 | Messages: 24
```

> **Note:** The TUI automatically exits when the discussion completes — no need to press 'q'.

### Agent Status Indicators

**Spinner (⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏)** - Agent is actively working
**Checkmark (✓)** - Agent has completed its task
**Ready** - Agent is idle, waiting for its turn

### Color Scheme

Each persona agent has a unique color for easy identification:

- 🎖️ **Captain Rex** (Leader) - Gold
- ⚡ **Sparky** (Ideation) - Green
- ⚖️ **The Judge** (Moderator) - Blue
- 📚 **Doc Sage** (Researcher) - Purple
- 🧐 **Nitpick** (Critic) - Orange
- 🔧 **Wrench** (Implementer) - Cyan
- 🎨 **Pixel** (UI Creator) - Pink

### Controls

- **Ctrl+C** - Force-quit the application (the TUI auto-exits on completion)

### Phases

The TUI shows you exactly where the discussion is:

1. **Team Leader Kickoff** - Initial direction setting
2. **Exploration & Ideation** - Generating and refining ideas
3. **Leader Synthesis** - Leader summarizes each round
4. **Validation & Scoring** - Moderator evaluates ideas
5. **Final Selection** - Team Leader chooses best idea
6. **Creating Idea Sheet** - UI Creator generates visualization

### Team Configurations

The TUI supports all team configurations:

**Standard (4 agents)**
- Fast execution
- Clear visualization
- Perfect for quick ideation

**Extended (6 agents)** - Recommended
- More dynamic with Researcher and Critic
- Multi-round refinement visible in real-time
- Best balance of depth and watchability

**Full (7 agents)**
- Maximum agent activity
- Most dynamic visualization
- Watch all 7 agents collaborate

### Tips for Best Experience

**Terminal Size**
- Minimum: 80 columns x 24 rows
- Recommended: 100+ columns x 30+ rows
- Use full screen for best experience

**Color Support**
- Works best with true color terminals
- iTerm2, Alacritty, or modern terminal.app on macOS
- Windows Terminal on Windows
- Most modern Linux terminals

**Font**
- Works with any monospace font
- Nerd Fonts add extra visual appeal
- Standard Unicode support required

### Comparison: TUI vs Standard CLI

| Feature | Standard CLI | TUI |
|---------|-------------|-----|
| Visual feedback | Text lines | Dynamic spinners & progress |
| Agent status | Text messages | Real-time colored indicators |
| Progress | Percentage text | Visual progress bar |
| Ideas | Listed at end | Appear live as generated |
| Experience | Functional | Beautiful & engaging |

### After TUI Completes

Once the TUI finishes, you'll see:
1. Full discussion summary
2. Final selected idea with pros/cons
3. Saved HTML idea sheet path
4. Complete statistics

### Troubleshooting

**TUI not displaying correctly**
- Check terminal size (at least 80x24)
- Try a different terminal emulator
- Ensure UTF-8 support is enabled

**Colors not showing**
- Enable true color support in your terminal
- Check TERM environment variable
- Try: `export TERM=xterm-256color`

**Spinners not animating**
- Some terminals may not support all Unicode characters
- The discussion will still work, just less animated

**TUI freezes**
- The agents are working! It can take several minutes
- Watch for the timer and progress bar updates
- Press Ctrl+C to force-quit if needed (progress will be lost)

### Why Use the TUI?

**Visual Engagement**
- More engaging than scrolling text
- See exactly which agent is working
- Feel the collaboration happening

**Better Understanding**
- Visual phases make the process clear
- Progress bar shows how much is left
- Round tracking for multi-round discussions

**Professional Presentation**
- Impressive for demos
- Great for presentations
- Shows the sophistication of the system

### Integration

Want to use TUI in your own code?

```go
import "github.com/yourusername/ai-agent-team/internal/tui"

config := models.ExtendedTeamConfig()
discussion, err := tui.Run(apiKey, config, topic)
```

### Technical Details

Built with:
- **Bubbletea** - The Elm-inspired TUI framework
- **Lipgloss** - Style definitions and layout
- **Bubbles** - Progress bars and spinners

The TUI runs the orchestration in a goroutine and receives updates via Bubbletea messages, creating a reactive, real-time interface.

---

**Try it now!**

```bash
./bin/cli-tui
```

Watch your AI agents collaborate in beautiful, real-time terminal graphics!
