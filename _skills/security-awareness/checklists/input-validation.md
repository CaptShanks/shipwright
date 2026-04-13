# Input Validation Checklist

## Strategy: Reject by Default

Accept only what you explicitly expect. Everything else is rejected.

## Validation Steps

For every input field or parameter, verify:

- [ ] **Type** -- is the value the expected type (string, int, bool, etc.)?
- [ ] **Length** -- does the string/array fit within min and max bounds?
- [ ] **Range** -- for numeric values, is it within the expected range?
- [ ] **Format** -- does it match the expected pattern (email, UUID, date)?
- [ ] **Character set** -- are only allowed characters present?
- [ ] **Encoding** -- is the value properly decoded before validation (no double-encoding)?

## Whitelist Over Blacklist

Prefer allowlists ("only these characters are permitted") over denylists ("reject these characters"). Denylists are incomplete by definition -- attackers will find characters or patterns you didn't think to block.

## Context-Specific Encoding

After validation, encode output based on where it's used:
- **HTML context** -- HTML-entity encode (`<` becomes `&lt;`)
- **SQL context** -- use parameterized queries, never string interpolation
- **Shell context** -- use library functions to escape arguments, never concatenate
- **URL context** -- percent-encode special characters
- **JSON context** -- use the language's JSON serializer, never manual string building

## Boundaries to Validate At

- API endpoint handlers (first line of defense)
- Deserialization points (JSON/YAML/protobuf parsing)
- File read operations (validate file path, size, and content type)
- Database query inputs (even internal service-to-service)
- Configuration loading (environment variables, config files)
