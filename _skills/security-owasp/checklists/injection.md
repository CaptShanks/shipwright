# Injection — Code Review Checklist

Injection occurs when **untrusted input becomes part of a command or query** interpreted by another engine. The reviewer’s job is to find **sources** (user, file, queue, webhook, SSRF-fetched content), **concatenation or unsafe APIs**, and **sinks** (database, OS, LDAP, XML, template, etc.).

---

## General signals (any language)

- String **concatenation**, **template literals**, or **format strings** that include variables in:
  - SQL / CQL / SPARQL / ORM raw fragments
  - Shell commands
  - LDAP filters / DN assembly
  - XPath / XQuery
  - OS file paths (especially with `..` segments)
  - HTML / SVG / Markdown render pipelines
  - Email headers (CRLF injection)
- **Dynamic evaluation**: `eval`, `exec`, `Function()`, `setTimeout(string)`, reflection-based invocation from user input.
- **Second-order injection**: Stored values later used in a query or command without re-parameterization.
- **ORM “escape hatch”**: `.raw()`, `.unsafe()`, `$queryRaw`, `literal()`, `text()` wrapping user fragments.

For each finding, note: **who can control the input**, **authentication/authorization**, **reachability**, and **data sensitivity** — these drive CVSS.

---

## SQL injection

### High-risk patterns

| Pattern | Examples | Notes |
|--------|-----------|--------|
| **Concatenated SQL** | `"SELECT * FROM u WHERE id=" + id` | Classic. Often **Critical** if DB holds sensitive data. |
| **Interpolated SQL** | `` f"SELECT ... WHERE email='{email}'" `` | Same class as concatenation. |
| **String-built dynamic identifiers** | `"ORDER BY " + column` | Metadata injection; whitelist columns, never quote-escape names. |
| **Raw fragments in ORMs** | `WHERE ${sqlFragment}`, Knex `.whereRaw(req.query)` | ORM does not save you. |
| **Like-clause injection** | `LIKE '%" + user + "%'` | `%` and `_` wildcards; escape or parameterize pattern. |
| **Bulk raw SQL in migrations/tools** | Admin scripts reading CSV into SQL strings | Often forgotten in reviews. |

### Safer patterns (verify they are used correctly)

- **Prepared statements / bind parameters** end-to-end (no string pasting into the query text).
- **Query builders** only if **every** dynamic part is a **bound value** or **allowlisted** identifier.
- **ORM** with typed APIs; watch for `literal()` / raw SQL.

### Tests to mentally run

- Single quote `'` and double `--` in strings.
- Stacked queries `; DROP` (if driver allows).
- Unicode homoglyphs and encoding tricks if DB collation is loose.

---

## Command injection (OS)

### High-risk patterns

| Pattern | Examples | Notes |
|--------|-----------|--------|
| **Shell invocation with string** | `subprocess.run(cmd, shell=True)` | **Critical** when `cmd` includes user input. |
| **Java `Runtime.getRuntime().exec(String)`** | Built string with spaces — often broken into wrong tokens; still dangerous with crafted input. | Prefer `ProcessBuilder` with argument list. |
| **Node** | `child_process.exec`, `execSync` with concatenation | RCE vector. |
| **PHP** | `shell_exec`, `system`, backticks | Same. |
| **Go** | `exec.Command("sh", "-c", user)` | Sh -c is a shell; avoid. |

### Safer patterns

- **Argv array** APIs with **no shell**: `subprocess.run(["bin", arg], shell=False)`.
- **Fixed executable** path; user input only as **discrete arguments**, validated (type, charset, length).
- **Avoid** `sh -c` unless absolutely required and input is strictly allowlisted.

### Wrinkles

