# Quick Start Guide

## Setup (30 seconds)

1. **Get your API key** from [Anthropic Console](https://console.anthropic.com/)

2. **Set it as an environment variable:**
   ```bash
   export ANTHROPIC_API_KEY="your-api-key-here"
   # or
   export ANTHROPIC_KEY="your-api-key-here"
   ```

3. **Choose your interface:**

## ✨ Beautiful TUI (Recommended!)

**NEW!** Watch your agents collaborate in real-time with a gorgeous terminal interface!

```bash
./bin/cli-tui
```

**What you get:**
- 🔄 Animated spinners for each agent
- 📊 Live progress bars
- 💡 Ideas appearing in real-time
- 🎨 Color-coded agents
- ⏱️ Live statistics
- **Visual engagement like never before!**

**Perfect for:** Anyone who wants to *see* the magic happening!

[Full TUI Guide →](TUI_GUIDE.md)

---

## Version 1: Original (Simple & Fast)

**Best for:** Quick ideation with a focused 4-agent team

```bash
./bin/cli
```

**What you get:**
- 4 agents (Leader, Ideation, Moderator, UI Creator)
- 1 discussion round
- ~1-2 minutes
- Perfect for straightforward topics

## Version 2: Configurable (Recommended)

**Best for:** Deeper analysis with customizable teams

```bash
./bin/cli-v2
```

**What you get:**
- Choose team size: 4, 6, or 7 agents
- Multiple discussion rounds
- Specialized agents (Researcher, Critic, Implementer)
- ~1-10 minutes depending on config

### Three Configurations:

#### ⚡ Standard (4 agents, 1 round)
- Fast and focused
- Good for quick decisions
- 1-2 minutes

#### 🔬 Extended (6 agents, 2 rounds) ← Recommended
- Adds Researcher and Critic
- Multiple rounds for refinement
- 3-5 minutes
- **Best balance of depth and speed**

#### 🚀 Full (7 agents, 3 rounds)
- All 7 agents
- Maximum depth
- Includes implementation planning
- 5-10 minutes

## Web Interface

Prefer a visual interface?

```bash
./bin/server-v2
```

Then open: **http://localhost:8080**

**Features:**
- Click to select team configuration
- Real-time progress updates
- View results in browser
- No command line needed

## First Run Example

Let's try v2 with the Extended team:

1. **Start it:**
   ```bash
   ./bin/cli-v2
   ```

2. **Select configuration:**
   ```
   Choose configuration (1-4) [default: 2]: 2
   ```
   (Just press Enter for Extended)

3. **Enter a topic:**
   ```
   What topic would you like the AI team to explore?
   > Ways to improve team collaboration in remote work
   ```

4. **Watch the magic:**
   - 6 AI agents discuss your topic
   - 2 rounds of discussion with Leader synthesis
   - Ideas are generated, researched, and critically evaluated
   - Final HTML idea sheet is created

5. **View results:**
   - Open the generated `idea_sheet_*.html` file in your browser
   - See all ideas, scores, pros/cons, and the final selection

## Example Topics

### For Standard (Quick)
- "App features for habit tracking"
- "Blog post ideas for tech startups"
- "UX improvements for checkout flow"

### For Extended (Deeper)
- "Strategy to reduce food waste in restaurants"
- "Employee retention program for remote teams"
- "Product roadmap for AI writing assistant"

### For Full (Maximum Depth)
- "Complete go-to-market strategy for new SaaS"
- "Platform architecture for decentralized social network"
- "End-to-end solution for supply chain optimization"

## Tips for Success

### ✅ Good Topics:
- Specific and focused
- Open to multiple approaches
- Practical and actionable

### ❌ Avoid:
- Too broad: "How to fix everything"
- Too narrow: "What color should my logo be"
- Yes/no questions

### 💡 Pro Tips:
1. **Start with Extended config** - best balance
2. **Be specific** - "mobile app for X" better than "app ideas"
3. **Use Full config** for important decisions worth 5-10 minutes
4. **Try web interface** for better visibility

## What You'll Get

Every discussion produces:

1. **Console Output:**
   - Real-time progress
   - Ideas as they're generated
   - Final scores and summary

2. **HTML Idea Sheet:**
   - Beautiful visualization
   - All ideas explored
   - Pros and cons
   - Final recommendation
   - Discussion journey

3. **Structured Data:**
   - All ideas with scores
   - Complete message history
   - Round-by-round breakdown

## Programmatic Use

Want to integrate into your app?

```go
import (
    "github.com/yourusername/ai-agent-team/internal/models"
    "github.com/yourusername/ai-agent-team/internal/orchestrator"
)

config := models.ExtendedTeamConfig()
orch := orchestrator.NewConfigurableOrchestrator(apiKey, config)
err := orch.StartDiscussion("Your topic here")
```

See `example_v2.go` for complete examples.

## Troubleshooting

### "API key is required"
Set your environment variable:
```bash
export ANTHROPIC_API_KEY="sk-ant-..."
# or
export ANTHROPIC_KEY="sk-ant-..."
```

### "Permission denied"
Make binaries executable:
```bash
chmod +x bin/cli-v2 bin/server-v2
```

### Slow performance
- Normal! Extended takes 3-5 minutes
- Use Standard config for speed
- Full config can take 5-10 minutes

### High costs
- Extended: ~$0.40-$0.70 per discussion
- Use Standard for routine tasks
- Reserve Full for important decisions

## Next Steps

1. **Read README_V2.md** for detailed documentation
2. **Check USAGE.md** for advanced usage
3. **Try example_v2.go** for programmatic integration
4. **Experiment** with different team configurations

## Need Help?

- Check the error message - they're usually helpful
- Try with a simpler topic first
- Verify your API key is correct
- Ensure you have internet connectivity

---

**Ready? Let's go!**

```bash
# Try the beautiful TUI first!
export ANTHROPIC_API_KEY="your-key"
./bin/cli-tui
```

**Or use the standard CLI:**
```bash
./bin/cli-v2
```

Pick **Extended (option 2)**, enter your topic, and watch your AI team collaborate!
