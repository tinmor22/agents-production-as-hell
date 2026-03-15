---
name: build-accumulated-context
description: Loads and merges all prior stage artifacts from the run directory into the accumulated context JSON for the next pipeline agent. Use between every stage in the orchestrator pipeline.
argument-hint: [run-dir] [next-stage-num]
allowed-tools: Read, Glob
user-invocable: false
---

Build the accumulated context JSON to pass to the next pipeline agent.

## Arguments

- `$ARGUMENTS[0]` — run directory path (e.g. `runs/2026-03-13-001`)
- `$ARGUMENTS[1]` — the next stage number (determines which artifacts to load)

If arguments are not provided, infer from the current orchestrator context.

## Steps

1. **List all artifacts** in the run directory using Glob: `{run-dir}/*.json`
2. **Read each artifact** that exists for stages prior to the next stage (see table below).
3. **Merge** into a single accumulated context object following the merge rules.
4. **Return** the constructed JSON — the orchestrator passes this as input to the next agent.

## What each stage receives

Build the context object by including all fields from the rows up to and including the current stage:

| Stage | Artifact to load | Fields to extract |
|---|---|---|
| 00 | `00-input.json` | `topic`, `quantity`, `constraints` |
| 01 | `01-nora-output.json` | `problems` |
| 02 | `02-leo-output.json` | `ideas` |
| 02a | `02a-user-selection.json` | `choosed_idea`, `other_ideas` |
| 03 | `03-maya-output.json` | `synthesis`, `solution_options`, `shortlist` |
| 04 | `04-sam-output.json` | `decision`, `monetization`, `metrics`, `mvp_definition` |
| 05 | `05-dani-output.json` | `design` |
| 06 | `06-omar-output.json` | `codebase_blueprint`, `plan` |
| 07 | `07-viktor-output.json` | `implementation` |
| 08 | `08-priya-output.json` | `observability` |
| 09 | `09-nate-output.json` | `deployment` |

## Context shape per agent

### → Maya (stage 03)
```json
{
  "topic": "...",
  "quantity": 5,
  "constraints": [...],
  "problems": [...],
  "ideas": [...],
  "choosed_idea": { "source": "nora|leo", "title": "...", "detail": {...} },
  "other_ideas": [...]
}
```

### → Sam (stage 04)
```json
{
  "choosed_idea": {...},
  "solution_options": [...],
  "shortlist": [...]
}
```

### → Dani (stage 05)
```json
{
  "choosed_idea": {...},
  "choosed_solution": {
    "chosen_option": "...",
    "target_user": "...",
    "north_star": "...",
    "positioning": "...",
    "must_have": [...],
    "non_goals": [...]
  },
  "decision": {...},
  "monetization": {...},
  "metrics": {...},
  "mvp_definition": {...}
}
```

### → Omar (stage 06)
```json
{
  "choosed_idea": {...},
  "choosed_solution": {...},
  "design": {...}
}
```

### → Viktor (stage 07)
```json
{
  "choosed_idea": {...},
  "choosed_solution": {...},
  "design": {...},
  "codebase_blueprint": {...},
  "plan": {...}
}
```

### → Priya / Nate (stages 08–09) — brief only
```json
{
  "chosen_option": "...",
  "target_user": "..."
}
```

### → Ada (stage 10) — summary only
```json
{
  "topic": "...",
  "chosen_option": "...",
  "target_user": "...",
  "north_star": "...",
  "stage_outcomes": {
    "nora": "<N> problems, top: <title>",
    "leo": "<N> ideas, top: <title>",
    "user_selection": "<title> (from <source>)",
    "maya": "shortlist: <names>",
    "sam": "<chosen_option> for <target_user>",
    "dani": "<design.brief_description>",
    "omar": "<N> milestones, week1: <slices>",
    "viktor": "<N> files, tests: <passed>",
    "priya": "<N> files instrumented",
    "nate": "deploy: <deploy_command>"
  }
}
```

## Rules

- If an artifact file is missing, skip it silently and note the gap in a comment.
- Never pass raw full artifacts to Priya, Nate, or Ada — use the brief/summary form only.
- The `choosed_solution` summary block must be constructed by the orchestrator from Sam's output fields, not copied directly.
- Only include fields the next agent actually needs — do not bloat context with irrelevant prior stages.
