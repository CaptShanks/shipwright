# Creating custom skills for Shipwright

This guide describes how to add reusable **skills** to the Shipwright framework so agents can load domain expertise, checklists, and examples consistently.

## What is a skill?

A skill is a **reusable knowledge module** that agents can equip at runtime. It packages how to reason about a problem space: principles, tradeoffs, and concrete patterns -- not a single script or command list.

Skills typically include:

- **Domain expertise** (security, testing, language idioms, review standards)
- **Checklists** you can verify against real work
- **Examples** that show good and bad shapes of solutions

Shipwright distinguishes two kinds:

| Kind | Role |
|------|------|
| **Foundational** | Cross-cutting habits that many agents share (e.g. code quality, security awareness) |
| **Specialized** | Deep focus on one domain or stack (e.g. OWASP details, Go idioms) |

## Skill directory structure

Each skill lives under `_skills/` in the Shipwright repository:

```text
_skills/{skill-name}/
  SKILL.md          # Main skill definition (required)
  checklists/       # Actionable checklists (optional)
    checklist-1.md
  examples/         # Code examples and patterns (optional)
    example-1.md
  scripts/          # Executable code (optional, per spec)
  references/       # Additional documentation (optional, per spec)
  assets/           # Templates and static resources (optional, per spec)
```

- **`SKILL.md`** is the entry point agents resolve first.
- **`checklists/`** holds one file per concern (auth, injection, naming, etc.).
- **`examples/`** holds focused snippets or walkthroughs referenced from `SKILL.md` or checklists.
- **`scripts/`**, **`references/`**, **`assets/`** are also valid optional directories. Any additional directories are allowed, which is why we use `checklists/` and `examples/` for clarity.

Use a short, kebab-case **`skill-name`** that matches the folder name everywhere (frontmatter, agent `skills` list, symlinks).

### Naming and consistency

- **`name` in frontmatter** must equal the directory name under `_skills/` and the symlink basename under `plugins/.../skills/`.
- **One skill, one theme.** If a file tries to cover "all of backend," split it: foundational content in a thin `SKILL.md`, specialized depth in checklists or a separate specialized skill.
- **Cross-references:** Use relative paths from the skill root (e.g. `checklists/auth.md`) in prose so maintainers can grep and move files predictably.

## Writing `SKILL.md`

Start with YAML **frontmatter** so tooling and agents can discover the skill:

```yaml
---
name: skill-name
description: >-
  What this skill covers and when to use it. Include activation triggers
  so agents know when to load the skill. Use when writing, reviewing, or
  debugging code that involves X.
license: MIT
metadata:
  author: CaptShanks
  version: "1.0.0"
---
```

**Required fields** (per the Agent Skills spec):

- `name` -- Max 64 characters. Lowercase letters, numbers, and hyphens only. Must not start or end with a hyphen. Must match the parent directory name.
- `description` -- Max 1024 characters. Must describe both **what the skill does** and **when to use it** (activation triggers help agents decide when to load the skill).

**Optional fields:**

- `license` -- License name or reference to a bundled license file.
- `compatibility` -- Max 500 characters. Environment requirements (e.g. "Requires Python 3.14+ and uv").
- `metadata` -- Arbitrary key-value mapping (author, version, etc.).
- `allowed-tools` -- Space-separated string of pre-approved tools the skill may use (experimental).

Body content should emphasize **principles, patterns, and decision guidance**: when to choose A vs. B, what invariants matter, how to recognize the wrong abstraction. Prefer "how to think" over a flat "what to do" bullet list -- checklists carry the verifiable steps.

Guidelines:

