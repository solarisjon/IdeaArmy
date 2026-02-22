# Copilot Instructions for IdeaArmy (AI Agent Team)

## Build & Run

```bash
make build          # Build all binaries to bin/
make cli-tui        # Build just the TUI binary
make check          # Format + vet + build (run before committing)
make run-tui        # Build and run TUI (requires an LLM API key env var)
make run-cli        # Build and run CLI v2
make run-server     # Build and run web server (default port 8080)
```

There are no automated tests. Validation is manual — build all binaries and run with a real API key.

## Environment

The system supports multiple LLM backends. Auto-detection priority:

1. `LLM_BACKEND` — explicit backend: `"anthropic"` or `"openai"` (skips auto-detection)
2. `ANTHROPIC_API_KEY` or `ANTHROPIC_KEY` → uses Anthropic (Claude) backend
3. `LLMPROXY_KEY`, `OPENAI_API_KEY`, or `LLM_API_KEY` → uses OpenAI-compatible backend

Additional env vars:
- `LLM_API_KEY` — explicit API key (highest priority, any backend)
- `LLM_BASE_URL` — override API base URL (default: Anthropic API for anthropic, NetApp LLM proxy for openai)
- `LLM_MODEL` — override model name (default: `claude-sonnet-4-20250514` for anthropic, `gpt-4o` for openai)
- `LLMPROXY_KEY` — NetApp LLM proxy key in `user=xxx&key=sk_xxx` format (key portion auto-extracted)

## Architecture

This is a multi-agent orchestration system. The user provides a topic, and a configurable team of AI agents collaboratively ideates, validates, and selects the best idea, then generates an HTML report.

**Data flow:** User Input → Orchestrator → Agents (via `llm.Client` interface) → Discussion model → HTML Report

### Key components

- **`internal/llm/`** — `Client` interface and `Message` type shared by all backends. `resolve.go` handles env-var-based auto-detection. All agents and orchestrators program to this interface.
- **`internal/llmfactory/`** — Factory functions (`NewClient`, `NewClientAuto`) that create the appropriate backend client. Separated from `llm/` to avoid import cycles.
- **`internal/agents/`** — Each agent embeds `BaseAgent` and implements the `Agent` interface (`Process()` method). Agents are constructed with an `llm.Client` and have role-specific system prompts and temperatures.
- **`internal/orchestrator/`** — `ConfigurableOrchestrator` (v2) instantiates agents from a `TeamConfig`, stores them in a `map[AgentRole]Agent`, and drives a phased discussion: kickoff → ideation → validation → selection → report generation. Progress updates go through an `OnProgress` callback.
- **`internal/models/`** — Core data types (`Discussion`, `Idea`, `Message`, `AgentResponse`) and team presets (`StandardTeamConfig`, `ExtendedTeamConfig`, `FullTeamConfig`) in `config.go`. New agent roles must be added to `types.go` as `AgentRole` constants.
- **`internal/claude/`** — Anthropic Messages API client implementing `llm.Client`. `claude.Message` is a type alias for `llm.Message`.
- **`internal/openai/`** — OpenAI-compatible API client implementing `llm.Client`. Works with any OpenAI-compatible endpoint (OpenAI, Azure, LLM proxies).
- **`internal/tui/`** — Bubbletea-based terminal UI with Lipgloss styling.
- **`cmd/`** — Entry points: `cli/main.go` (v1), `cli/main_v2.go` (configurable), `cli/main_tui.go` (TUI), `server/main.go` (v1), `server/main_v2.go` (configurable).

### Adding a new agent

1. Create `internal/agents/<name>.go` — struct embedding `*BaseAgent`, constructor taking `llm.Client`, and `Process()` method
2. Add `Role<Name> AgentRole = "<name>"` constant in `internal/models/types.go`
3. Add `Include<Name> bool` field to `TeamConfig` in `internal/models/config.go` and update presets
4. Wire it into `ConfigurableOrchestrator` initialization and discussion flow in `internal/orchestrator/orchestrator_v2.go`

## Conventions

- **Commits:** Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, `chore:`)
- **Branches:** `feature/*`, `bugfix/*`, `docs/*` off `develop`
- **Formatting:** Always run `go fmt ./...` before committing
- **Error wrapping:** Use `fmt.Errorf("context: %w", err)`
- **Agent communication:** Append `models.Message` structs to `discussion.Messages` after each agent response
- **TUI updates:** Send typed messages (e.g., `AgentUpdateMsg`) via `p.Send()` to the Bubbletea program
