# Development Guide

## Quick Start for Developers

### Prerequisites
- Go 1.21+
- Git
- An LLM API key (Anthropic, OpenAI, or NetApp LLM Proxy)

### Clone and Build
```bash
git clone https://github.com/yourusername/ai-agent-team.git
cd ai-agent-team
make deps
make build
```

### Run Tests
```bash
# Check formatting and vet
make check

# Run specific binary (set at least one LLM key)
export ANTHROPIC_API_KEY="your-key"       # Anthropic Claude
# or: export OPENAI_API_KEY="your-key"    # OpenAI-compatible
# or: export LLMPROXY_KEY="user=me&key=sk_xxx"  # NetApp LLM Proxy
make run-tui
```

## Project Architecture

### High-Level Overview

```
User Input → Orchestrator → Agents → LLM Backend (Anthropic / OpenAI / LLM Proxy)
                ↓              ↓
            Discussion ←  Messages → Report
```

### Core Components

**1. Agents (`internal/agents/`)**
- `BaseAgent` - Common functionality
- Specialized agents with unique system prompts
- Process() method for handling tasks
- Query() method for Claude API calls

**2. Orchestrator (`internal/orchestrator/`)**
- Manages multi-round discussions
- Routes messages between agents
- Tracks discussion state
- Coordinates phases (kickoff, ideation, validation, etc.)

**3. Models (`internal/models/`)**
- Data structures (Discussion, Idea, Message, etc.)
- Team configurations (Standard, Extended, Full)
- Agent roles and types

**4. LLM Interface (`internal/llm/`)**
- Backend-agnostic `LLMClient` interface
- Common request/response types shared by all backends

**5. Claude Client (`internal/claude/`)**
- Anthropic Claude API implementation of `LLMClient`
- Message handling and token management

**6. OpenAI Client (`internal/openai/`)**
- OpenAI-compatible API implementation of `LLMClient`
- Also used for NetApp LLM Proxy via `LLM_BASE_URL`
- HTTP timeout of 120s for long-running completions

**7. LLM Factory (`internal/llmfactory/`)**
- Auto-detects backend from environment variables
- Instantiates the correct `LLMClient` implementation
- Supports `LLM_BACKEND` override and `LLM_MODEL` / `LLM_BASE_URL` overrides

**8. TUI (`internal/tui/`)**
- Bubbletea model for reactive UI
- Lipgloss styling
- Real-time updates via messages

### Data Flow

#### Discussion Lifecycle

1. **Initialization**
   ```go
   config := models.ExtendedTeamConfig()
   orch := orchestrator.NewConfigurableOrchestrator(apiKey, config)
   ```

2. **Kickoff**
   - Team Leader sets direction
   - Context established

3. **Multi-Round Exploration**
   - Each round:
     - Researcher (if present) provides context
     - Ideation generates/refines ideas
     - Critic (if present) challenges assumptions
     - Implementer (if present) plans execution
     - Leader synthesizes round

4. **Validation**
   - Moderator scores all ideas
   - Pros/cons identified

5. **Selection**
   - Leader chooses best idea
   - Auto-select highest score as fallback

6. **Report Generation**
   - Report Generator creates comprehensive HTML
   - 8 sections with runner-up analysis

### Adding New Features

#### New Agent Type

1. Create agent file:
```go
// internal/agents/analyzer.go
package agents

type AnalyzerAgent struct {
    *BaseAgent
}

func NewAnalyzerAgent(client *claude.Client) *AnalyzerAgent {
    return &AnalyzerAgent{
        BaseAgent: &BaseAgent{
            Role:         "analyzer",
            Name:         "Data Analyzer",
            SystemPrompt: "Your detailed prompt...",
            Client:       client,
            Temperature:  0.5,
        },
    }
}

func (a *AnalyzerAgent) Process(ctx *models.Discussion, input string) (*models.AgentResponse, error) {
    // Implementation
}
```

2. Add to models:
```go
// internal/models/types.go
const RoleAnalyzer AgentRole = "analyzer"
```

3. Add to config:
```go
// internal/models/config.go
type TeamConfig struct {
    // ...
    IncludeAnalyzer bool
}
```

4. Update orchestrator to call it

5. Add to TUI styles with icon and color

#### New Phase in Orchestration

```go
// internal/orchestrator/orchestrator_v2.go
func (o *ConfigurableOrchestrator) runYourNewPhase() error {
    o.notify("🔬 New Phase Starting")

    // Your logic here

    return nil
}
```

Add it to the main discussion flow.

#### New Report Section

Update the Report Generator system prompt in `internal/agents/ui_creator.go` to include your new section.

## Code Patterns

### Error Handling
```go
if err := someFunction(); err != nil {
    return fmt.Errorf("descriptive context: %w", err)
}
```

### Agent Communication
```go
response, err := agent.Process(discussion, "Task description")
if err != nil {
    return err
}

// Add to discussion
discussion.Messages = append(discussion.Messages, models.Message{
    From:    agent.GetRole(),
    To:      "team",
    Content: response.Content,
    Type:    "response",
})
```

### TUI Updates
```go
// Send message to TUI
p.Send(AgentUpdateMsg{
    Role:    "agent_role",
    Status:  "working",
    Message: "Doing something...",
})
```

## Configuration Management

### Team Configs

