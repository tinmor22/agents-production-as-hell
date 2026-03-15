---
name: validate-agent-input
description: Validates the JSON input for a pipeline agent before it runs. Use when an agent receives input and needs to check required fields are present and non-empty before proceeding.
argument-hint: [agent-name]
allowed-tools: Read
---

You are validating the JSON input for **$ARGUMENTS** before the agent runs.

## Steps

1. **Identify the expected schema** for the agent from the list below (or from the agent's own SKILL.md / `.claude/agents/*.md` definition if not listed).

2. **Check each required field**:
   - Field exists in the input JSON
   - Field is not null, empty string, empty array, or empty object
   - Field has the correct type (string, array, object, bool)

3. **Report the result**:
   - If valid: confirm with `✓ Input valid for <agent>` and list the key fields found.
   - If invalid: list every failing field with a clear reason, then **stop** — do not proceed with the agent task. Output:
     ```
     ✗ Input invalid for <agent>
     - <field>: <reason>
     ```

## Known agent schemas

### Nora (1_problematic_nora)
```
{ "topic": string, "quantity": int (>0), "constraints": string[] }
```

### Leo (1_dreamer_leo)
```
{ "topic": string, "quantity": int (>0), "constraints": string[] }
```

### Maya (2_brainstormer_maya)
```
{ "problems": array (non-empty), "ideas": array (non-empty) }
```
Each problem must have: `title`, `problem_statement`, `severity`, `confidence`
Each idea must have: `title`, `one_liner`, `target_user`

### Sam (3_solver_sam)
```
{ "synthesis": string, "solution_options": array (non-empty), "shortlist": array (2–3 items) }
```

### Dani (4_designer_dani)
```
{ "decision": { chosen_option, rationale, target_user }, "metrics": { north_star }, "mvp_definition": { must_have: array (non-empty) } }
```
Hard gate: `decision.target_user` and `metrics.north_star` must be non-empty strings.

### Omar (5_planner_omar)
```
{ "design": { architecture: { components: array (non-empty) }, api_design: array, data_model: array } }
```

### Viktor (6_coder_viktor)
```
{ "codebase_blueprint": { module_name, go_version, packages }, "plan": { milestones: array (non-empty) } }
```
Hard gate: `plan.milestones` must be non-empty.

### Priya (7_observability_priya) / Nate (8_deployer_nate)
```
{ "chosen_option": string, "target_user": string }
```
Both fields must be non-empty strings.

### Rosa (9_maintainer_rosa)
```
{ "signals": array (non-empty) }
```
Each signal must have: `type` (one of: error, metric, feedback), `message`

### Ada (10_retro_ada)
```
{ "topic": string, "chosen_option": string, "target_user": string, "north_star": string, "stage_outcomes": object }
```

## Rules
- Fail on the first structural error that would cause the agent to crash silently.
- Do not attempt to fix or default missing fields — just report them.
- If the agent name is unknown, ask the user to paste the expected schema.
