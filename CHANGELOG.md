# Changelog

## [Unreleased]

### Added — Multi-Backend LLM Support & War Room TUI

#### Multi-Backend LLM Support (Anthropic, OpenAI, NetApp LLM Proxy)
- **New package: `internal/llm/`** — Backend-agnostic `LLMClient` interface and shared types
- **New package: `internal/openai/`** — OpenAI-compatible API client with 120s HTTP timeout
- **New package: `internal/llmfactory/`** — Auto-detects backend from environment and instantiates the correct client
- Environment variables for backend selection:
  - `ANTHROPIC_API_KEY` / `ANTHROPIC_KEY` → Anthropic Claude
  - `LLMPROXY_KEY` → NetApp LLM Proxy (format: `user=username&key=sk_xxx`)
  - `OPENAI_API_KEY` / `LLM_API_KEY` → OpenAI-compatible
  - `LLM_BACKEND` forces backend, `LLM_MODEL` overrides model, `LLM_BASE_URL` overrides endpoint

#### War Room TUI Redesign
- **Persona-named agents:** Captain Rex (Leader), Sparky (Ideation), The Judge (Moderator), Doc Sage (Researcher), Nitpick (Critic), Wrench (Implementer), Pixel (UI Creator)
- **Speech bubbles** showing each agent's contributions in the terminal
- **Auto-exit** after discussion completes (no more pressing 'q' to quit)

#### Other Improvements
- Server HTML updated — API key field is now optional
- HTTP timeout (120s) added to OpenAI client for long-running completions
- Visualization errors are now non-fatal (discussion continues on failure)

### Added - February 21, 2025

#### Comprehensive Strategic Reports
- **ENHANCED: Report Generator** (formerly UI Creator)
- Now generates detailed 8-section reports (not single-page summaries)
- Includes top 3-4 runner-up ideas with full analysis
- Explains specifically why each alternative wasn't selected
- When each alternative might be the better choice
- Open questions and recommendations section
- Discussion journey across all rounds
- Comparative analysis of top ideas
- Implementation considerations and next steps
- Uses 8,192 tokens for comprehensive output
- 3-5 screens of detailed, executive-level content

**Report Structure:**
1. Executive Summary
2. Recommended Solution (with selection reasoning)
3. Runner-Up Ideas (detailed analysis of alternatives)
4. All Ideas Explored
5. Discussion Journey
6. Comparative Analysis
7. Open Questions & Next Steps
8. Recommendations & Considerations

- **New documentation:** `REPORTS.md`

#### Beautiful Terminal UI (TUI)
- **NEW: Dynamic visual interface** with Charm Bracelet libraries
- Real-time agent spinners showing active work
- Live progress bars for phases and rounds
- Color-coded agents with unique icons (🎯💡🔍📚🤔🔧🎨)
- Ideas appear as they're generated
- Recent activity log with smooth updates
- Live statistics and timer
- Professional styling with gradients and animations
- Built with Bubbletea, Lipgloss, and Bubbles

- **New files:**
  - `internal/tui/model.go` - Bubbletea model
  - `internal/tui/styles.go` - Lipgloss styling
  - `internal/tui/runner.go` - TUI orchestration
  - `cmd/cli/main_tui.go` - TUI CLI application
  - `TUI_GUIDE.md` - Complete TUI documentation

- **New binary:**
  - `bin/cli-tui` (9.7MB)

#### Support for ANTHROPIC_KEY environment variable
- The system now accepts both `ANTHROPIC_API_KEY` and `ANTHROPIC_KEY` environment variables
- `ANTHROPIC_API_KEY` is checked first, then falls back to `ANTHROPIC_KEY`
- Updated in:
  - Claude API client (`internal/claude/client.go`)
  - CLI v1 (`cmd/cli/main.go`)
  - CLI v2 (`cmd/cli/main_v2.go`)
  - Example files (`example.go`, `example_v2.go`)
  - All documentation

#### v2 - Configurable Multi-Agent System
- **New specialized agents:**
  - Researcher - Provides factual research and real-world context
  - Critic - Challenges assumptions and identifies risks
  - Implementer - Focuses on practical execution planning

- **Configurable team compositions:**
  - Standard: 4 agents, 1 round (quick, focused)
  - Extended: 6 agents, 2 rounds (deeper analysis) - Recommended
  - Full: 7 agents, 3 rounds (maximum depth)
  - Custom: Build your own team configuration

- **Multi-round discussions:**
  - Iterative refinement across multiple rounds
  - Leader synthesis after each round
  - Agents build on each other's contributions
  - Better context building and idea evolution

- **New files added:**
  - `internal/agents/researcher.go`
  - `internal/agents/critic.go`
  - `internal/agents/implementer.go`
  - `internal/models/config.go`
  - `internal/orchestrator/orchestrator_v2.go`
  - `cmd/cli/main_v2.go`
  - `cmd/server/main_v2.go`
  - `example_v2.go`
  - `README_V2.md`
  - `QUICKSTART.md`

- **New binaries:**
  - `bin/cli-v2`
  - `bin/server-v2`

### Changed
- Updated all documentation to mention both environment variable options
- Error messages now suggest both `ANTHROPIC_API_KEY` and `ANTHROPIC_KEY`

## [1.0.0] - Initial Release

### Added
- Team of 4 AI agents (Team Leader, Ideation, Moderator, UI Creator)
- Single-round discussion flow
- CLI interface
- Web server interface
- Beautiful HTML idea sheet generation
- Anthropic Claude API integration
- Complete documentation

### Features
- Collaborative ideation with specialized agents
- Idea validation with 0-10 scoring
- Pros/cons analysis
- Visual HTML output
- Both CLI and web interfaces
