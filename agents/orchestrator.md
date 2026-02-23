# Orchestrator Workflow

The orchestrator coordinates a configurable team of agents through a structured, multi-phase discussion. It manages agent execution order, message flow, and decision-making.

> **Source:** `internal/orchestrator/orchestrator_v2.go`

## Overview

```
User Topic
    │
    ▼
┌─────────────────┐
│  Phase 1:       │
│  Kickoff        │  Team Leader frames the discussion
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Phase 2:       │  ┌──────────────────────────────────┐
│  Exploration    │──│ Repeat for MaxRounds (1-3 rounds) │
│  Rounds         │  └──────────────────────────────────┘
│                 │
│  Each round:    │
│  1. Researcher  │  (if available)
│  2. Ideation    │  Generate/refine ideas
│  3. Critic      │  (if available, and ideas exist)
│  4. Implementer │  (if available, and ideas exist)
│  5. Leader      │  Synthesize the round
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Phase 3:       │
│  Validation &   │  Moderator scores all ideas
│  Selection      │  Team Leader selects the winner
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Phase 4:       │
│  Visualization  │  UI Creator generates HTML report
└─────────────────┘
```

## Phase 1: Kickoff

**Who:** Team Leader (required)

**What happens:**
1. The orchestrator initializes a new `Discussion` with the user's topic
2. The Team Leader receives the topic and a list of team members
3. The Team Leader sets direction: what each team member should focus on

**Prompt sent to Team Leader:**
```
We have a team of {N} agents to explore: {topic}

Team members: {list of agent names, excluding Team Leader and UI Creator}

Please set the direction for this discussion. What should each team member focus on?
```

**Output:** A kickoff message added to the discussion context with type `"kickoff"`.

> **Why Kickoff First?** The Team Leader frames the problem and assigns focus areas. Without this, agents would all explore the same angles independently, wasting effort.

## Phase 2: Exploration Rounds

**Who:** Researcher → Ideation → Critic → Implementer (in that order, if available)

**Repeats:** `MaxRounds` times (1 for Standard, 2 for Extended, 3 for Full)

Each round follows this sequence:

### Step 1: Research (if Researcher is in team)
- **Prompt:** "Provide research and context for this topic"
- **Purpose:** Ground the discussion in facts before creative work begins

### Step 2: Ideation
- **Prompt (Round 1):** "Generate creative ideas based on the discussion so far"
- **Prompt (Round 2+):** "Building on previous ideas and feedback, generate refined or new creative ideas"
- **Purpose:** Generate 3-5 structured ideas as JSON
- **Special behavior:** Ideas are parsed from JSON and added to `Discussion.Ideas`

### Step 3: Critical Analysis (if Critic is in team, and ideas exist)
- **Prompt:** "Challenge the assumptions in these ideas. What could go wrong?"
- **Purpose:** Stress-test ideas before they advance
- **Guard:** Only runs if there are ideas to critique

### Step 4: Implementation Thinking (if Implementer is in team, and ideas exist)
- **Prompt:** "How would we actually implement these ideas? What's the practical approach?"
- **Purpose:** Ground ideas in practical reality
- **Guard:** Only runs if there are ideas to plan for

### Step 5: Leader Synthesis (after each round)
- **Prompt:** "Synthesize the contributions from round {N}. What are the key insights? What should the team focus on in the next round? If this is the final round, identify which ideas are strongest."
- **Purpose:** The Team Leader digests the round and redirects the team

> **Why this order?** Research provides context for ideation. Ideation generates material for critique. Critique identifies weaknesses for the implementer to address. The leader synthesizes everything. Each step builds on the previous one.

## Phase 3: Validation & Selection

### Final Validation (Moderator)

**Who:** Moderator (if available; if not, skips to selection)

**Prompt:** "Provide final scores and comprehensive evaluation of all ideas discussed"

**What happens:**
1. Moderator evaluates all ideas using a 0-10 scoring rubric
2. Evaluations are returned as JSON with scores, pros, cons, and feedback
3. Ideas in `Discussion.Ideas` are updated with scores and validation status

**Evaluation criteria:**
- Feasibility (0-10)
- Innovation (0-10)
- Impact (0-10)
- Clarity (0-10)
- Completeness (0-10)

### Final Selection (Team Leader)

**Who:** Team Leader (if available; otherwise auto-select highest score)

**Prompt:** "Based on all the discussion, evaluation, and team input, select the best idea and explain your decision"

**What happens:**
1. Team Leader reviews all scored ideas and team input
2. Makes a final selection with explanation
3. The highest-scored idea is set as `Discussion.FinalIdea`

> **Decision Rule:** The system always selects the highest-scored idea as the final idea, regardless of the Team Leader's stated preference. The leader's commentary provides qualitative context, but the score is authoritative.

## Phase 4: Visualization

**Who:** UI Creator (optional — discussion succeeds without it)

**What happens:**
1. The UI Creator receives the complete discussion context (all messages, all ideas with scores, the final selection)
2. It generates a comprehensive, multi-section HTML report
3. The HTML is stored as a message with type `"visualization"`

**Error handling:** If report generation fails, the orchestrator logs a warning but does not fail the discussion. The visualization is a nice-to-have, not a requirement.

## Message Flow

Every agent contribution is recorded as a `Message` in the discussion:

```
Message {
    ID:        unique identifier
    From:      agent role (e.g., "ideation") or "system"
    To:        "team" (broadcast) or specific role
    Content:   the agent's full response text
    Timestamp: when the message was created
    Type:      "kickoff" | "ideation" | "researcher" | "critic" |
               "implementer" | "synthesis" | "validation" | "selection" |
               "visualization"
}
```

### Context Accumulation

Each agent receives the **full discussion history** when it runs. The context is built by `BuildContext()`:

```
Topic: {topic}

Previous Discussion:
[from -> to]: {content}
[from -> to]: {content}
...

Current Ideas:
1. {title} - {description}
   Score: {score}/10
2. ...
```

This means later agents have more context than earlier ones. The Implementer in round 2 sees everything the Researcher, Ideation Specialist, Critic, and Team Leader said in rounds 1 and 2.

## Decision Rules

| Parameter | Standard | Extended | Full | Purpose |
|-----------|----------|----------|------|---------|
| MaxRounds | 1 | 2 | 3 | How many exploration cycles |
| MinIdeas | 3 | 4 | 5 | Minimum ideas to generate |
| DeepDive | false | true | true | Enable extra back-and-forth |
| MinScoreThreshold | 6.0 | 7.0 | 7.5 | Minimum score to consider |

## Error Handling

- **Agent failure during exploration:** Logged as a warning, but the round continues. One agent failing doesn't stop the others.
- **Team Leader required:** The kickoff phase will fail if no Team Leader is configured.
- **No ideas to validate:** Final validation fails if no ideas were generated.
- **Visualization failure:** Non-fatal. The discussion is marked as completed even if the report can't be generated.