- Keep it **concise and actionable**; aim for **under ~500 lines / ~5000 tokens** in `SKILL.md` (per the spec's progressive disclosure model) and push detail into checklists or examples.
- Link or name checklist/example files so agents can drill in without loading everything at once.
- Order sections from **stable principles** to **typical decisions** to **pointers** to checklists/examples, so the main file reads like a map, not an encyclopedia.

### Decision guidance (what "how to think" looks like)

Good skill prose answers questions such as: *When is duplication acceptable?* *Where should validation live?* *What evidence would change your default?* Bad skill prose is only a flat policy list with no tradeoffs. Use short scenarios ("given a public HTTP handler...") only when they disambiguate a rule; otherwise keep scenarios in `examples/`.

## Writing checklists

Each checklist file should cover **one specific concern** (e.g. secret handling, SQL injection, error propagation).

- Use **actionable, verifiable** items (something a reviewer or agent can confirm yes/no).
- Include **good vs. bad** micro-examples where it clarifies the rule.
- Call out **patterns** you want repeated and **anti-patterns** to avoid.
- Keep items parallel in style (imperative verbs, consistent depth).

## Writing examples

Examples belong in `examples/` when they would clutter `SKILL.md` or are reused from several checklists.

- Show **complete, working** code when the point is syntax or wiring; otherwise isolate the minimal slice that teaches one idea.
- Prefer **problematic vs. clean** side-by-side (or before/after) when teaching a refactor or habit.
- **Annotate** non-obvious decisions briefly (why this API, why this boundary).
- **One concept per file** keeps skills composable and easier to maintain.

Use **language-specific** examples when the skill targets a language or runtime; use language-agnostic patterns when the skill is about process or architecture.

## Registering your skill

1. **Create** the directory `_skills/{skill-name}/` with `SKILL.md` (and optional `checklists/`, `examples/`).
2. **Symlink** the skill into each plugin that should expose it: from the plugin root, add `skills/{skill-name}` pointing at `../../../_skills/{skill-name}` (adjust depth if the plugin layout differs).
3. **Update** that plugin's `.claude-plugin/plugin.json`: add `"skills/{skill-name}"` to the `skills` array (paths are relative to the plugin root).
4. **Optionally** ship a **standalone skill plugin** if the skill should be installable without bundling a specific agent -- same structure, but the plugin's only job is to publish `skills/` and `plugin.json`.

After registration, agents that list the skill can load the same content whether it is linked from one plugin or many.

If you add a **new top-level plugin** to the Shipwright marketplace, update `.claude-plugin/marketplace.json` so the plugin appears in the catalog with an accurate description (including which skills it bundles).

## Assigning skills to agents

- In the agent markdown file (under `plugins/.../agents/`), add the skill **name** to the YAML **`skills`** array (e.g. `security-awareness`), matching `name` in `SKILL.md` frontmatter.
- Ensure the agent's plugin directory contains the **`skills/` symlink** (or copy, if your fork uses a different layout) so the resolver can find `_skills/{skill-name}` via the plugin's `skills` entry in `plugin.json`.

Agents only use skills that are both **listed in frontmatter** and **available through the plugin's `plugin.json` + `skills/` links**.

### Verification

After wiring a skill:

1. Confirm the symlink resolves: `ls -l plugins/<your-plugin>/skills/<skill-name>` should point at `_skills/<skill-name>`.
2. Confirm `plugin.json` includes `"skills/<skill-name>"` and the agent's frontmatter lists the same logical `name` as in `SKILL.md`.
3. Open `SKILL.md` and one checklist in the client you use for agents to ensure paths and formatting render as expected.

## Tips

- **Start with checklists** -- they give the fastest payoff for agent behavior and human review.
- Prefer **language-specific** examples when the skill is meant for implementers in that stack; use **language-agnostic** examples for process skills (review, incident response, design docs).
- **Cap `SKILL.md` length** (~500 lines); move depth to checklists and examples so the main file stays skimmable in one pass.
- **Iterate** skills when you see repeated mistakes or great outputs in agent runs -- that feedback loop is how the library stays honest.
- **Avoid duplicate rules** across skills: if two skills repeat the same invariant, extract a foundational skill and link to it from specialized ones.
- **Version mentally, not just in git:** when you change a rule, update every checklist item and example that implied the old behavior.
- **Validate** your skills locally: `python3 scripts/ci_validate.py skills` from the repo root.

## Common pitfalls

| Pitfall | What to do instead |
|--------|---------------------|
| `SKILL.md` is a giant tutorial | Split into `examples/`; keep the main file principled and short |
| Checklist items are vague ("be careful with errors") | Rewrite as verifiable checks ("errors include cause; no empty `catch`") |
| Skill `name` does not match folder name or symlink | Align all three; mismatches break discovery |
| Skill listed in agent YAML but missing from `plugin.json` | Add the `skills/...` entry and symlink |
| Only bad examples, no fix | Always show the preferred pattern or a labeled "better" variant |
| `description` only says what, not when | Add activation triggers: "Use when..." |

For reference layouts, browse existing skills under `_skills/` in this repository and mirror their naming and symlink patterns in your plugin.
