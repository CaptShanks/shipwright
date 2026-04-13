---
name: code-review-standards
description: >-
  Code review methodology, severity classification, and standards for writing
  actionable, constructive feedback. Use when reviewing pull requests, providing
  code feedback, or establishing review processes for a team.
license: MIT
metadata:
  author: CaptShanks
  version: "1.0.0"
---

# Code Review Standards

Code review is a shared engineering practice: it improves quality, spreads knowledge, and catches defects before they reach production. This skill defines how to review constructively—prioritizing signal over noise, clarity over ego, and outcomes over ceremony.

## Mindset: Collaboration, Not Gatekeeping

Treat the pull request as a **joint product** between author and reviewers. Your role is to help the change land safely and maintainably, not to prove superiority or block progress.

- **Assume good intent.** Ambiguous code may reflect time pressure, incomplete context, or an honest mistake—not carelessness.
- **Review the code, not the person.** Use neutral language: “this branch could race” rather than “you introduced a race.”
- **Default to curiosity.** Ask “what happens if X?” before asserting “this is wrong.” You may be missing constraints the author had.
- **Celebrate strong choices.** Brief praise for a clean abstraction or thorough tests reinforces the behaviors you want repeated.
- **Escalate scope deliberately.** If the change reveals a larger design problem, distinguish “must fix for this PR” from “follow-up ticket” so the author can ship incrementally.

Gatekeeping signals to avoid: dismissive tone, unexplained “no” votes, demands for personal style preferences framed as universal rules, or reopening settled team decisions without new information.

## What Makes Good Review Feedback

Effective comments share several properties:

### Specific and anchored

Tie feedback to **observable behavior** in the diff: file, function, or line when possible. Vague comments (“this feels off”) force the author to guess your concern.

**Strong:** “If `ctx` is cancelled mid-batch, `processItem` may leave `state.partial` inconsistent with the DB. Consider deferring a rollback or documenting that callers must reconcile.”

**Weak:** “Error handling seems sketchy.”

### Actionable

The author should know **what to do next**: change the logic, add a test, extract a helper, or document an invariant. If multiple fixes are valid, name one acceptable path or offer a short tradeoff framing.

### Explains the “why”

Standards and taste matter, but **reasoning** teaches the team. Link feedback to risk (correctness, security, operability), maintainability, or agreed conventions—not “I wouldn’t do it that way” without justification.

### Appropriately scoped

Match comment depth to severity. A typo in a log message needs one line; a new public API may warrant a short design note. Avoid essay-length threads on nitpicks.

### Respectful of the author’s time

Batch related points. If you have five nits in one function, one comment with a numbered list often beats five separate notifications.

## Severity Classification

Use a **consistent severity model** so authors can triage and teams can measure review health. Align labels with your tracker or merge rules if you have them; the model below maps cleanly to most workflows.

| Level | Name | Meaning | Typical merge impact |
|-------|------|---------|----------------------|
| **S0** | **Blocker** | Correctness bug, security flaw, data loss risk, broken contract, or violation of mandatory policy. | Must fix before merge (or revert if already merged). |
| **S1** | **Major** | Significant maintainability or reliability risk; missing tests for critical paths; unclear API that will cause misuse. | Should fix before merge unless explicitly waived with documented risk acceptance. |
| **S2** | **Minor** | Style within team guidelines, small refactors, naming, local clarity. | Fix before merge if quick; otherwise acceptable as follow-up if team allows. |
| **S3** | **Nit / suggestion** | Preference, optional polish, hypothetical future improvement. | Never blocking; author may decline without discussion. |

### Blocker (S0) — use sparingly

Reserve blockers for issues where merging would likely cause **harm or irreversible cost**: wrong authorization check, SQL injection, secret in repo, race that corrupts data, breaking change to a stable API without versioning, production config that points at the wrong cluster.

### Major (S1) — substance without emergency

Examples: error paths untested for a payment flow, new dependency with unclear license, public function with misleading name that will confuse callers, missing migration for a schema change that will fail deploys.

### Minor (S2) — team quality bar

Examples: function too long per team guide, duplicated logic that should be extracted, log level inconsistent with observability standards.

### Nit (S3) — optional

Examples: alternate naming you slightly prefer, micro-style that the linter does not enforce, “consider X someday.”

**Prefix comments in text** when tools do not encode severity: `[Blocker]`, `[Major]`, `[Minor]`, `[Nit]` so scanning the PR is fast.

