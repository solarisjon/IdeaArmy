# AI Agent Team - Collaborative Ideation System

A sophisticated multi-agent AI system built in Go that brings together specialized AI agents to collaboratively explore complex ideas, validate concepts, and produce beautiful visualizations of the ideation process.

## Overview

This system orchestrates a team of four specialized AI agents, each with unique capabilities:

### The Team

1. **Team Leader Agent** 🎯
   - Manages the overall discussion flow
   - Coordinates between team members
   - Makes final decisions on idea selection
   - Ensures comprehensive exploration of topics

2. **Ideation Agent** 💡
   - Generates creative, well-researched ideas
   - Explores concepts from multiple angles
   - Provides detailed reasoning and examples
   - Thinks deeply about feasibility and innovation

3. **Moderator/Facilitator Agent** 🔍
   - Validates idea quality and feasibility
   - Provides objective scoring (0-10 scale)
   - Identifies pros, cons, and potential issues
   - Ensures rigorous evaluation of all concepts

4. **UI Creator Agent** 🎨
   - Creates beautiful HTML visualizations
   - Designs comprehensive idea sheets
   - Presents discussion journey and outcomes
   - Makes complex information scannable and attractive

## Features

- **Deep Ideation**: Agents research and explore topics thoroughly, drawing from various domains
- **Collaborative Discussion**: Agents build on each other's contributions
- **Rigorous Validation**: All ideas are critically evaluated with scores and feedback
- **Beautiful Visualizations**: Final ideas are presented in professional HTML format
- **Dual Interface**: Use via CLI for quick sessions or web UI for richer experience
- **Real-time Progress**: Track the discussion as it unfolds

## Prerequisites

- Go 1.21 or higher
- Anthropic API key (get one at https://console.anthropic.com/)

## Installation

1. Clone or download this repository:
```bash
cd ai-agent-team
```

2. Install dependencies:
```bash
go mod download
```

3. Set your Anthropic API key:
```bash
export ANTHROPIC_API_KEY="your-api-key-here"
# or
export ANTHROPIC_KEY="your-api-key-here"
```

## Usage

### CLI Mode

Run quick ideation sessions from the command line:

```bash
go run cmd/cli/main.go
```

You'll be prompted to:
1. Enter your API key (if not set in environment)
2. Provide a topic for discussion

Example topics:
- "Innovative ways to reduce food waste in urban areas"
- "AI-powered tools for improving remote team collaboration"
- "Sustainable transportation solutions for small cities"

The CLI will:
- Show real-time progress as agents work
- Display the final selected idea with pros/cons
- Save a beautiful HTML idea sheet to your current directory

### Web Server Mode

Launch the web interface for a richer experience:

```bash
go run cmd/server/main.go
```

Then open your browser to `http://localhost:8080`

The web interface provides:
- Easy form-based input
- Real-time progress visualization
- In-browser viewing of the idea sheet
- Clean, modern UI

## How It Works

### Discussion Flow

The orchestrator guides the team through five phases:

1. **Kickoff** - Team Leader introduces the topic and sets direction
2. **Ideation** - Ideation Agent generates 3-5 creative, well-researched ideas
3. **Validation** - Moderator evaluates ideas, assigns scores, identifies pros/cons
4. **Selection** - Team Leader chooses the best idea based on evaluations
5. **Visualization** - UI Creator generates a beautiful HTML idea sheet

### Agent Communication

Agents communicate through a structured message system:
- Each phase is clearly delineated
- Agents build context from previous discussion
- All ideas and evaluations are tracked
- Final output includes the complete discussion journey

## Project Structure

```
ai-agent-team/
├── cmd/
│   ├── cli/          # CLI application
│   └── server/       # Web server
├── internal/
│   ├── agents/       # Agent implementations
│   │   ├── agent.go          # Base agent interface
│   │   ├── team_leader.go    # Team Leader agent
│   │   ├── ideation.go       # Ideation agent
│   │   ├── moderator.go      # Moderator agent
│   │   └── ui_creator.go     # UI Creator agent
│   ├── orchestrator/ # Agent coordination
│   ├── models/       # Data structures
│   └── claude/       # Claude API client
├── web/
│   ├── templates/    # HTML templates
│   └── static/       # CSS, JS
├── go.mod
├── go.sum
└── README.md
```

## Configuration

### Environment Variables

- `ANTHROPIC_API_KEY` or `ANTHROPIC_KEY` - Your Anthropic API key (required)
- `PORT` - Server port (default: 8080, web mode only)

### Model Configuration

By default, the system uses `claude-sonnet-4-20250514`. You can change this in `internal/claude/client.go`:

```go
const DefaultModel = "your-preferred-model"
```

### Agent Personalities

Each agent has a customizable system prompt. To modify agent behavior, edit the `systemPrompt` in their respective files:
- `internal/agents/team_leader.go`
- `internal/agents/ideation.go`
- `internal/agents/moderator.go`
- `internal/agents/ui_creator.go`

## API Endpoints (Web Mode)

- `GET /` - Web interface
- `POST /api/start` - Start a new discussion
  ```json
  {
    "api_key": "your-key",
    "topic": "your topic"
  }
  ```
- `GET /api/status/:id` - Get discussion status
- `GET /api/result/:id` - Get discussion result with HTML

## Examples

### Example Output

When discussing "Innovative solutions for urban vertical farming", the system might:

1. Generate ideas like:
   - AI-optimized growing systems
   - Modular stackable growing units
   - Community-integrated vertical farms

2. Evaluate each with:
   - Feasibility scores
   - Implementation complexity
   - Market potential
   - Environmental impact

3. Select the best idea with detailed justification

4. Create an HTML sheet showcasing:
   - All ideas considered
   - Evaluation scores and reasoning
   - Final idea with comprehensive pros/cons
   - Discussion highlights

## Troubleshooting

### "API error" messages
- Verify your API key is correct
- Check your Anthropic account has available credits
- Ensure you have internet connectivity

### "No ideas generated"
- Try making your topic more specific
- Ensure the topic is appropriate for ideation
- Check API rate limits

### Web server won't start
- Check if port 8080 is already in use
- Try setting a different port: `PORT=3000 go run cmd/server/main.go`

## Future Enhancements

Potential additions:
- Persistent storage for discussions
- Export to PDF/Markdown
- Multiple discussion rounds
- Agent memory across sessions
- Custom agent configurations
- Team size customization
- Integration with external research APIs

## License

MIT License - feel free to use and modify as needed.

## Contributing

This is a demonstration project. Feel free to fork and customize for your needs!

## Acknowledgments

Built with:
- Go programming language
- Anthropic Claude API
- Modern web standards (HTML5, CSS3, JavaScript)

---

**Note**: This system makes API calls to Anthropic's Claude service. Usage incurs costs based on Anthropic's pricing. Monitor your API usage accordingly.
