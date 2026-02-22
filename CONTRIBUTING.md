# Contributing to AI Agent Team

Thank you for your interest in contributing to AI Agent Team!

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- An LLM API key for testing (Anthropic, OpenAI, or NetApp LLM Proxy)

### Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/ai-agent-team.git
cd ai-agent-team
```

2. Install dependencies:
```bash
go mod download
```

3. Build the project:
```bash
go build -o bin/cli ./cmd/cli/main.go
go build -o bin/cli-v2 ./cmd/cli/main_v2.go
go build -o bin/cli-tui ./cmd/cli/main_tui.go
go build -o bin/server ./cmd/server/main.go
go build -o bin/server-v2 ./cmd/server/main_v2.go
```

## Development Workflow

### Branch Strategy

- `main` - Production-ready code
- `develop` - Integration branch for features
- `feature/*` - New features
- `bugfix/*` - Bug fixes
- `docs/*` - Documentation updates

### Making Changes

1. Create a new branch:
```bash
git checkout -b feature/your-feature-name
```

2. Make your changes

3. Test your changes:
```bash
# Build all binaries
make build  # or manually build each binary

# Test with an LLM API key (set at least one)
export ANTHROPIC_API_KEY="your-key"       # Anthropic Claude
# or: export OPENAI_API_KEY="your-key"    # OpenAI-compatible
# or: export LLMPROXY_KEY="user=me&key=sk_xxx"  # NetApp LLM Proxy
./bin/cli-tui
```

4. Format your code:
```bash
go fmt ./...
```

5. Commit your changes:
```bash
git add .
git commit -m "feat: add your feature description"
```

## Commit Message Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Adding tests
- `chore:` - Maintenance tasks

Examples:
```
feat: add new researcher agent for fact-checking
fix: correct progress bar calculation in TUI
docs: update REPORTS.md with new sections
refactor: simplify orchestrator message handling
```

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Keep functions focused and small
- Add comments for exported functions
- Use meaningful variable names

## Adding New Agents

To add a new agent:

1. Create a new file in `internal/agents/`:
```go
package agents

type YourAgent struct {
    *BaseAgent
}

func NewYourAgent(client *claude.Client) *YourAgent {
    systemPrompt := `Your agent's system prompt here...`

    return &YourAgent{
        BaseAgent: &BaseAgent{
            Role:         "your_role",
            Name:         "Your Agent Name",
            SystemPrompt: systemPrompt,
            Client:       client,
            Temperature:  0.7,
        },
    }
}

func (a *YourAgent) Process(context *models.Discussion, input string) (*models.AgentResponse, error) {
    // Implementation
}
```

2. Add the role to `internal/models/types.go`:
```go
const (
    // ... existing roles
    RoleYourAgent AgentRole = "your_role"
)
```

3. Update `internal/models/config.go` to include your agent in configs

4. Update documentation

## Testing

### Manual Testing

Test with various configurations:

```bash
# Standard team
./bin/cli-v2  # Select option 1

# Extended team
./bin/cli-v2  # Select option 2

# Full team
./bin/cli-v2  # Select option 3

# TUI
./bin/cli-tui
```

### Test Topics

Use diverse topics to test:
- Simple: "Mobile app features for habit tracking"
- Complex: "Strategy for entering sustainable packaging market"
- Technical: "Microservices architecture for e-commerce platform"

## Documentation

When adding features:

1. Update relevant `.md` files
2. Add inline code comments
3. Update CHANGELOG.md
4. If changing behavior, update README_V2.md

## Pull Request Process

1. Update documentation
2. Ensure code is formatted (`go fmt ./...`)
3. Test all configurations work
4. Update CHANGELOG.md
5. Create pull request with clear description
6. Reference any related issues

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] New feature
- [ ] Bug fix
- [ ] Documentation update
- [ ] Refactoring

## Testing
How you tested the changes

## Checklist
- [ ] Code formatted with `go fmt`
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] All binaries build successfully
- [ ] Tested with real API
```

## Project Structure

```
ai-agent-team/
├── cmd/                    # Entry points
│   ├── cli/               # CLI applications
│   └── server/            # Server applications
├── internal/              # Internal packages
│   ├── agents/           # Agent implementations
│   ├── orchestrator/     # Discussion orchestration
│   ├── models/           # Data models
│   ├── llm/              # Backend-agnostic LLM interface
│   ├── claude/           # Anthropic Claude client
│   ├── openai/           # OpenAI-compatible client
│   ├── llmfactory/       # Auto-detect & create LLM client
│   └── tui/              # Terminal UI
├── .github/              # GitHub workflows
├── bin/                  # Built binaries (gitignored)
└── docs/                 # Documentation
```

## Adding Dependencies

When adding Go dependencies:

```bash
go get github.com/package/name
go mod tidy
```

Then commit both `go.mod` and `go.sum`.

## Questions?

- Open an issue for bugs or feature requests
- Start a discussion for questions
- Check existing documentation in `/docs`

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Code of Conduct

Be respectful, inclusive, and constructive. We're building something cool together!