Create preset configs in `internal/models/config.go`:

```go
func MyCustomTeamConfig() *TeamConfig {
    return &TeamConfig{
        IncludeTeamLeader:  true,
        IncludeIdeation:    true,
        IncludeModerator:   true,
        IncludeYourAgent:   true,
        MaxRounds:          2,
        DeepDive:           true,
        MinScoreThreshold:  7.0,
    }
}
```

### LLM API Settings

The LLM backend is selected automatically based on which environment variable is set:

| Variable | Backend |
|----------|---------|
| `ANTHROPIC_API_KEY` / `ANTHROPIC_KEY` | Anthropic Claude |
| `LLMPROXY_KEY` | NetApp LLM Proxy (format: `user=username&key=sk_xxx`) |
| `OPENAI_API_KEY` / `LLM_API_KEY` | OpenAI-compatible |

Override controls:
- `LLM_BACKEND` — force a specific backend (`anthropic`, `openai`, `llmproxy`)
- `LLM_MODEL` — override the default model name
- `LLM_BASE_URL` — override the API endpoint URL

Default model settings are in the respective client packages (`internal/claude/`, `internal/openai/`).

## Testing Strategy

### Manual Testing Checklist

- [ ] Standard team (4 agents, 1 round)
- [ ] Extended team (6 agents, 2 rounds)
- [ ] Full team (7 agents, 3 rounds)
- [ ] Custom team configuration
- [ ] TUI interface
- [ ] Web interface
- [ ] Report quality (all 8 sections)
- [ ] Runner-up analysis present
- [ ] Different topic types (simple, complex, technical)

### Test Topics

**Simple:**
- "Mobile app features for habit tracking"

**Medium:**
- "Employee retention program for remote teams"

**Complex:**
- "Complete go-to-market strategy for new SaaS product"

**Technical:**
- "Microservices vs monolithic architecture for scaling"

## Performance Considerations

### Token Usage
- Standard discussion: ~10K-20K tokens
- Extended discussion: ~25K-40K tokens
- Full discussion: ~40K-60K tokens
- Report generation: Uses 8K tokens (2x normal)

### Timing
- Standard: 1-2 minutes
- Extended: 3-5 minutes
- Full: 5-10 minutes

### Optimization Tips
1. Use lower temperature for faster, more deterministic responses
2. Limit rounds for quicker results
3. Smaller teams = faster execution
4. Concurrent agent calls where possible

## Debugging

### Common Issues

**1. Agent not responding**
- Check system prompt
- Verify API key is valid for the configured backend
- Check token limits
- Verify `LLM_BACKEND` / `LLM_BASE_URL` if using non-default backend

**2. TUI not updating**
- Ensure messages are being sent
- Check spinner.Tick is being called
- Verify Update() handles all message types

**3. Poor report quality**
- Increase max tokens for Report Generator
- Ensure all ideas have scores
- Check discussion context is complete

### Debug Logging

Add temporary logging:
```go
log.Printf("DEBUG: Agent %s received: %s", agent.Role, input)
```

## Building and Releasing

### Build All Binaries
```bash
make build
```

### Build Specific Binary
```bash
make cli-tui
make server-v2
```

### Clean Build
```bash
make clean
make build
```

### Check Before Commit
```bash
make check  # Runs fmt + vet + build
```

## Directory Structure Deep Dive

```
ai-agent-team/
├── .github/
│   ├── workflows/          # CI/CD
│   └── ISSUE_TEMPLATE/     # GitHub templates
├── cmd/
│   ├── cli/
│   │   ├── main.go        # Original CLI
│   │   ├── main_v2.go     # Configurable CLI
│   │   └── main_tui.go    # TUI version
│   └── server/
│       ├── main.go        # Original server
│       └── main_v2.go     # Configurable server
├── internal/
│   ├── agents/            # All agent implementations
│   ├── orchestrator/      # Discussion coordination
│   ├── models/           # Data structures
│   ├── llm/              # Backend-agnostic LLM interface
│   ├── claude/           # Anthropic Claude client
│   ├── openai/           # OpenAI-compatible client
│   ├── llmfactory/       # Auto-detect & create LLM client
│   └── tui/              # Terminal UI
├── bin/                  # Built binaries (gitignored)
├── Makefile             # Build automation
├── go.mod & go.sum      # Dependencies
├── *.md                 # Documentation
└── example*.go          # Usage examples
```

## Git Workflow

### Branch Naming
- `feature/agent-memory` - New features
- `bugfix/tui-spinner-crash` - Bug fixes
- `docs/update-readme` - Documentation

### Commit Messages
```
feat: add memory persistence for agents
fix: correct progress calculation in round 2
docs: update REPORTS.md with examples
refactor: simplify orchestrator message routing
```

### Before Pushing
```bash
make check           # Format, vet, build
git add .
git commit -m "..."
git push origin feature/your-branch
```

## Resources

- [Go Documentation](https://go.dev/doc/)
- [Bubbletea Tutorial](https://github.com/charmbracelet/bubbletea)
- [Anthropic API Docs](https://docs.anthropic.com/)
- [OpenAI API Docs](https://platform.openai.com/docs/)
- [Conventional Commits](https://www.conventionalcommits.org/)

## Questions?

- Check CONTRIBUTING.md
- Open an issue
- Start a discussion

Happy coding! 🚀
