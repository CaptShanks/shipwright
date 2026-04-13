---
name: security-owasp
description: >-
  Deep OWASP Top 10 (2021) review guidance: map each category to concrete code
  patterns, static-review heuristics, and CVSS-aligned severity bands. Use when
  performing security reviews, threat modeling, or triaging vulnerabilities
  against OWASP categories.
license: MIT
metadata:
  author: CaptShanks
  version: "1.0.0"
---

# Security OWASP (Top 10 Specialist)

Use this skill when you need **systematic, code-grounded** analysis aligned with **OWASP Top 10:2021**, not generic security platitudes. Treat every finding as a hypothesis: confirm exploitability, scope, and data sensitivity before assigning severity.

## How to use this skill

1. **Classify the issue** against the OWASP category below (primary; secondary if needed).
2. **Match code patterns** from the tables to the implementation under review.
3. **Estimate CVSS v3.1** (Base) using attack vector, privileges required, user interaction, scope, and impact on confidentiality / integrity / availability. Map the numeric score to **Critical / High / Medium / Low** using standard bands:
   - **Critical**: 9.0–10.0  
   - **High**: 7.0–8.9  
   - **Medium**: 4.0–6.9  
   - **Low**: 0.1–3.9  
4. **Cross-check** focused checklists: `checklists/injection.md`, `checklists/auth.md`, `checklists/secrets.md`.

OWASP categories are **risk themes**. A single bug may span multiple themes (e.g., SSRF plus broken access control). Pick the **root cause** category for the primary label.

---

## A01:2021 — Broken Access Control

**Theme:** Users can act on objects or actions they should not (horizontal/vertical IDOR, missing function-level checks, forced browsing, metadata tampering).

### Code-level detection patterns

| Pattern | What to look for | Typical CVSS band |
|--------|-------------------|-------------------|
| **Object references from the client** | IDs in path/query/body with no server-side ownership check | Often **High–Critical** if sensitive data |
| **Role checks only at UI/route** | `@RolesAllowed` / front-end gating without service-layer enforcement | **High** when bypass trivial |
| **Copy-paste authorization** | Same `if (isAdmin)` block; no per-resource policy | **Medium–High** |
| **Mass assignment** | `update(req.body)` / ORM `.assign(dto)` into privileged fields | **High** |
| **Feature flags as security** | “Hidden” admin endpoints enabled by config clients control | **Medium–High** |
| **JWT claims trusted blindly** | `role` or `tenant_id` from token without server-side binding to session | **High–Critical** |
| **GraphQL / batch APIs** | Resolvers that omit field- or object-level authz | **High** |

### Severity heuristics (CVSS-aligned)

- **Critical**: Unauthenticated or low-privilege access to bulk PII, credentials, or destructive admin actions across tenants.
- **High**: Authenticated IDOR on sensitive records, or missing authz on state-changing APIs.
- **Medium**: Leakage of non-sensitive metadata, or issues requiring unusual knowledge/chaining.
- **Low**: Theoretical misconfiguration with no demonstrated impact on protected resources.

---

## A02:2021 — Cryptographic Failures

**Theme:** Sensitive data exposed or weakly protected in transit, at rest, or in application logic (not “crypto library bugs” only — often **misuse**).

### Code-level detection patterns

| Pattern | What to look for | Typical CVSS band |
|--------|-------------------|-------------------|
| **Hardcoded keys / IVs / salts** | Fixed `AES_KEY`, `iv = 0`, same salt per user | **High–Critical** |
| **ECB mode / “encrypt” without auth** | `AES/ECB`, raw AES without MAC/AEAD | **High** |
| **MD5/SHA1 for passwords** | `md5(password)`, unsalted SHA1 | **High** |
| **Custom “encoding” as secrecy** | Base64, XOR, “proprietary” obfuscation | **Medium–High** |
| **TLS off / wrong cert validation** | `verify=False`, `InsecureSkipVerify`, mixed content | **High–Critical** (context) |
| **Secrets in logs/errors** | Crypto operations logging keys, plaintext, or full payloads | **High** |
| **Predictable tokens** | Random from weak PRNG, time-based only, short numeric OTP | **Medium–High** |

### Severity heuristics

- **Critical**: Recoverable secrets at scale (live keys, database of password hashes with weak algorithm).
- **High**: Ciphertext or hashes trivially attackable, or TLS validation disabled in production paths.
- **Medium**: Legacy algorithm with limited exposure window or non-production code path.
- **Low**: Theoretical weakness with strong operational controls and no sensitive data.

