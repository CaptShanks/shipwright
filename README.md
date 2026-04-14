# Shipwright

AI agents and skills for automated software development. A dual-purpose framework that works as a **Claude Code plugin marketplace** and a **GitHub Actions reusable workflow source**.

## What is Shipwright?

Shipwright provides a catalog of specialized AI agents and reusable skills that automate the software development lifecycle:

- **Triage** issues to determine if they're actionable
- **Architect** solutions by exploring the codebase
- **Implement** changes following project conventions
- **Write tests** with comprehensive edge case coverage
- **Security review** code for vulnerabilities
- **Review PRs** for quality and correctness

Each agent is deeply opinionated about how to do its job well -- not just "write code," but "write code with security awareness, proper naming, error handling, and scalability in mind."

## Three Ways to Use Shipwright

### Path 1: Ship CLI (Universal)

Install agents, skills, and MCPs into any supported AI tool (Cursor, Claude Code, VS Code, Codex):

```shell
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

```shell
# Add the marketplace (one time)
/plugin marketplace add CaptShanks/shipwright

# Install individual agents (skills are bundled automatically)
/plugin install security-reviewer@shipwright
/plugin install implementer-agent@shipwright

# Or install everything
/plugin install shipwright-full@shipwright
```

### Path 3: GitHub Actions (CI/CD Automation)

Use reusable workflows to automate the issue-to-PR pipeline:

```yaml
# In your repo's .github/workflows/ai-triage.yml
name: AI Triage
on:
  issues:
    types: [opened]

jobs:
  triage:
    uses: CaptShanks/shipwright/.github/workflows/ai-triage.yml@v1
    with:
      ai-provider: codex
      project-context-path: .github/shipwright/project-context.md
    secrets:
      AI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

## Available Plugins

| Plugin | Type | Contents |
|--------|------|----------|
| `triage-agent` | Agent | Issue triage and analysis |
| `architect-agent` | Agent + Skills | Solution design (+ security-awareness, scalability-resilience) |
| `implementer-agent` | Agent + Skills | Code implementation (+ security-awareness, scalability-resilience, code-quality-fundamentals) |
| `test-engineer` | Agent + Skills | Test writing (+ code-quality-fundamentals, test-patterns) |
| `security-reviewer` | Agent + Skills | Security review (+ security-awareness, security-owasp) |
| `pr-reviewer` | Agent + Skills | PR review (+ security-awareness, code-quality-fundamentals, code-review-standards) |
| `go-skills` | Skill | Go development idioms and patterns |
| `shipwright-full` | Bundle | All 6 agents + all 7 skills + all 4 MCPs |

### Available MCPs

| MCP | Description | Env Vars |
|-----|-------------|----------|
| `context7` | Library and framework documentation via Context7 | None |
| `serena` | Semantic code navigation and symbol-level editing | None |
| `mcp-atlassian` | Jira + Confluence integration | 5 required |
| `bitbucket-cloud` | Bitbucket Cloud PR and repo management | 3 required |

## Integrating with Your Repo

Add a knowledge base file so the agents understand your project:

1. Create `.github/shipwright/project-context.md` describing your project (see [Knowledge Base Format](docs/knowledge-base-format.md))
2. Add thin caller workflows (see [Getting Started](docs/getting-started.md))
3. Optionally add `.claude/settings.json` to auto-register the marketplace for Claude Code users

## Creating Custom Agents and Skills

- [Creating Agents](docs/creating-agents.md)
- [Creating Skills](docs/creating-skills.md)

## License

[MIT](LICENSE)
