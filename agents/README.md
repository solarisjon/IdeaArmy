# Markdown-Based Agent Army

A no-code agent framework where each agent is a `.md` file defining its role, personality, skills, and behavior. A master orchestrator file defines how they interact.

This mirrors the full 7-agent team from [IdeaArmy](../CLAUDE.md) — the same system prompts, temperatures, and workflow phases — but expressed entirely as readable markdown that an LLM can orchestrate directly.

## Directory Structure

```
agents/
├── README.md                    # You are here
├── orchestrator.md              # Master workflow: phases, agent order, message flow
├── team-configs.md              # Standard / Extended / Full team presets
│
├── agents/
│   ├── team-leader.md           # Coordinates the team, makes final decisions
│   ├── ideation-specialist.md   # Generates creative ideas
│   ├── moderator.md             # Evaluates and scores ideas
│   ├── researcher.md            # Provides research and factual grounding
│   ├── critical-analyst.md      # Challenges assumptions, finds weaknesses
│   ├── implementation-specialist.md  # Plans practical execution
│   └── ui-creator.md            # Generates the final HTML report
│
└── examples/
    └── sample-session.md        # Walkthrough of a full session
```

## How to Use

### 1. Study the Agents

Read each file in `agents/` to understand how agent skills differ and why. Pay attention to:
- **Temperature settings** — why a Researcher uses 0.4 but an Ideation Specialist uses 0.9
- **System prompts** — how each agent's personality shapes its output
- **Phase behavior** — what each agent does at each stage of the workflow

### 2. Run a Session with an LLM

Ask an LLM (like Claude Code) to orchestrate a session:

> "Read `agents/orchestrator.md` and run a session on the topic 'sustainable packaging for e-commerce' using the agent definitions in `agents/agents/`. Use the Standard team config from `agents/team-configs.md`."

The LLM will:
1. Read the orchestrator workflow
2. Load the relevant agent definitions
3. Simulate each agent's contribution in sequence
4. Follow the phase structure (Kickoff → Exploration → Validation → Selection → Visualization)

### 3. Modify and Experiment

- Change a temperature and see how output changes
- Add a new agent `.md` file with a custom role
- Remove an agent and observe what's missing
- Tweak a system prompt to shift an agent's behavior

### 4. Compare Team Configurations

Try the same topic with different team sizes (see `team-configs.md`):
- **Standard (4 agents)** — fast, focused
- **Extended (6 agents)** — deeper analysis with research and critique
- **Full (7 agents)** — maximum depth with implementation planning

## Key Learning Concepts

### Why Temperature Matters

Temperature controls randomness in LLM output (0.0 = deterministic, 1.0 = maximum creativity):

| Agent | Temperature | Why |
|-------|------------|-----|
| Researcher | 0.4 | Needs factual accuracy — less randomness = more reliable |
| Moderator | 0.5 | Analytical evaluation — consistent scoring needed |
| Critic | 0.6 | Some creativity for "what if" questions, but grounded |
| Implementer | 0.6 | Practical thinking — creative enough for solutions, realistic enough for plans |
| UI Creator | 0.6 | Balanced — creative presentation, structured output |
| Team Leader | 0.7 | Needs flexibility to synthesize diverse inputs |
| Ideation Specialist | 0.9 | Maximum creativity — wild ideas are the goal |

### Context Accumulation

Each agent sees **all prior messages** in the discussion. This means:
- The Researcher's findings inform the Ideation Specialist's ideas
- The Critic's challenges are visible to the Implementer
- The Team Leader sees everything when making the final selection

This shared context is what makes multi-agent cooperation more than just running prompts in parallel.

### Phase Sequencing

The phases run in a specific order for a reason:
1. **Kickoff** — Frame the problem before generating solutions
2. **Exploration** — Research → Ideate → Critique → Implement (each builds on the last)
3. **Validation** — Score ideas only after they've been fully explored
4. **Selection** — Choose only after scoring
5. **Visualization** — Report only after the decision is made

Reordering these phases (e.g., critiquing before ideating) would fundamentally change the output quality.

## Source Code Mapping

Each markdown agent was extracted from the corresponding Go source file:

| Markdown File | Go Source |
|--------------|-----------|
| `orchestrator.md` | `internal/orchestrator/orchestrator_v2.go` |
| `team-configs.md` | `internal/models/config.go` |
| `agents/team-leader.md` | `internal/agents/team_leader.go` |
| `agents/ideation-specialist.md` | `internal/agents/ideation.go` |
| `agents/moderator.md` | `internal/agents/moderator.go` |
| `agents/researcher.md` | `internal/agents/researcher.go` |
| `agents/critical-analyst.md` | `internal/agents/critic.go` |
| `agents/implementation-specialist.md` | `internal/agents/implementer.go` |
| `agents/ui-creator.md` | `internal/agents/ui_creator.go` |
