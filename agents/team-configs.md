# Team Configurations

Three pre-built team configurations that trade off speed vs depth. Each adds more agents and exploration rounds.

> **Source:** `internal/models/config.go`

## Configuration Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `IncludeTeamLeader` | bool | Team Leader agent (coordinates team, makes decisions) |
| `IncludeUICreator` | bool | UI Creator agent (generates HTML report) |
| `IncludeIdeation` | bool | Ideation Specialist (generates creative ideas) |
| `IncludeModerator` | bool | Moderator (evaluates and scores ideas) |
| `IncludeResearcher` | bool | Researcher (provides factual grounding) |
| `IncludeCritic` | bool | Critical Analyst (challenges assumptions) |
| `IncludeImplementer` | bool | Implementation Specialist (plans execution) |
| `MaxRounds` | int | Number of exploration rounds |
| `MinIdeas` | int | Minimum ideas to generate |
| `DeepDive` | bool | Enable deeper back-and-forth |
| `MinScoreThreshold` | float64 | Minimum score for an idea to be considered |

---

## Standard Team (4 Agents)

The original, lean configuration. Fast results with the core workflow.

| Setting | Value |
|---------|-------|
| **Agents** | Team Leader, Ideation Specialist, Moderator, UI Creator |
| **MaxRounds** | 1 |
| **MinIdeas** | 3 |
| **DeepDive** | false |
| **MinScoreThreshold** | 6.0 |

**Active agents:**

| Agent | Role |
|-------|------|
| Team Leader | Frames the topic, synthesizes, selects winner |
| Ideation Specialist | Generates 3-5 creative ideas |
| Moderator | Scores and evaluates all ideas |
| UI Creator | Generates the final HTML report |

**Workflow:**
```
Kickoff → 1 Round (Ideation only) → Leader Synthesis → Validation → Selection → Report
```

**Best for:** Quick brainstorming sessions, simple topics, time-constrained scenarios.

**Trade-offs:** No research grounding, no critical analysis, no implementation planning. Ideas may be creative but unvetted.

---

## Extended Team (6 Agents)

Adds research and critical analysis for deeper exploration.

| Setting | Value |
|---------|-------|
| **Agents** | Team Leader, Ideation Specialist, Moderator, Researcher, Critical Analyst, UI Creator |
| **MaxRounds** | 2 |
| **MinIdeas** | 4 |
| **DeepDive** | true |
| **MinScoreThreshold** | 7.0 |

**Additional agents over Standard:**

| Agent | What It Adds |
|-------|-------------|
| Researcher | Facts, data, case studies, market context |
| Critical Analyst | Assumption challenges, risk identification, "what if" scenarios |

**Workflow:**
```
Kickoff
  → Round 1: Research → Ideation → Critique → Leader Synthesis
  → Round 2: Research → Ideation (refined) → Critique → Leader Synthesis
→ Validation → Selection → Report
```

**Best for:** Topics requiring factual grounding, complex domains, decisions with real consequences.

**Trade-offs:** Takes longer (2 rounds, 6 agents per round). The higher MinScoreThreshold (7.0) means only strong ideas survive. No implementation planning.

---

## Full Team (7 Agents)

All agents active. Maximum depth and analysis.

| Setting | Value |
|---------|-------|
| **Agents** | Team Leader, Ideation Specialist, Moderator, Researcher, Critical Analyst, Implementation Specialist, UI Creator |
| **MaxRounds** | 3 |
| **MinIdeas** | 5 |
| **DeepDive** | true |
| **MinScoreThreshold** | 7.5 |

**Additional agent over Extended:**

| Agent | What It Adds |
|-------|-------------|
| Implementation Specialist | Practical execution plans, MVP thinking, resource requirements |

**Workflow:**
```
Kickoff
  → Round 1: Research → Ideation → Critique → Implementation → Leader Synthesis
  → Round 2: Research → Ideation (refined) → Critique → Implementation → Leader Synthesis
  → Round 3: Research → Ideation (final) → Critique → Implementation → Leader Synthesis
→ Validation → Selection → Report
```

**Best for:** Strategic planning, product decisions, topics where you need both creative ideas AND practical execution plans.

**Trade-offs:** Slowest configuration (3 rounds, 7 agents per round = 21+ LLM calls). The highest MinScoreThreshold (7.5) is very selective. Most expensive in API costs.

---

## Comparison Matrix

| Feature | Standard | Extended | Full |
|---------|----------|----------|------|
| Agents | 4 | 6 | 7 |
| Rounds | 1 | 2 | 3 |
| Research | No | Yes | Yes |
| Critique | No | Yes | Yes |
| Implementation | No | No | Yes |
| Min Ideas | 3 | 4 | 5 |
| Score Threshold | 6.0 | 7.0 | 7.5 |
| Deep Dive | No | Yes | Yes |
| Approx. LLM Calls | ~5 | ~14 | ~24 |
| Speed | Fast | Moderate | Slow |
| Depth | Surface | Deep | Comprehensive |

> **Scaling Insight:** Each additional agent and round adds multiplicative cost. Going from Standard (4 agents, 1 round) to Full (7 agents, 3 rounds) is roughly a 5x increase in LLM calls. Choose the configuration that matches your needs — not every topic requires the full team.

## Custom Configurations

You can create your own team configuration by mixing and matching agents. For example:

**Research-Heavy Team (5 agents, 2 rounds):**
- Team Leader, Researcher, Ideation Specialist, Moderator, UI Creator
- Focus: Well-researched ideas without the overhead of critique and implementation

**Critique-Heavy Team (5 agents, 2 rounds):**
- Team Leader, Ideation Specialist, Critical Analyst, Moderator, UI Creator
- Focus: Stress-tested ideas without research or implementation detail

**Implementation-Focused Team (5 agents, 1 round):**
- Team Leader, Ideation Specialist, Implementation Specialist, Moderator, UI Creator
- Focus: Practical, actionable ideas with execution plans
