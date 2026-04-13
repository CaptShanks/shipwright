# Creating custom agents for Shipwright

This guide describes how to author agent definitions (Markdown with YAML frontmatter), register them as Claude plugins in this repository, and validate them locally in Claude Code and in GitHub Actions.

**Audience** — Maintainers adding a new agent plugin, or teams forking Shipwright to encode their own engineering rituals.

## Agent anatomy

Each agent is a single `.md` file. The file **must** begin with YAML frontmatter closed by a second `---` line. The loader and CI expect a `name` field at minimum.

```yaml
---
name: agent-name
description: What this agent does
skills: [skill-1, skill-2]
---
```

Optional frontmatter keys (used by some agents in this repo) include `model`, `effort`, and `maxTurns`. Skills named in frontmatter are a hint to authors; **Claude loads skills from `plugin.json`**, so keep those two in sync when you use skills.

The body of the file is plain Markdown: headings, lists, and fenced blocks that the model reads as the agent’s system-style instructions.

**Naming** — Pick a stable `name` in frontmatter (often `kebab-case`, matching how operators refer to the agent). Workflow inputs and docs should use the same string family to avoid “two names for one agent” confusion.

## Required sections

Structure the body so the model always knows **what it is**, **how to think**, **what rules bind it**, **how to work**, and **what to emit**.

1. **Role** — One tight paragraph: remit, boundaries (what is in and out of scope), and who consumes the output (humans, bots, downstream agents). State negatives plainly (“does not commit code,” “does not choose sprint dates”) so the model does not drift into a neighboring job.

2. **Mindset** — Bullets that shape tone and tradeoffs (e.g. speed vs. precision, when to ask questions vs. decide). This is the agent’s “default stance,” not a repeat of the role. Good mindsets name the failure mode they prevent (“resist the urge to paste a patch during triage”).

3. **Core principles** — **8–12 numbered rules** that resolve conflicts (e.g. “completeness over optimism,” “structured output always”). Order them so earlier rules trump later ones when ambiguous. Each principle should be enforceable in a yes/no check, not a vibe.

4. **Process** — A **numbered workflow** from intake to completion. Each step should be actionable (“read X, then Y”) so the model can follow the same path every time. Call out where tools or repo search are expected (“search issues for duplicates before classifying”).

5. **Quality checklist** — A short self-verification list the model runs mentally before sending output. Example pattern:

   - Does the answer satisfy every item in the output contract?
   - Did I apply the process steps in order, skipping none?
   - Did I handle the stated edge cases (missing repro, mixed requests, policy conflicts)?
   - Is anything asserted without evidence from the provided context?
   - Would a downstream agent know the next action without follow-up chat?

6. **Anti-patterns** — Explicit “do not” list tied to this role: generic platitudes, skipping gates, mixing roles (e.g. triage writing patches), vague closing summaries, invented repository facts, or “fixing” scope without recording the assumption. Tie each anti-pattern to a principle or process step it violates.

7. **Output contract** — Exact shape of the response: required headings, fields, JSON blocks, labels, or comment templates. Downstream automation and humans should parse it without guessing. Include a **minimal good example** and, if useful, a machine-readable snippet (JSON or YAML) with `null`/placeholder values.

Additional sections (e.g. type-specific criteria, glossaries, or “when to escalate”) are welcome when they reduce ambiguity; keep the seven above as the spine.

## Writing effective agents

- **Be specific and opinionated** — Name real artifacts (issues, PRs, files), real policies, and real failure modes for *this* stack. Avoid “be helpful” and “consider security” without concrete gates.
- **Use decision frameworks** — When A vs B, spell out the decision tree (“if repro missing → needs-info; if duplicate found → link and close”). Prefer “if / then / else” over prose paragraphs the model can skim past.
- **Define “good” with examples** — Show a minimal acceptable vs. insufficient snippet (comment text, JSON skeleton, plan section). Contrast beats adjectives every time.
- **Address edge cases** — Empty inputs, conflicting instructions, secrets in context, “do everything” requests, and when to refuse or escalate. If humans disagree on policy, tell the agent which side wins or to surface the conflict explicitly.
- **Give a mental model** — One sentence the agent can reuse (“Every issue is a contract…”) plus invariants (“I never write production code”) sticks better than a flat rule dump. The model will interpolate; give it a compass.

Reuse verbs and nouns across **Process** and **Output contract** so the same concepts appear in the same words (e.g. always say `needs-info`, not sometimes “blocked”).

### Output contract: concrete pattern

When automation parses results, specify:

- A fixed set of top-level headings or a single JSON object schema.
- Required vs optional fields and allowed enum values.
- What to emit when information is missing (`null`, `"unknown"`, or an explicit questions block—pick one and stick to it).
- Whether the human summary duplicates machine fields or only references them.

