---
name: semantic-versioning
description: >-
  Breaking change detection, semver decision rules, pre-release semantics, and
  version bump classification across contract types. Use when generating
  changelogs, cutting releases, classifying changes, or deciding version bumps.
license: MIT
metadata:
  author: CaptShanks
  version: "1.0.0"
---

# Semantic Versioning

This skill encodes the decision framework for classifying changes and determining version bumps. Semver is a **communication protocol** between producers and consumers: the version number is a promise about compatibility. A wrong version is a broken promise—and broken promises cause production incidents, delayed upgrades, and eroded trust.

## The Contract

Given a version `MAJOR.MINOR.PATCH`:

- **MAJOR** — you broke a public contract. Consumers must review and may need to change their code, config, or workflows to upgrade.
- **MINOR** — you added capability. Consumers can upgrade without changes, but new features are available.
- **PATCH** — you fixed a defect. Consumers should upgrade; nothing will change except the bug is gone.

This is not a "how big is the change" scale. A one-line diff that renames a public function is MAJOR. A 5,000-line diff that adds an entire subsystem behind a new flag is MINOR. **Impact on consumers determines the bump, not effort or line count.**

## What Constitutes a Breaking Change

A breaking change is anything that could cause a consumer's existing, working integration to fail, produce different results, or require modification after upgrading. The bar is: **would a consumer who reads the changelog and upgrades without changing their code have a bad time?**

### API contracts (REST, gRPC, GraphQL)

| Change | Breaking? | Notes |
|--------|-----------|-------|
| Remove an endpoint or field | **Yes** | Always MAJOR |
| Rename an endpoint or field | **Yes** | Removal + addition is still breaking |
| Change a field's type | **Yes** | `string` → `int`, nullable → required |
| Add a required field to a request body | **Yes** | Existing callers don't send it |
| Add an optional field to a response | No | Consumers should tolerate unknown fields |
| Change error codes or status codes | **Yes** | Consumers may match on specific codes |
| Change pagination behavior | **Yes** | Different result ordering, page sizes |
| Tighten input validation | **Yes** | Previously accepted input now rejected |
| Loosen input validation | No | Unless it changes output semantics |
| Change rate limits | **Maybe** | If documented as part of the contract |

### CLI contracts

| Change | Breaking? | Notes |
|--------|-----------|-------|
| Remove a flag or subcommand | **Yes** | Scripts will fail |
| Rename a flag | **Yes** | `--output` → `--out` breaks scripts |
| Change a flag's default value | **Yes** | Silent behavior change |
| Change exit code semantics | **Yes** | CI pipelines match on exit codes |
| Change stdout/stderr format | **Yes** | If consumers parse output (common in CLI tools) |
| Add a new flag | No | Existing invocations unchanged |
| Add a new subcommand | No | Unless it changes `help` behavior scripts rely on |

### Configuration contracts

| Change | Breaking? | Notes |
|--------|-----------|-------|
| Remove a config key | **Yes** | System may fail to start or silently misconfig |
| Rename a config key | **Yes** | Removal + addition |
| Change a default value | **Yes** | Existing deployments get different behavior |
| Add a required config key | **Yes** | Existing config files are now incomplete |
| Add an optional key with a backward-compatible default | No | System behaves identically without it |
| Change environment variable names | **Yes** | Deployment scripts and containers will break |

### Library/SDK contracts

| Change | Breaking? | Notes |
|--------|-----------|-------|
| Remove or rename a public function/type/constant | **Yes** | Compile failure for consumers |
| Change a function signature (add required param, change return type) | **Yes** | Compile failure |
| Change observable behavior of a public function | **Yes** | Even without signature change |
| Raise minimum language/runtime version | **Yes** | Consumers on older versions cannot compile |
| Remove a transitive dependency consumers relied on | **Yes** | Even if undocumented |
| Add a new exported function | No | MINOR |
| Add an optional parameter with default (if language supports it) | No | Existing call sites unchanged |

### Data and schema contracts

| Change | Breaking? | Notes |
|--------|-----------|-------|
| Remove a column/field from a persisted schema | **Yes** | Existing queries and code reference it |
| Change a column type or constraints | **Yes** | Data may not migrate cleanly |
| Rename a message queue topic or event type | **Yes** | Consumers are subscribed to the old name |
| Change the shape of an event payload | **Yes** | Consumers deserialize the old shape |
| Add an optional field to an event | No | Consumers should tolerate unknown fields |
| Change serialization format (JSON → protobuf) | **Yes** | Even if logically equivalent |

