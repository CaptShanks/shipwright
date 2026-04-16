# Shipwright

AI agents, skills, and MCPs for automated software development. A universal plugin framework that works as a **Claude Code plugin marketplace**, a **GitHub Actions CI/CD automation source**, and installs natively into **Cursor**, **Codex**, and **Claude Code** via the `ship` CLI.

[![Browse Plugins](https://img.shields.io/badge/browse-plugin%20catalog-blue)](https://captshanks.github.io/shipwright/)
[![Release](https://img.shields.io/github/v/release/CaptShanks/shipwright)](https://github.com/CaptShanks/shipwright/releases)
[![License: MIT](https://img.shields.io/badge/license-MIT-green)](LICENSE)

## What is Shipwright?

Shipwright provides a catalog of specialized AI agents, reusable skills, and MCP integrations that automate the full software development lifecycle -- from issue triage to production release:

- **Triage** issues to determine actionability and request clarification
- **Architect** solutions with tradeoff analysis and Mermaid diagrams
- **Implement** changes following project conventions and security best practices
- **Write tests** with comprehensive edge case and integration coverage
- **Security review** code against OWASP Top 10 and threat models
- **Review PRs** for quality, correctness, and adherence to standards
- **Generate changelogs** with semantic versioning and audience-aware release notes
- **Analyze codebases** by tracing runtime paths, dependency graphs, and architectural boundaries

Each agent is deeply opinionated about how to do its job well -- not just "write code," but "write code with security awareness, proper naming, error handling, and scalability in mind." Agents bundle domain-specific skills so they carry their expertise with them.

## Three Ways to Use Shipwright

### Path 1: Ship CLI (Universal)

Install agents, skills, and MCPs into any supported AI tool (Cursor, Claude Code, Codex). The CLI writes to each tool's native directory structure:

| Tool | Agents | Skills |
|------|--------|--------|
| Claude Code | `.claude/agents/*.md` | `.claude/skills/<name>/SKILL.md` |
| Codex | `.codex/agents/*.toml` | `.agents/skills/<name>/SKILL.md` |
| Cursor | `.cursor/agents/*.md` | `.cursor/skills/<name>/SKILL.md` |

```bash
# Install the CLI
go install github.com/CaptShanks/shipwright/cli/cmd/ship@latest

# Install a plugin locally (current project, all tools)
ship install architect-agent

# Install for a specific tool
ship install architect-agent --target cursor

# Install globally
ship install shipwright-full --global

# Browse the marketplace
ship search
ship info pr-reviewer

# Manage installed plugins
ship list
ship uninstall triage-agent
ship update

# Manage MCP servers
ship mcp list
ship mcp install context7
ship mcp install mcp-atlassian --target cursor --global
ship mcp remove context7
```

### Path 2: Claude Code Plugin (Local Development)

Install agents and skills directly into Claude Code:

```bash
# Add the marketplace (one time)
/plugin marketplace add CaptShanks/shipwright

# Install individual agents (skills are bundled automatically)
/plugin install security-reviewer@shipwright
/plugin install implementer-agent@shipwright

# Or install everything
/plugin install shipwright-full@shipwright
```

### Path 3: GitHub Actions (CI/CD Automation)

Use reusable workflows to automate the issue-to-PR pipeline. Supports both Claude Code (via OAuth) and Codex (via API key):

```yaml
# .github/workflows/ai-triage.yml
name: AI Triage
on:
  issues:
    types: [opened]

jobs:
  triage:
    uses: CaptShanks/shipwright/.github/workflows/ai-triage.yml@main
    with:
      ai-provider: claude
      project-context-path: .github/shipwright/project-context.md
    secrets:
      CLAUDE_CODE_OAUTH_TOKEN: ${{ secrets.CLAUDE_CODE_OAUTH_TOKEN }}
```

Reusable workflows available: `ai-triage.yml`, `ai-implement.yml`, `ai-pr-review.yml`.

## Available Plugins

### Agents (8)

| Plugin | Description | Bundled Skills |
|--------|-------------|----------------|
| `triage-agent` | Analyzes issues for actionability, requests clarification, classifies for downstream agents | -- |
| `architect-agent` | Designs solution architecture with tradeoff analysis and implementation plans | security-awareness, scalability-resilience |
| `implementer-agent` | Implements solutions following architecture plans and project conventions | security-awareness, scalability-resilience, code-quality-fundamentals |
| `test-engineer` | Writes unit and integration tests with edge case coverage | code-quality-fundamentals, test-patterns |
| `security-reviewer` | Reviews code for vulnerabilities using OWASP knowledge and threat modeling | security-awareness, security-owasp |
| `pr-reviewer` | Reviews PRs for quality, correctness, and security with actionable feedback | security-awareness, code-quality-fundamentals, code-review-standards |
| `changelog-generator` | Produces semantically versioned, audience-aware changelogs from git history and PRs | -- |
| `codebase-analyzer` | Produces structured analyses of runtime paths, architectural boundaries, and dependency graphs | -- |

### Skills (10)

| Skill | Description |
|-------|-------------|
| `security-awareness` | Input validation, secret handling, least privilege, and defensive coding |
| `code-quality-fundamentals` | Naming conventions, SOLID principles, error handling, and code organization |
| `scalability-resilience` | Graceful degradation, idempotency, retry patterns, and failure domain analysis |
| `security-owasp` | OWASP Top 10 review guidance with concrete code patterns and severity bands |
| `test-patterns` | Testing strategies, mocking, fixture management, and the testing pyramid |
| `code-review-standards` | Code review methodology, severity classification, and actionable feedback |
| `semantic-versioning` | Breaking change detection, semver decision rules, and version bump classification |
| `architecture-patterns` | Architectural style identification, practiced vs aspirational architecture analysis |
| `go-skills` | Go development idioms, patterns, and conventions |
| `terraform-development` | Terraform/HCL module architecture, state management, provider patterns, and IaC security |

Skills are standalone and usable by any agent or independently. Use the `additional-skills` input in CI/CD workflows to inject language/framework-specific skills, or add them to an agent's frontmatter for local development.

### MCPs (4)

| MCP | Description | Auth Required |
|-----|-------------|---------------|
| `context7` | Up-to-date library and framework documentation | None |
| `serena` | Semantic code navigation and symbol-level editing | None |
| `mcp-atlassian` | Jira + Confluence integration | 5 env vars |
| `bitbucket-cloud` | Bitbucket Cloud PR and repo management | 3 env vars |

### Bundle

| Plugin | Contents |
|--------|----------|
| `shipwright-full` | All 8 agents + all 10 skills + all 4 MCPs |

## Integrating with Your Repo

1. Create `.github/shipwright/project-context.md` describing your project (see [Knowledge Base Format](docs/knowledge-base-format.md))
2. Add thin caller workflows (see [Getting Started](docs/getting-started.md))
3. Optionally activate language/framework skills via `additional-skills` in workflows or agent frontmatter locally

## Creating Custom Agents and Skills

- [Creating Agents](docs/creating-agents.md) -- agent frontmatter, persona structure, and skill dependencies
- [Creating Skills](docs/creating-skills.md) -- follows the [Agent Skills specification](https://agentskills.io/specification)

## License

[MIT](LICENSE)
