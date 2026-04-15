# Getting started with Shipwright

Shipwright provides **Claude Code plugins** (agents and skills for local development) and **GitHub Actions reusable workflows** (triage, implementation, and PR review in CI). This guide walks through both paths.

---

## Path 1: Claude Code plugin (local development)

Use this path when you want Shipwright’s agents and skills inside [Claude Code](https://docs.anthropic.com/en/docs/claude-code) on your machine.

### 1. Prerequisites

- **Claude Code** installed and working with your Anthropic account.

### 2. Add the marketplace

In Claude Code, register the Shipwright marketplace (one-time per environment):

```text
/plugin marketplace add CaptShanks/shipwright
```

### 3. Browse available plugins

List everything the marketplace exposes:

```text
/plugin marketplace list
```

You should see agent plugins (for example `triage-agent`, `architect-agent`, `implementer-agent`, `test-engineer`, `security-reviewer`, `pr-reviewer`), skill-only bundles (for example `go-skills`), and the full bundle `shipwright-full`.

### 4. Install individual plugins

Install only what you need by name, scoped to the `shipwright` marketplace:

```text
/plugin install security-reviewer@shipwright
```

Repeat for other plugins (for example `triage-agent@shipwright`, `pr-reviewer@shipwright`).

### 5. Install the full bundle

To pull in **all six agents and all bundled skills** in one step:

```text
/plugin install shipwright-full@shipwright
```

### 6. How agents and skills become available

After installation, Claude Code loads the plugin’s **agent definitions** and **skills** according to the plugin manifest (for example `shipwright-full` ships every agent under `agents/` and every skill under `skills/`). In practice:

- **Agents** show up as invokable roles or subagent-style capabilities aligned with each plugin’s `agents/*.md` definitions (triage, architect, implementer, test-engineer, security-reviewer, pr-reviewer).
- **Skills** are available as reference material and checklists the model can apply when relevant (security, code quality, testing patterns, Go idioms, and so on).

If something does not appear after install, confirm the marketplace add succeeded, run `/plugin marketplace list` again, and restart Claude Code.

---

## Path 2: GitHub Actions (CI/CD automation)

Use this path to run Shipwright **in GitHub Actions** against your repository: issue triage on open/reopen, manual implementation runs that open PRs, and automated PR review comments.

### Prerequisites

| Requirement | Why it matters |
|-------------|----------------|
| **GitHub repository** | Workflows run in your repo and need `contents`, `issues`, and/or `pull-requests` access. |
| **`OPENAI_API_KEY` repository secret** | Reusable workflows expect a secret named `AI_API_KEY` **passed from** your repo; storing the key as `OPENAI_API_KEY` and mapping it in the caller is a common convention. The default provider is **Codex** (`ai-provider: codex`), which uses the OpenAI API. |
| **`dev` branch** | The **implement** workflow defaults `target-branch` to `dev` when opening PRs. Create `dev` (or override `target-branch` in the caller). |

> **Secret wiring:** Shipwright’s reusable workflows declare `secrets: AI_API_KEY`. In each caller below, pass `AI_API_KEY: ${{ secrets.OPENAI_API_KEY }}` (or map from another secret name you prefer).

### Step 1: Project knowledge base

Create a concise knowledge file so agents understand your stack, conventions, and boundaries:

**Path:** `.github/shipwright/project-context.md`

Follow the structure and sections described in **[knowledge base format](./knowledge-base-format.md)**. The workflows pass this path into the runner; if the file is missing, triage and PR review **continue without** injected context (you will see a notice in logs).

### Step 2: GitHub issue templates (optional, recommended)

Add issue templates under `.github/ISSUE_TEMPLATE/` so reporters supply repro steps, expected behavior, and scope. Better issue bodies improve triage and downstream implementer prompts.

### Step 3: Caller workflows

Add **thin callers** in your repository that reference pinned versions of Shipwright’s reusable workflows (example uses `@v1`—adjust to the tag or SHA you trust).

#### `.github/workflows/ai-triage.yml`

Runs when an issue is opened or reopened; posts triage comments and applies labels.

```yaml
name: AI issue triage

on:
  issues:
    types: [opened, reopened]

permissions:
  issues: write
  contents: read

concurrency:
  group: shipwright-triage-${{ github.event.issue.number }}
  cancel-in-progress: true

jobs:
  triage:
    uses: CaptShanks/shipwright/.github/workflows/ai-triage.yml@v1
    with:
      ai-provider: codex
      project-context-path: .github/shipwright/project-context.md
    secrets:
      AI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

#### `.github/workflows/ai-implement.yml`

Manual workflow: you provide the issue number; Shipwright runs architect → implementer → test-engineer and opens a PR against `dev`.

```yaml
name: AI implement

on:
  workflow_dispatch:
    inputs:
      issue_number:
        description: GitHub issue number to implement
        required: true
        type: string
      target_branch:
        description: Base branch for the pull request
        required: false
        default: dev

permissions:
  contents: write
  pull-requests: write
  issues: write

concurrency:
  group: shipwright-implement-${{ github.event.inputs.issue_number }}
  cancel-in-progress: false

jobs:
  implement:
    uses: CaptShanks/shipwright/.github/workflows/ai-implement.yml@v1
    with:
      issue-number: ${{ github.event.inputs.issue_number }}
      target-branch: ${{ github.event.inputs.target_branch }}
      ai-provider: codex
      project-context-path: .github/shipwright/project-context.md
    secrets:
      AI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

#### `.github/workflows/ai-pr-review.yml`

Runs on pull request activity; posts security review, code review, and a summary comment.

```yaml
name: AI PR review

on:
  pull_request:
    types: [opened, synchronize, reopened]

permissions:
  contents: read
  pull-requests: write
  issues: write

concurrency:
  group: shipwright-pr-review-${{ github.event.pull_request.number }}
  cancel-in-progress: true

jobs:
  review:
    uses: CaptShanks/shipwright/.github/workflows/ai-pr-review.yml@v1
    with:
      ai-provider: codex
      project-context-path: .github/shipwright/project-context.md
    secrets:
      AI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

### Step 4: Smoke test

1. Merge the three caller workflows and `.github/shipwright/project-context.md` to your default branch (and ensure `dev` exists if you use the default base branch).
2. Under **Settings → Secrets and variables → Actions**, add `OPENAI_API_KEY`.
3. Open a **sample issue** with a clear title and body (for example a small, well-scoped bug or doc fix).
4. Confirm the **AI issue triage** workflow ran and left a comment on the issue.
5. (Optional) Run **AI implement** manually with that issue’s number and verify a PR appears against `dev`.
6. (Optional) Open or update a PR and confirm **AI PR review** comments appear.

---

## Configuration

### Cost management

- **Triage on `opened` only** (or add `reopened` deliberately)—each run invokes the model once per issue event.
- **Implement is the highest spend** (architect + implementer + test-engineer). Keep it **manual** (`workflow_dispatch`) until you trust prompts and context.
- **PR review** runs per configured `pull_request` activity; narrow triggers with `paths` / `paths-ignore` if reviews should skip docs-only churn.
- **Pin** `uses: ...@v1` (or a full SHA) so behavior and costs stay predictable across upstream changes.
- Maintain a **focused** `project-context.md` so agents need fewer clarification rounds and less repeated reasoning.

### Concurrency controls

Use `concurrency` in **caller** workflows (as in the examples) so duplicate events for the same issue or PR do not stack expensive runs. Choose `cancel-in-progress: true` when only the latest state matters (triage, PR review); use `false` for long implement runs if you prefer not to abort mid-flight.

### Activating language/framework skills

Shipwright agents are language-agnostic by design. Language-specific knowledge (Go idioms, Terraform patterns, etc.) lives in standalone skills that you activate per-project.

**In CI/CD workflows** — pass skill names via the `additional-skills` input on the `run-ai-agent` composite action:

```yaml
- uses: CaptShanks/shipwright/actions/run-ai-agent@main
  with:
    agent-role: implementer
    additional-skills: "go-development"
    # ...
```

Multiple skills are comma-separated: `additional-skills: "go-development,terraform-development"`. Each name resolves to `plugins/<name>/skills/<name>/SKILL.md` in the Shipwright repository. The skill content is injected between the agent definition and the project context in the prompt.

**In local Claude Code** — install the skill plugin, then add it to the agent's frontmatter:

```text
# Install the skill
/plugin install go-skills@shipwright
```

Then edit the agent's `.md` file (inside your installed plugin directory) and add the skill name to the `skills:` array:

```yaml
skills:
  - security-awareness
  - scalability-resilience
  - code-quality-fundamentals
  - go-development          # added by you
```

Claude Code's progressive disclosure engine will surface the skill content when relevant to the conversation.

### Customizing agent behavior

Edit `.github/shipwright/project-context.md` to describe:

- Tech stack, build/test commands, and folder layout
- Coding standards, error-handling patterns, and review expectations
- Security constraints (secrets, auth, PII) and out-of-scope areas

The composite action merges this file into the prompt for each agent step when the path is set and the file exists.

---

## Troubleshooting

| Symptom | Likely cause | What to do |
|--------|----------------|------------|
| Workflow fails: missing `github.event.issue` | Caller not triggered from an `issues` event | Use `on: issues:` (for example `opened`) for `ai-triage.yml`, not `workflow_dispatch`. |
| Triage succeeds but labels missing | Label permissions or API errors | Confirm job `permissions: issues: write` and check the Actions log for label step warnings. |
| Implement cannot open PR | Missing `dev` or wrong base branch | Create `dev` or set `target_branch` input / `target-branch` default to your real integration branch. |
| `AI_API_KEY` errors | Secret not passed to reusable workflow | Map `secrets.AI_API_KEY` in the **caller** to `secrets.OPENAI_API_KEY` (or your chosen secret name). |
| “Continuing without project context” | Path wrong or file missing | Ensure `.github/shipwright/project-context.md` exists or pass a valid `project-context-path`. |
| PR review never runs | Event filter or permissions | Check `on: pull_request` types and that the workflow file is on the default branch (for first-time setup). |
| Claude provider in Actions | `ai-provider: claude` not fully implemented | Keep `ai-provider: codex` until Claude support is documented for your org. |

For reproducible behavior, pin the reusable workflow ref (`@v1` or SHA), keep `project-context.md` under version control, and treat Shipwright output as **assistance**: review PRs and issue comments before merging or acting on them.