---

## A03:2021 — Injection

**Theme:** Untrusted input interpreted as code or structure by an interpreter (SQL, OS, LDAP, XPath, template engine, etc.).

See **`checklists/injection.md`** for exhaustive pattern lists.

### Code-level detection patterns (summary)

| Pattern | What to look for | Typical CVSS band |
|--------|-------------------|-------------------|
| **String-built SQL** | `"SELECT ... WHERE id = " + id`, f-strings in queries | **High–Critical** |
| **Shell with user input** | `exec`, `subprocess(shell=True)`, `Runtime.getRuntime().exec` with concatenation | **Critical** (often RCE) |
| **Unsafe HTML APIs** | `dangerouslySetInnerHTML`, `.html()`, `v-html` with partial user data | **Medium–High** (XSS → session) |
| **Template injection** | `eval`, `render_template_string`, `${...}` in server templates with user input | **High–Critical** |
| **NoSQL injection** | `$where`, JSON operators built from strings | **High** |

### Severity heuristics

- **Critical**: Remote code execution, full database read/write, or OS command execution from unauthenticated or low-priv vectors.
- **High**: Authenticated SQLi with meaningful data impact; stored XSS in high-trust contexts.
- **Medium**: Reflected XSS with mitigations; blind injection with slow exfiltration.
- **Low**: Injection in dead code, test fixtures, or with strong parser separation (still fix).

---

## A04:2021 — Insecure Design

**Theme:** Missing or wrong **security controls by design** — not a single implementation bug but an architecture that cannot be safe (e.g., no rate limits on auth, “security by URL obscurity,” trust boundaries drawn incorrectly).

### Code-level / design detection patterns

| Pattern | What to look for | Typical CVSS band |
|--------|-------------------|-------------------|
| **No abuse controls** | Login, password reset, OTP, webhooks without throttling or lockout | **Medium–High** |
| **Global identifiers** | Sequential public IDs for all tenants without scoping story | **Medium–High** |
| **Trusting the network** | “Internal microservice; no auth between services” | **High–Critical** |
| **PII in URLs** | Tokens or emails in query strings logged everywhere | **Medium–High** |
| **Dangerous defaults** | Open CORS `*`, debug on, permissive CORS with credentials | **Medium–High** |

### Severity heuristics

- Tie severity to **maximum credible incident** if design is deployed as-is. Often **High** when compensating controls are absent; **Medium** when some limits exist.

---

## A05:2021 — Security Misconfiguration

**Theme:** Unsafe defaults, verbose errors, open permissions, exposed admin interfaces, directory listings, cloud bucket policies.

### Code-level detection patterns

| Pattern | What to look for | Typical CVSS band |
|--------|-------------------|-------------------|
| **Debug in prod** | `DEBUG=true`, stack traces to clients, Django `DEBUG` | **Medium–High** |
| **Default creds** | `admin/admin`, default DB passwords in compose | **Critical** |
| **Overbroad CORS** | `Access-Control-Allow-Origin: *` with `Allow-Credentials: true` | **High** |
| **Directory listing / backup files** | `.git`, `.env`, `*.bak` served | **High–Critical** |
| **Wildcard IAM / `**`** in policies** | S3 `Principal: "*"`, `Action: "*"` | **High–Critical** |

### Severity heuristics

- **Critical**: Public data store or admin with default/weak protection and sensitive data.
- **High**: Full config or secret material exposed via misconfiguration.
- **Medium**: Information disclosure aiding further attacks.
- **Low**: Hardening gaps with no direct data exposure.

---

## A06:2021 — Vulnerable and Outdated Components

**Theme:** Known CVEs in dependencies, unpinned versions, transitive risk.

### Code-level detection patterns

| Pattern | What to look for | Typical CVSS band |
|--------|-------------------|-------------------|
| **Unpinned deps** | `*`, `latest`, wide ranges in lockfile-absent ecosystems | Use **NVD CVSS** for the specific CVE |
| **Bundled/vendored libs** | Old OpenSSL, zlib, image parsers copied into repo | Match CVE |
| **Container base images** | `FROM` without digest, stale tags | Match CVE / config exposure |

### Severity heuristics

- Use the **published CVSS for the CVE** and adjust with **exploitability in your deployment** (exposed port, attack path, compensating controls). Do not invent a lower score without evidence.

---

## A07:2021 — Identification and Authentication Failures

**Theme:** Weak login flows, session fixation, credential stuffing enablers, bad MFA, weak recovery.

