# Version Bump Decision Checklist

## Principle: Impact on Consumers Determines the Bump, Not Effort or Line Count

A one-line rename of a public function is MAJOR. A 5,000-line internal refactor is PATCH. Classify by what changes for the consumer, not by how hard you worked.

## Decision Flowchart

Work through these questions in order. Stop at the first "yes":

### Step 1: Is It MAJOR?

- [ ] Does any public contract change in a way that could break existing consumers?
- [ ] Was a public function, type, endpoint, flag, or config key **removed or renamed**?
- [ ] Was a function signature changed (added required param, changed return type)?
- [ ] Was a required field added to an API request body?
- [ ] Was a field type changed in a response, config, or schema?
- [ ] Was a default value changed for a flag, config key, or function parameter?
- [ ] Was input validation tightened (previously accepted input now rejected)?
- [ ] Was observable behavior changed in a way consumers may depend on?
- [ ] Was the minimum language, runtime, or platform version raised?
- [ ] Was a previously deprecated item actually removed?

**If any box is checked → MAJOR.**

### Step 2: Is It MINOR?

- [ ] Was a new public function, type, endpoint, flag, or config key **added**?
- [ ] Was new capability made available to consumers without changing existing behavior?
- [ ] Was a feature deprecated (but not yet removed)?
- [ ] Was a new optional field added to an API response or event payload?
- [ ] Was a new optional config key added with a backward-compatible default?
- [ ] Was input validation loosened in a way that does not change output semantics?

**If any box is checked and nothing from Step 1 applies → MINOR.**

### Step 3: Is It PATCH?

- [ ] Was a bug fixed without altering the public contract?
- [ ] Was a security vulnerability patched?
- [ ] Was a performance improvement made with no behavioral change?
- [ ] Were tests, documentation, or CI configuration updated?
- [ ] Was internal code refactored without changing public interfaces?
- [ ] Were dependencies updated with no impact on the public surface?

**If only these apply → PATCH.**

## Edge Case Decisions

### Bug fixes that change behavior

- [ ] Does the fix align with **documented** behavior? → PATCH is defensible
- [ ] Was the broken behavior **undocumented** but widely depended on? → Consider MAJOR
- [ ] Is the behavioral change prominent in the changelog with migration guidance? → Required regardless of bump level

### Dependency bumps

- [ ] Did a transitive dependency have a MAJOR version change?
- [ ] Does the transitive MAJOR change leak through your public surface? → Your release is MAJOR
- [ ] Is the transitive change fully encapsulated? → Match your own change level (MINOR or PATCH)

### Performance changes

- [ ] Is the change **faster** with no behavioral difference? → PATCH
- [ ] Does the change **significantly increase** resource usage (memory, connections, file handles)? → Consider MINOR or MAJOR depending on deployment impact
- [ ] Does the change affect SLA-relevant latency? → Document prominently, consider MINOR

### The 0.x.y exception

- [ ] Is the version below 1.0.0? → No stability promise; state your convention and follow it
- [ ] Common convention: 0.MINOR for potentially breaking changes, 0.PATCH for non-breaking

## Multiple Changes in One Release

When a release contains multiple changes:

- [ ] List all changes with their individual bump classification
- [ ] The **highest-impact change determines the release bump** (one MAJOR among fifty PATCHes = MAJOR release)
- [ ] Group changes in the changelog by classification (Breaking, Added, Fixed)

```markdown
## v3.0.0

### Breaking
- Removed `parseV1()` function (use `parse()` instead)
- Changed `--output` flag default from `text` to `json`

### Added
- New `validate()` function for input checking

### Fixed
- Fixed off-by-one error in pagination
```

## Pre-Release Versions

- [ ] Use pre-release suffixes for unstable versions: `1.0.0-alpha.1`, `2.0.0-rc.1`
- [ ] Pre-release versions sort before the release: `1.0.0-alpha.1 < 1.0.0`
- [ ] Pre-release versions carry lower stability expectations — breaking changes are expected
- [ ] Increment the pre-release identifier, not the version, for changes within a pre-release cycle

## Changelog Requirements

Every version bump should have a changelog entry that answers:

- [ ] **What** changed (factual description)
- [ ] **Why** this bump level (MAJOR/MINOR/PATCH rationale)
- [ ] **Migration path** (for MAJOR: what must consumers change?)
- [ ] **Deprecation notice** (for MINOR: what will be removed in a future MAJOR?)

## Final Validation

Before tagging the release:

- [ ] Re-read the diff between this version and the last release — does the bump level match what you see?
- [ ] Have contract tests or consumer-driven tests passed against the new version?
- [ ] Does the changelog accurately describe every consumer-visible change?
- [ ] For MAJOR bumps: is there a migration guide or upgrade path documented?
- [ ] For MINOR bumps: are deprecated items marked with a target removal version?

## Anti-Patterns

- **Bumping by effort** — "This was a huge refactor, so MAJOR." If the public surface is unchanged, it is PATCH.
- **Treating MINOR as 'small breaking change'** — There is no such thing. Any breakage is MAJOR. You do not know all your consumers.
- **Skipping MAJOR because it is scary** — Sneaking breaking changes into MINOR is worse than a MAJOR bump. MAJOR exists to say "read the changelog before upgrading."
- **Conflating code size with version significance** — Lines changed is noise. A one-character change to an auth check can be the most impactful change in the project's history.
- **Not stating the rationale** — The changelog should explain *why* this is MAJOR/MINOR/PATCH, not just state the number. Consumers use this to assess upgrade risk.