A lightweight pattern is: a **Summary** section for humans, then a **Result (machine)** section containing a single JSON object, for example:

```json
{
  "status": "ready|needs-info|reject",
  "labels": [],
  "confidence": 0.0
}
```

Document in the agent body that those headings and keys are mandatory so parsers and humans agree on the same surface.

## Registering your agent

1. **Create the plugin directory**: `plugins/{name}/` (use a stable id; convention here is often `{role}-agent`).

2. **Add the agent file**: `plugins/{name}/agents/{name}.md` (filename may differ; paths must match `plugin.json`).

3. **Create `plugin.json`**: `plugins/{name}/.claude-plugin/plugin.json` with `name`, `version`, `description`, `agents` (array of paths like `agents/foo.md`), and `skills` (array of paths under the plugin root, typically `skills/...`). Example shape:

```json
{
  "name": "my-agent",
  "version": "1.0.0",
  "description": "Short summary for marketplace and UIs",
  "agents": ["agents/my-agent.md"],
  "skills": ["skills/some-skill"]
}
```

4. **Register in the marketplace**: Add an entry to `.claude-plugin/marketplace.json` under `plugins` with `name`, `source` (directory name under `plugins/`), `description`, `version`, and metadata (`category`, `tags`, `keywords`) consistent with sibling entries.

5. **Symlink required skills**: For each path in `plugin.json`’s `skills` array, add a **directory symlink** under `plugins/{name}/skills/` pointing at the canonical skill under `_skills/`, e.g. `plugins/my-agent/skills/security-awareness -> ../../../_skills/security-awareness`. CI resolves symlinks and requires each target to exist as a directory.

Keep agent files **substantive**: validation enforces a minimum size (currently greater than 1024 bytes) and valid frontmatter with a non-empty `name`. Thin stubs fail `ci_validate.py agents` on purpose—expand sections until the file is genuinely instructive.

## Testing your agent

**Locally (Claude Code)** — Install or develop against this marketplace, load your plugin, and invoke the agent with realistic prompts (minimal input, noisy input, adversarial input). Confirm the **output contract** is stable across runs: same headings, same field names, same order. Regression-test after edits to **Core principles** or **Process**, since small wording changes can shift behavior.

**In CI** — The repo runs `python3 scripts/ci_validate.py plugins` and `python3 scripts/ci_validate.py agents` (see `.github/workflows/ci.yml`). These checks cover `marketplace.json`, each `plugin.json`, agent file presence, frontmatter, minimum content size, skill paths (after symlink resolution), and broken symlinks under `plugins/`. Fix all reported errors before merging.

Optional deeper checks: run the full CI job locally if you change validation scripts; grep workflows under `.github/workflows/` for `agent-role` or agent names to ensure new agents are discoverable if you intend to call them from automation.

Workflows such as `ai-implement.yml` and `ai-pr-review.yml` load agents by role via the local `run-ai-agent` action; aligning `name` / role with how workflows reference the agent avoids integration surprises.

**Manual validation commands** (from repository root):

```bash
python3 scripts/ci_validate.py marketplace
python3 scripts/ci_validate.py plugins
python3 scripts/ci_validate.py agents
python3 scripts/ci_validate.py symlinks
```

Run `plugins` and `agents` after any change to marketplace entries, `plugin.json`, agent bodies, or skill symlinks.

## Example

The **triage** agent is the reference implementation: clear **Role** and **Mindset**, ten **Core Principles**, a detailed **Triage Process**, per-type **Assessment Criteria**, **Anti-patterns**, and an explicit **Output Contract** with structured comment and JSON. It demonstrates how to separate human-readable explanation from bot-consumable structure—downstream agents depend on that split.

Study these paths together:

- `plugins/triage-agent/agents/triage.md` — full instruction layout and depth.
- `plugins/triage-agent/.claude-plugin/plugin.json` — `agents` and empty `skills` for a dependency-free agent.
- `.claude-plugin/marketplace.json` — the `triage-agent` entry’s `source` and metadata.

When your agent needs shared guidance, mirror **implementer-agent** or **pr-reviewer**: multiple skills in `plugin.json`, each mirrored under `plugins/.../skills/` as symlinks into `_skills/`.

---

### Checklist before opening a PR

- [ ] Agent file exceeds minimum size and passes frontmatter checks.
- [ ] `plugin.json` paths match real files; skill symlinks resolve into `_skills/`.
- [ ] `marketplace.json` entry includes required fields and correct `source`.
- [ ] **Output contract** tested with at least three prompts (happy path, thin input, contradictory input).
- [ ] If wired into workflows, `agent-role` and docs reference the same agent identity.