See **`checklists/auth.md`**.

### Code-level detection patterns

| Pattern | What to look for | Typical CVSS band |
|--------|-------------------|-------------------|
| **Weak password policy only** | No MFA for sensitive accounts | **Medium** (policy) |
| **Session not rotated on login** | Same session ID before/after auth | **Medium–High** |
| **JWT in localStorage** | XSS → token theft | **High** (with XSS presence) |
| **Long-lived refresh tokens** | No rotation/revocation | **Medium–High** |
| **User enumeration** | Different errors for “bad user” vs “bad password” | **Low–Medium** |

### Severity heuristics

- **Critical**: Complete account takeover at scale (e.g., missing auth on password change).
- **High**: Session hijack, broken MFA bypass, or credential stuffing without friction on high-value targets.
- **Medium**: Enumeration, partial weaknesses requiring chaining.
- **Low**: Minor UX leaks with no direct account compromise.

---

## A08:2021 — Software and Data Integrity Failures

**Theme:** Unsigned updates, unsafe deserialization, CI/CD trust, supply chain (malicious packages, typosquatting).

### Code-level detection patterns

| Pattern | What to look for | Typical CVSS band |
|--------|-------------------|-------------------|
| **Unsafe deserialization** | `pickle.loads`, `ObjectInputStream` on untrusted bytes, YAML `unsafe_load` | **Critical** |
| **Unsigned plugins** | Dynamic code load from URL without signature | **High–Critical** |
| **Dependency confusion** | Private package names without registry pinning | **High** |
| **CI secrets in forks** | PR workflows with write tokens | **Critical** |

### Severity heuristics

- **Critical**: RCE via deserialization or CI that can push to prod.
- **High**: Integrity bypass leading to code execution or trusted artifact tampering.
- **Medium**: Partial integrity issues without proven execution path.

---

## A09:2021 — Security Logging and Monitoring Failures

**Theme:** Insufficient detection/response — attackers operate undetected (also covers **logging sensitive data**, which overlaps A02/A03).

### Code-level detection patterns

| Pattern | What to look for | Typical CVSS band |
|--------|-------------------|-------------------|
| **Swallowed security exceptions** | Empty `catch` around auth/payment | **Medium** (detection gap) |
| **No audit trail** | Admin actions without actor, object, before/after | **Medium** |
| **PII in logs** | Passwords, tokens, health data in application logs | **High** (A02 overlap) |
| **Log injection** | Unsanitized `\n` in log fields forging entries | **Low–Medium** |

### Severity heuristics

- Often **Medium** as *direct* CVSS unless combined with active breach impact; elevate when regulations require auditability or when absence enables sustained compromise.

---

## A10:2021 — Server-Side Request Forgery (SSRF)

**Theme:** Server induced to request attacker-chosen URLs, hitting cloud metadata, internal services, or file URLs.

### Code-level detection patterns

| Pattern | What to look for | Typical CVSS band |
|--------|-------------------|-------------------|
| **User-controlled URL fetch** | `fetch(userUrl)`, `requests.get(url)`, `HttpClient` with tainted host | **High–Critical** |
| **URL parsers trusting user** | Bypass via `127.0.0.1`, `0x7f000001`, IPv6, redirects, DNS rebinding | **High–Critical** |
| **PDF/SSRF gadgets** | Image loaders, webhook validators, “preview” features | **High** |
| **Allowlists missing scheme/port** | `http://169.254.169.254` allowed by partial checks | **Critical** (cloud) |

### Severity heuristics

- **Critical**: Access to cloud metadata credentials or unrestricted internal network from production.
- **High**: Reach internal admin panels or read internal APIs.
- **Medium**: Blind SSRF or limited protocols.
- **Low**: Outbound blocked by egress firewall with evidence.

---

## Review workflow (recommended)

1. **Inventory trust boundaries** (browser, mobile, partner API, batch, admin).
2. **Trace data flow** from each untrusted input to sinks (DB, shell, HTTP client, template, file path).
3. **Verify authn** on every entry point; **authz** on every resource access.
4. **Score with CVSS**; document assumptions (network exposure, auth requirements, data classification).
5. **Propose fixes** at the right layer: parameterization, policy engine, framework defaults, platform controls (WAF, IAM, egress filtering).

---

## Relationship to `security-awareness`

The **`security-awareness`** skill is the **baseline** for all engineers. This skill is the **deep OWASP map** for reviewers: richer pattern vocabulary, explicit OWASP taxonomy, and CVSS-aligned triage. Use both together in large reviews.