- **Environment variable injection** if user controls env (`ENV` child poisoning).
- **Indirect command injection** via filenames passed to tools (`;`, `|`, `` ` `` in names).

---

## Path / directory traversal

### High-risk patterns

| Pattern | Examples | Notes |
|--------|-----------|--------|
| **User file name joined to base** | `open(base + userFile)` | `../` escapes. |
| **URL path → filesystem** | Static file servers mapping URL to disk | Classic LFI/RFI. |
| **Zip slip** | Extracting archives without normalizing paths | Overwrite arbitrary files. |
| **Include from user path** | `include $_GET['page'] . '.php'` | Remote/local file include. |

### Safer patterns

- Resolve with **`realpath` / `GetFullPath`** and verify **prefix** under base directory.
- **Allowlist** filenames; reject `..`, absolute paths, null bytes (legacy), odd encodings.
- Use **content-addressed** storage IDs instead of raw names from users.

---

## LDAP injection

### High-risk patterns

| Pattern | Examples | Notes |
|--------|-----------|--------|
| **Filter concatenation** | `"(uid=" + uid + ")"` | Metacharacters `* ( ) \ NUL` alter filter logic. |
| **DN assembly** | `cn=" + name + ",ou=users,..."` | Can break out of DN context. |

### Safer patterns

- **Parameterized LDAP filters** or framework APIs that escape per RFC 4515.
- **Validate** input charset and length; **allowlist** usernames if possible.
- **Least-privilege bind** for app (not directory admin).

---

## Template injection (server-side)

### High-risk patterns

| Pattern | Examples | Notes |
|--------|-----------|--------|
| **User content as template** | Jinja2 `Template(user)`, Flask `render_template_string` | Often **Critical** (RCE in many engines). |
| **Expression languages in frameworks** | OGNL, SpEL in user-controlled views | RCE history. |
| **“Format” with evaluation** | Some logging or message formatters with `${}` execution | Log4j-class issues; treat as injection. |

### Safer patterns

- **Fixed templates** with **data-only** context objects.
- **Sandbox** only if proven; prefer **no user-controlled template source**.
- For **client-side** templates, separate from server SSTI; still check for **XSS** (different category, similar taint).

---

## NoSQL injection

### High-risk patterns

| Pattern | Examples | Notes |
|--------|-----------|--------|
| **Mongo `$where` / JS** | String includes user input | Can execute JS server-side. |
| **Operator injection** | Parsing JSON into query document without schema | `$gt`, `$ne` in manipulated objects. |
| **String-built query JSON** | Concat into Mongo query | Same as SQL mentally. |

### Safer patterns

- **Typed builders**, **explicit schemas**, **prohibit** raw document fragments from clients.
- For HTTP APIs: map DTO fields to **fixed** query shapes.

---

## XML / XPath / XQuery injection

### High-risk patterns

| Pattern | Examples | Notes |
|--------|-----------|--------|
| **XPath string concat** | `"//user[name='" + name + "']"` | Quote breaking; node theft. |
| **XXE** | `DocumentBuilder` / `libxml` with external entities enabled | File read, SSRF. |

### Safer patterns

- **Parameterized XPath** (if available) or strict **allowlist** on values.
- **Disable** external entities and DTDs unless required; use **safe parsers**.

---

## CRLF / header injection

### High-risk patterns

- User input written to **HTTP headers**, **SMTP headers**, or **log lines** without stripping `\r\n`.
- **Open redirect** + header injection chains.

### Safer patterns

- Reject or strip CR/LF; use **framework APIs** for redirects with **allowlisted** targets.

---

## Quick triage questions

1. Is the sink an **interpreter** (SQL, shell, LDAP, template, XPath)?
2. Is every dynamic part a **bound value** or **allowlisted literal**?
3. Is there a **second-order** path (DB → later query)?
4. Does the **ORM/raw split** hide a string fragment built from user input?
5. What is the **worst plausible primitive** (read row, write row, RCE, file read)?

Document answers when reporting severity; align with **CVSS** (especially **privileges required** and **confidentiality/integrity** impact).
