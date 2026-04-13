# Project knowledge base format (`project-context.md`)

Consumer repositories can provide a **project-specific knowledge base** at `.github/shipwright/project-context.md`. This page explains how to write it so AI agents understand your codebase, constraints, and workflows.

## Purpose

**Why it matters:** Generic agents only know public patterns and open files. A project context encodes *your* stack, layout, conventions, and non-obvious decisions so agents match your style, use the right test/build/deploy commands, and avoid repeating mistakes. Write it like onboarding for a **senior engineer new to the repo**.

**How it is used:** The file is **injected into every agent prompt** alongside the agent definition (role, tools, workflow). Treat it as authoritative for this repository unless the user explicitly overrides it in chat.

## Required sections

Keep sections **scannable** (bullets over prose). Prefer **stable facts**; link to ADRs or docs when detail would bloat the file.

### 1. Project overview

**Include:** Name; problem solved; languages/frameworks; architecture summary (monolith, CLI, services, etc.).

**Example:**

```markdown
**Name:** Acme Widget API · **Purpose:** REST API for widgets + billing webhooks  
**Stack:** Go 1.22, PostgreSQL 15, chi, sqlc · **Architecture:** One binary; jobs via LISTEN/NOTIFY worker
```

### 2. Repository structure

**Include:** Important directories; what belongs where (features vs infra).

**Example:** `- cmd/server, cmd/worker · internal/ app-only · pkg/ shared · migrations/ SQL · deploy/ Helm`

### 3. Development standards

**Include:** Format/lint; naming; errors; logging; doc expectations for public APIs.

**Example:** `golangci-lint per .golangci.yml · wrap errors with context · no panic in prod paths · godoc on exports`

### 4. Dependencies

**Include:** Runtime and key dev deps with **versions** when it matters; note non-obvious choices.

**Example:** `Go 1.22+ · PG 15 · chi/v5, pgx/v5 · sqlc (codegen, dev only)`

### 5. Testing

**Include:** Framework; unit vs integration; fixtures/tags; CI location; coverage or mocking rules.

**Example:** `go test ./... · //go:build integration · testcontainers locally · CI: .github/workflows/ci.yml · mocks behind internal/ports`

### 6. Build & run

**Include:** Local commands; env/secret **names** (not values); artifacts; deploy path.

**Example:** `make run (.env from .env.example) · make build → bin/server · docker compose · main → Actions → ECR → Helm`

### 7. Architecture decisions

**Include:** ADRs, package boundaries, data flow, invariants agents must not break.

**Example:** `ADR-003: webhook idempotency on event_id · handlers → ports interfaces · adapters implement`

---

## Optional sections

| Section | What to cover |
|--------|----------------|
| **8. Security considerations** | Auth model, secret storage (Vault/env), PII, compliance—**never paste secrets** |
| **9. Known issues & tech debt** | Fragile code, flaky tests, legacy areas, “do not refactor yet” |
| **10. PR guidelines** | Review focus, branch names (`feature/`, `fix/`), required checks, changelog rules |

## Template

Copy into `.github/shipwright/project-context.md` and fill in.

```markdown
# Project context

## Project overview
**Name:** … **Purpose:** … **Tech stack:** … **Architecture:**

## Repository structure
- …

## Development standards
- …

## Dependencies
- …

## Testing
- …

## Build & run
- …

## Architecture decisions
- …

## Security considerations (optional)
- …

## Known issues & tech debt (optional)
- …

## PR guidelines (optional)
- …
```

## Example (Go CLI — Bubble Tea TUI for Terraform)

Fictional “Terraprism”-style project: a Bubble Tea TUI wrapping the Terraform CLI.

```markdown
# Project context

## Project overview
**Name:** Terraprism · **Purpose:** TUI to browse/apply Terraform plans safely  
**Stack:** Go 1.22+, Bubble Tea / Bubbles / Lip Gloss; shells out to `terraform`  
**Architecture:** `cmd/` entry · `internal/tui` model/update/view · `internal/terraform` subprocess + parse (timeouts, structured errors)

## Repository structure
`cmd/terraprism/` CLI · `internal/tui/` screens & keys · `internal/terraform/` runner · `internal/config/` · `docs/` user docs (agents: don’t duplicate long prose here)

## Development standards
gofmt, vet, golangci-lint · immutable TUI updates; no cross-model globals · user errors in view layer; wrap exec failures with command context · `snake_case` JSON in config

## Dependencies
Go 1.22+ · charm libs (see `go.mod`) · `terraform` on PATH per README matrix

## Testing
`go test ./...` · table tests for parsing/config · fake runner for terraform package · avoid brittle full TUI snapshots unless isolated

## Build & run
`go run ./cmd/terraprism` or `make build` · releases via goreleaser if present · binaries on GitHub Releases (no server deploy)

## Architecture decisions
Apply/plan path: **always subprocess to Terraform CLI** (no embedded HCL for apply) · **context cancellation** on all long work; clean Ctrl+C · **no cloud secrets in repo**—use normal Terraform backend auth

## Security considerations (optional)
Redact plan output that may contain secrets · subprocess: fixed argv; avoid shell expansion on user paths

## Known issues & tech debt (optional)
Windows resize is best-effort; validate on macOS/Linux first · large plan JSON may need streaming later

## PR guidelines (optional)
Branches `feature/`, `fix/` · UX PRs: screenshot or asciinema · non-trivial logic: paste test output in description
```

## Tips

- **Concise but comprehensive:** Bullets and links beat pasting APIs.  
- **Update with the code:** Change this file when stack, commands, or ADRs change.  
- **Target a productive senior:** Commands, boundaries, and gotchas matter more than philosophy.  
- **Stable facts:** Point to tickets/ADRs for volatile detail.
