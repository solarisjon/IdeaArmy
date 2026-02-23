# Copilot Instructions for IdeaArmy

## Build & Validate

```bash
make check          # Format + vet + build (run before committing)
make build          # Build all 5 binaries to bin/
make cli-tui        # Build just the TUI binary
make run-tui        # Build and run TUI
make run-cli        # Build and run CLI v2
make run-server     # Build and run web server (default port 8080)
```

There are no automated tests. CI (`test.yml`) only runs `go build`, `go vet`, and `gofmt -s` checks. Validation is manual with a real LLM API key.

## Environment

LLM backend is auto-detected from env vars via `llm.ResolveBackend()` in `internal/llm/resolve.go`:

1. `LLM_BACKEND` — explicit: `"anthropic"` or `"openai"` (skips auto-detection)
2. `ANTHROPIC_API_KEY` or `ANTHROPIC_KEY` → Anthropic backend
3. `LLMPROXY_KEY`, `OPENAI_API_KEY`, or `LLM_API_KEY` → OpenAI-compatible backend

Overrides: `LLM_API_KEY` (key), `LLM_BASE_URL` (endpoint), `LLM_MODEL` (model name).

## Architecture

Multi-agent orchestration system: user provides a topic → configurable team of AI agents ideates, validates, and selects the best idea → generates an HTML "idea sheet" report.

**Data flow:** User Input → `ConfigurableOrchestrator` → Agents (via `llm.Client`) → `Discussion` model → HTML Report

### Core abstraction: `llm.Client` interface

All LLM interaction goes through `llm.Client` (`internal/llm/interface.go`): `SendMessage()`, `SendMessageWithTokens()`, `SimpleQuery()`. Two implementations exist (`internal/claude/`, `internal/openai/`). `internal/llmfactory/` creates the right client — it's separated from `llm/` to avoid import cycles. Never import a backend package directly from agents or orchestrators.

### Agent pattern

Every agent in `internal/agents/` follows the same structure:
- Struct embeds `*BaseAgent` (which holds `llm.Client`, system prompt, temperature)
- Constructor takes `llm.Client` as sole parameter (not a concrete backend type)
- Implements `Process(context *models.Discussion, input string) (*models.AgentResponse, error)`
- Uses `a.Query()` (inherited from `BaseAgent`) or `a.QueryWithTokens()` for LLM calls
- Agents that generate ideas parse JSON from the LLM response and return `[]models.Idea` in the `AgentResponse`

Current agents: `team_leader`, `ideation`, `moderator`, `researcher`, `critic`, `implementer`, `ui_creator`.

### Orchestrator phases

`ConfigurableOrchestrator` (`internal/orchestrator/orchestrator_v2.go`) drives discussion in phases:
1. **Kickoff** — Team leader sets direction
2. **Exploration rounds** (1–3 per `TeamConfig.MaxRounds`) — agents contribute sequentially: researcher → ideation → critic → implementer. Each appends to `discussion.Messages`.
3. **Final validation** — Moderator scores ideas
4. **Selection** — Leader picks best idea (falls back to highest score)
5. **Visualization** — `UICreatorAgent.GenerateIdeaSheet()` produces HTML (type-asserted from the agent map)

Progress updates go through `OnProgress` callback. Agent failures in exploration rounds are logged but non-fatal (the discussion continues).

### Team configuration

`internal/models/config.go` defines `TeamConfig` with `Include<Role> bool` fields and presets: `StandardTeamConfig` (4 agents), `ExtendedTeamConfig` (6), `FullTeamConfig` (7). The orchestrator reads these flags to decide which agents to instantiate.

### Adding a new agent

1. Create `internal/agents/<name>.go` — struct embedding `*BaseAgent`, constructor taking `llm.Client`, `Process()` method
2. Add `Role<Name> AgentRole = "<name>"` constant in `internal/models/types.go`
3. Add `Include<Name> bool` field to `TeamConfig` in `internal/models/config.go` and update the preset functions
4. Wire into `ConfigurableOrchestrator` constructor and `runExplorationRound()` in `internal/orchestrator/orchestrator_v2.go`

### TUI

Built with Bubbletea (`internal/tui/`), styled with Lipgloss. Send typed messages (e.g., `AgentUpdateMsg`) via `p.Send()` to update the UI.

## Conventions

- **Commits:** Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, `chore:`)
- **Branches:** `feature/*`, `bugfix/*`, `docs/*` off `develop`
- **Formatting:** `go fmt ./...` before every commit (CI enforces `gofmt -s`)
- **Error wrapping:** `fmt.Errorf("context: %w", err)`
- **Agent communication:** Append `models.Message` structs to `discussion.Messages` after each agent response
- **LLM client construction:** Always use `llmfactory.NewClientAuto("")` or `llmfactory.NewClient(cfg)` — never instantiate `claude.Client` or `openai.Client` directly from cmd/ or orchestrator code