## Edge Cases and Tricky Decisions

### The 0.x.y exception

Versions before `1.0.0` carry no stability promise. Any 0.x.y bump can contain breaking changes. However, **convention** (especially in Go modules) treats `0.MINOR` bumps as potentially breaking and `0.PATCH` as non-breaking. State your convention and follow it consistently.

### Behavioral changes without signature changes

A function that used to return results sorted ascending now returns them unsorted. The type signature is identical. **This is a breaking change** if consumers relied on the ordering, even if it was never documented. When in doubt, treat observable behavior changes as MAJOR unless you can prove no consumer depends on the behavior.

### Bug fixes that change behavior

A function that was "broken" may have consumers who depend on the broken behavior. If the fix changes output for valid inputs, it is technically a breaking change. Apply judgment: if the fix aligns with documented intent and the broken behavior was clearly unintended, a PATCH is defensible—but call it out prominently in the changelog.

### Deprecation is not removal

Marking something as deprecated is a MINOR change (it adds a signal). Actually removing the deprecated item is a MAJOR change. The deprecation period between them is the consumer's migration window.

### Transitive dependency bumps

If you bump a dependency and that dependency had a MAJOR version change, your consumers may be affected even if your public API did not change. Assess whether the transitive change leaks through your public surface. If it does, it is MAJOR.

### Performance changes

Faster is generally not breaking. Significantly slower might be—if consumers have SLAs that depend on your performance. Increased memory usage or new resource requirements (more file handles, more connections) can be breaking for deployment.

### Pre-release versions

Pre-release versions (`1.0.0-alpha.1`, `2.0.0-rc.1`) have lower stability expectations. They sort before the release version (`1.0.0-alpha.1 < 1.0.0`). Pre-release identifiers are compared left-to-right: numeric identifiers by value, alphanumeric by ASCII sort. `1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-beta < 1.0.0-rc.1 < 1.0.0`.

### Build metadata

Build metadata (`1.0.0+build.123`) is ignored for version precedence. Two versions that differ only in build metadata are considered equal for ordering. Use build metadata for traceability (commit hash, build number), not for versioning.

## Decision Framework

When classifying a change for version bump, ask in order:

1. **Does any public contract change in a way that could break existing consumers?** → MAJOR
2. **Does the change add new capability accessible to consumers?** → MINOR
3. **Does the change fix a defect without altering contracts?** → PATCH
4. **Does the change affect only internals (tests, CI, docs, refactors)?** → PATCH (or no bump if the project only bumps for user-visible changes)

When multiple changes are in a release, **the highest-impact change determines the bump**. One MAJOR change among fifty PATCH changes makes the entire release MAJOR.

### Ambiguity resolution

- When a behavioral change is subtle and you cannot determine if consumers rely on it, **default to MAJOR** and explain your reasoning. False alarm on MAJOR is cheap; false confidence on MINOR causes incidents.
- When a bug fix changes output, check whether the fix aligns with **documented behavior**. If it does, PATCH is defensible. If documentation was ambiguous, call it out.
- When a dependency bump is MAJOR but your public surface is unchanged, test whether the transitive change leaks. If unsure, bump MINOR and document the dependency change prominently.

## Common Mistakes

- **Bumping by effort, not impact.** "This was a huge refactor, so it must be MAJOR." No—if the public surface is unchanged, it is PATCH.
- **Treating MINOR as "small breaking change."** There is no such thing. Any breakage is MAJOR, regardless of how minor you think the impact is. You do not know all your consumers.
- **Skipping MAJOR because it is scary.** The purpose of MAJOR is to signal "read the changelog before upgrading." Avoiding it by sneaking breaking changes into MINOR is worse than a MAJOR bump.
- **Conflating code change size with version significance.** One character change (`=` → `!=` in an auth check) can be the most impactful change in the project's history. Lines changed is noise.
- **Ignoring configuration and environment contracts.** A renamed environment variable is a breaking change even though no code file changed from the consumer's perspective.
- **Not specifying the version bump rationale.** The changelog should explain *why* this is a MAJOR/MINOR/PATCH, not just state the number. Consumers use this to assess upgrade risk.