## Blocking Issues vs. Suggestions

**Blocking feedback** must satisfy at least one of:

1. **Objective harm** if merged (security, correctness, compliance, SLO risk).
2. **Explicit team policy** the change violates (e.g., no new `any`, required tests for parsers).
3. **Contract breakage** without migration path (API, event schema, storage format).

**Non-blocking suggestions** include refactors that improve readability but do not fix a defect, hypothetical edge cases with negligible probability, or personal style outside documented standards.

When you are unsure, **ask a clarifying question** first; upgrade to blocking only after you confirm impact. When you block, **state the failure mode**: “If Y happens, Z breaks because…”

Offer **escape hatches** when appropriate: “Blocking: missing input validation on `email`. Non-blocking: consider extracting validation to shared helper in a follow-up.”

## Time Management in Reviews

Sustainable review keeps latency low without sacrificing depth.

### Budget time by change risk

- **Small, localized changes** (docs, typos, isolated bugfix): quick pass—goal within minutes.
- **Medium features**: read description and tests first, then code; aim for first response same day if team SLA allows.
- **Large or cross-cutting changes**: skim architecture/description first; consider reviewing in layers (API → core logic → tests) or asking the author to split if the diff is unreviewable.

### Triage the diff

1. Read the PR title, description, and linked issue/ticket.
2. Scan file list and generated vs. hand-written changes.
3. Read tests and public API or config changes before implementation minutiae—tests encode intent.
4. Deep-dive only in hot spots: auth, concurrency, persistence, parsers, crypto, migrations, feature flags.

### Avoid perfection spirals

Not every PR must be “ideal.” **Good enough** that meets the bar and is safe to merge is the goal. Reserve deep redesign discussions for design docs or follow-up work.

### When to take it offline

Thread explosion on fundamental design usually means **insufficient upfront alignment**. Suggest a short sync or RFC reference rather than debating architecture in twenty inline comments.

### Reviewer load

If you are at capacity, **react early** (“I can review tomorrow morning”) so the author can reroute. Rotate reviews to avoid single points of failure.

## Review Anti-Patterns

| Anti-pattern | Why it hurts | Better approach |
|--------------|--------------|-----------------|
| **Rubber stamping** | Defeats the purpose of review; misses real defects. | Minimum: trace happy path, scan error handling, check tests exist for new behavior. |
| **Nitpicking without blockers** | Noise drowns signal; authors dismiss all feedback. | Batch nits; mark as `[Nit]`; use automated formatters/linters for style. |
| **Ambiguous LGTM** | “Looks fine I guess” leaves uncertainty. | Explicit: “No blockers; optional: consider extracting X.” |
| **Design debate in comments** | Slow, lossy, and demoralizing for large disagreements. | Move to doc/sync; return with agreed direction. |
| **Scope creep** | “While you’re here, rewrite the module.” | Separate PR or ticket unless tightly coupled. |
| **Commenting on generated code** | Wastes everyone’s time. | Exclude or collapse generated paths; comment on generator or config instead. |
| **Demanding rewrites for taste** | Blocks shipping without improving objective quality. | Suggest as non-blocking or document team standard first. |
| **Inconsistent standards** | Authors cannot predict what will be challenged. | Point to linter rules, style guide, or prior examples. |
| **Ghosting** | Long silence blocks the team. | Acknowledge receipt; give ETA or delegate. |
| **Approval with unstated reservations** | Surprises in production or later PRs. | Speak up or abstain with reason. |

## Pull Request Description Expectations (For Reviewers)

When reviewing, you may ask the author to clarify **intent and verification** if missing:

- **What** changed and **why** (problem statement).
- **How to test** (commands, feature flags, environments).
- **Risk notes** (rollout, migrations, backwards compatibility).
- **Screenshots or samples** for user-visible changes.

That context reduces guesswork and speeds up first-pass review.

## Using the Checklist

Use `_skills/code-review-standards/checklists/review-checklist.md` as a **structured pass** over important dimensions (naming, errors, tests, design, security, etc.). It is a menu, not a mandate: apply sections relevant to the change type—deep security checks for auth code, lighter touch for internal tooling-only scripts.

## Summary

Review to **improve the change and the system**, classify feedback by **severity**, separate **must-fix** from **optional**, **manage time** by risk, and avoid **anti-patterns** that erode trust or signal. Clear, kind, specific comments scale better than volume or authority alone.
