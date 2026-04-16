# Breaking Change Detection Checklist

## Principle: A Breaking Change Is Anything That Makes a Working Consumer Fail

The test is not "did the signature change?" — it is "would a consumer who upgrades without reading the changelog have a bad time?" Check every contract surface, not just the obvious ones.

## 1. API Contracts (REST, gRPC, GraphQL)

- [ ] Were any endpoints, fields, or methods **removed or renamed**?
- [ ] Were any **response field types changed** (string → int, nullable → required, array → object)?
- [ ] Were any **required fields added** to request bodies?
- [ ] Were **error codes or HTTP status codes changed** for existing scenarios?
- [ ] Was **pagination behavior altered** (page size defaults, sort order, cursor format)?
- [ ] Was **input validation tightened** (previously accepted values now rejected)?
- [ ] Were **rate limits changed** in a way that affects documented contracts?
- [ ] Were **authentication or authorization requirements changed** for existing endpoints?

```
# Compare OpenAPI specs for breaking changes
openapi-diff old-spec.yaml new-spec.yaml

# For gRPC, check proto compatibility
buf breaking --against '.git#branch=main'
```

## 2. CLI Contracts

- [ ] Were any **flags or subcommands removed or renamed**?
- [ ] Were any **flag default values changed**?
- [ ] Was **exit code meaning changed** (scripts and CI pipelines match on exit codes)?
- [ ] Was **stdout/stderr output format changed** (consumers may parse output)?
- [ ] Were **environment variable names changed** that the CLI reads?
- [ ] Was the **order of positional arguments changed**?
- [ ] Were **interactive prompts added** that break non-interactive/scripted usage?

```bash
# Before: --output flag exists
mytool export --output json

# After: renamed to --format (BREAKING — scripts using --output will fail)
mytool export --format json
```

## 3. Configuration Contracts

- [ ] Were any **config keys removed or renamed**?
- [ ] Were any **default values changed** for existing keys?
- [ ] Were any **required config keys added** without backward-compatible defaults?
- [ ] Were **environment variable names changed**?
- [ ] Was the **config file format changed** (YAML → TOML, flat → nested)?
- [ ] Were **validation rules tightened** on existing config values?

## 4. Library / SDK Contracts

- [ ] Were any **public functions, types, or constants removed or renamed**?
- [ ] Were any **function signatures changed** (added required params, changed return types)?
- [ ] Was the **observable behavior of a public function changed** (even without signature change)?
- [ ] Was the **minimum language or runtime version raised**?
- [ ] Were **transitive dependencies removed** that consumers may rely on?
- [ ] Were **exported interfaces changed** (added methods that implementors must satisfy)?

```go
// Before: 2 return values
func Parse(input string) (Result, error)

// After: 3 return values (BREAKING — callers expect 2)
func Parse(input string) (Result, []Warning, error)
```

## 5. Data and Schema Contracts

- [ ] Were any **columns or fields removed** from persisted schemas?
- [ ] Were any **column types or constraints changed** (VARCHAR → INT, nullable → NOT NULL)?
- [ ] Were **message queue topics or event type names changed**?
- [ ] Was the **shape of event payloads changed** (field removal, type change, restructuring)?
- [ ] Was the **serialization format changed** (JSON → protobuf, XML → JSON)?
- [ ] Were **database indexes removed** that consumers depend on for query performance?

## 6. Behavioral Changes (No Signature Change)

These are the hardest to catch — the type system won't flag them:

- [ ] Was **sort order changed** for returned collections?
- [ ] Was **rounding or precision changed** for numeric calculations?
- [ ] Were **default timeout or retry values changed**?
- [ ] Was **error message text changed** that consumers may match on (fragile, but real)?
- [ ] Were **side effects changed** (a function that used to log now doesn't, or vice versa)?
- [ ] Was **concurrency behavior changed** (thread-safe → not thread-safe, or ordering guarantees removed)?

## 7. Deployment and Operational Contracts

- [ ] Were **new system dependencies added** (requires Redis where before it didn't)?
- [ ] Were **resource requirements significantly increased** (memory, CPU, file handles, connections)?
- [ ] Were **health check endpoints changed** in behavior or response format?
- [ ] Were **metric names or label schemas changed** (breaks dashboards and alerts)?
- [ ] Were **log format or log level semantics changed**?

## 8. Cross-Cutting Checks

- [ ] Run **diff of all public API surfaces** (exported types, functions, constants) between versions
- [ ] Run **contract tests** or **consumer-driven contract tests** against the new version
- [ ] Check **dependency update impact**: did a transitive dependency's MAJOR bump leak through?
- [ ] Review **changelog entries** — does every breaking item have a migration path documented?
- [ ] Search for `TODO`, `FIXME`, `BREAKING` comments added during development

## Anti-Patterns

- **Checking only type signatures** — Behavioral changes with identical signatures are still breaking. A function returning results in a different order is breaking if consumers depend on ordering.
- **Assuming internal consumers don't count** — If another team's service calls your API, they are a consumer. Internal does not mean unbreakable.
- **Relying on "nobody uses that"** — You do not know all your consumers. If it was public, assume someone depends on it.
- **Conflating deprecation with removal** — Marking something deprecated is MINOR. Removing the deprecated thing is MAJOR. These are separate version bumps.
- **Ignoring configuration contracts** — A renamed environment variable is a breaking change even though no source code changed from the consumer's perspective.
