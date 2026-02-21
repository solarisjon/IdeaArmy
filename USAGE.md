# Usage Guide

## Quick Start

### 1. Set Your API Key

```bash
export ANTHROPIC_API_KEY="your-api-key-here"
# or
export ANTHROPIC_KEY="your-api-key-here"
```

Or create a `.env` file:
```
ANTHROPIC_API_KEY=your-api-key-here
# or
ANTHROPIC_KEY=your-api-key-here
```

### 2. Choose Your Interface

#### Option A: CLI (Command Line)

```bash
# Using go run
go run cmd/cli/main.go

# Or using the built binary
./bin/cli
```

**Perfect for:**
- Quick ideation sessions
- Terminal workflows
- Automation scripts
- CI/CD integration

#### Option B: Web Interface

```bash
# Using go run
go run cmd/server/main.go

# Or using the built binary
./bin/server
```

Then open your browser to `http://localhost:8080`

**Perfect for:**
- Interactive sessions
- Sharing results with teams
- Better visualization
- Non-technical users

#### Option C: Programmatic

```bash
go run example.go
```

**Perfect for:**
- Custom integrations
- Batch processing
- Building on top of the system
- Advanced automation

## Example Topics

### Product Ideas
- "Innovative mobile apps for elderly users to stay connected with family"
- "Eco-friendly alternatives to common household products"
- "AI-powered tools for small business inventory management"

### Business Strategy
- "Growth strategies for a sustainable fashion startup"
- "Ways to improve customer retention for SaaS products"
- "Marketing approaches for B2B software in healthcare"

### Social Impact
- "Solutions to reduce food waste in restaurants"
- "Programs to improve financial literacy in underserved communities"
- "Initiatives to promote mental health awareness in workplaces"

### Technology
- "Applications of AR in education"
- "Blockchain use cases beyond cryptocurrency"
- "IoT solutions for smart cities"

### Process Improvement
- "Ways to make remote team meetings more effective"
- "Strategies to reduce meeting time while maintaining productivity"
- "Systems to improve developer onboarding"

## Understanding the Output

### Discussion Phases

The agents work through five phases:

1. **Kickoff** - Team Leader sets the direction
2. **Ideation** - Creative ideas are generated
3. **Validation** - Ideas are critically evaluated
4. **Selection** - Best idea is chosen
5. **Visualization** - HTML idea sheet is created

### Idea Sheet Contents

The generated HTML file includes:

- **Summary Section**: Overview of the discussion
- **All Ideas**: Every idea that was considered
- **Idea Scores**: Validation scores (0-10)
- **Pros & Cons**: For each idea
- **Final Selection**: The chosen idea with full details
- **Journey**: Key insights from the discussion

### Reading the Scores

Ideas are scored on multiple criteria:

- **8-10**: Excellent - High feasibility and impact
- **6-7**: Good - Solid concept with minor concerns
- **4-5**: Fair - Needs work or has significant challenges
- **0-3**: Poor - Major issues or not viable

## Tips for Best Results

### Writing Good Topics

✅ **Good Topics:**
- Specific and focused
- Open to multiple approaches
- Practical and actionable
- Clear scope

❌ **Topics to Avoid:**
- Too broad ("How to fix everything")
- Too narrow ("What color should my logo be")
- Subjective preferences ("Best ice cream flavor")
- Questions with yes/no answers

### Topic Examples

**Too Broad:**
"How can we improve education?"

**Better:**
"Interactive tools to help high school students learn calculus concepts"

**Too Narrow:**
"Should I use React or Vue?"

**Better:**
"Frontend architecture approaches for a real-time collaborative editing app"

## Advanced Usage

### Customizing Agent Behavior

Edit the system prompts in:
- `internal/agents/team_leader.go`
- `internal/agents/ideation.go`
- `internal/agents/moderator.go`
- `internal/agents/ui_creator.go`

### Changing the Model

Edit `internal/claude/client.go`:
```go
const DefaultModel = "claude-sonnet-4-20250514"
```

Available models:
- `claude-sonnet-4-20250514` - Balanced (default)
- `claude-opus-4-20250514` - Most capable
- `claude-3-5-sonnet-20241022` - Previous generation

### Adjusting Temperature

Each agent has a temperature setting in their constructor:

```go
Temperature: 0.7  // Lower = more focused, Higher = more creative
```

- Team Leader: 0.7 (balanced)
- Ideation: 0.9 (creative)
- Moderator: 0.5 (analytical)
- UI Creator: 0.6 (balanced)

### Custom Orchestration

Create your own flow by modifying `internal/orchestrator/orchestrator.go`:

```go
// Add a new phase
func (o *Orchestrator) runPhase6_CustomPhase() error {
    // Your custom logic
}
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ANTHROPIC_API_KEY` or `ANTHROPIC_KEY` | (required) | Your Anthropic API key |
| `PORT` | `8080` | Web server port (server mode only) |

## Output Files

### CLI Mode
- Saves: `idea_sheet_<timestamp>.html` in current directory

### Web Mode
- Displays inline in browser
- Can save via browser's "Save Page As"

### Programmatic Mode
- Full access to `Discussion` object
- Can save anywhere with custom names
- Can export to JSON, etc.

## Performance Notes

### Typical Discussion Times

- **Simple topics**: 30-60 seconds
- **Complex topics**: 1-3 minutes
- **Very detailed analysis**: 3-5 minutes

### API Usage

Each discussion typically uses:
- 5-10 API calls
- 10,000-30,000 tokens total
- Cost: ~$0.15-$0.50 per discussion (varies by model)

### Optimization Tips

1. **Use specific topics** - Reduces token usage
2. **CLI mode** - Faster startup than web server
3. **Reuse orchestrator** - In programmatic mode
4. **Lower temperature** - Slightly faster responses

## Troubleshooting

### "Rate limit exceeded"
Wait a moment and try again. Anthropic has rate limits.

### "Context length exceeded"
Your topic may be too complex. Try breaking it into smaller topics.

### Slow performance
- Check your internet connection
- Consider using a faster model
- Reduce the number of ideas generated (edit `ideation.go`)

### Empty HTML output
- Check the discussion completed successfully
- Verify all 5 phases ran
- Check for error messages in the log

## Integration Examples

### With Slack

```go
// Post idea sheet to Slack channel
html := orch.GetIdeaSheetHTML()
// Convert to Slack blocks or upload as file
```

### With Database

```go
// Save discussion to database
discussion := orch.GetDiscussion()
db.Save(discussion)
```

### With Email

```go
// Email the idea sheet
html := orch.GetIdeaSheetHTML()
emailClient.Send(recipient, "Idea Sheet", html)
```

## Best Practices

1. **Start Simple**: Try the CLI first
2. **Review Prompts**: Customize agent prompts for your domain
3. **Iterate**: Run multiple discussions on related topics
4. **Share Results**: Use the HTML output for presentations
5. **Monitor Costs**: Track your API usage
6. **Save Discussions**: Keep a library of generated ideas

## Support

For issues or questions:
- Check the README.md
- Review example.go for code samples
- Examine the agent system prompts
- Test with simple topics first

---

Happy ideating! 🚀
