# Authentication & Authorization — Code Review Checklist

Separate **authentication (authn)** — *who is this actor?* — from **authorization (authz)** — *what may this actor do on this resource?* Bugs cluster where one is present without the other, or where checks live only at the edge.

---

## Part A — Authentication (identity establishment)

### Session-based web auth

| Check | What to look for | Risk notes |
|-------|------------------|------------|
| **Session fixation** | Session ID issued pre-login and **not rotated** after successful authentication | Attacker plants ID; victim elevates it. Often **Medium–High**. |
| **Cookie flags** | Missing `HttpOnly`, `Secure`, sensible `SameSite` | XSS or network theft. **High** if XSS exists. |
| **Cookie scope** | Overbroad `Domain`, path `/` on shared hosts | Session leakage across apps. **Medium–High**. |
| **Session storage** | Server-side session with opaque ID vs self-contained client token | Prefer server-side invalidation for high-security apps. |
| **Logout** | Session destroyed server-side, not only cookie cleared | **Medium** if “logout” illusory. |

### Token-based auth (JWT, opaque tokens, API keys)

| Check | What to look for | Risk notes |
|-------|------------------|------------|
| **Algorithm confusion** | JWT `alg: none` accepted, or HMAC secret reused as RSA “public key” confusion patterns | **Critical** class; verify library defaults. |
| **Weak / missing signature verification** | `verify=False`, skipping validation in “internal” routes | **Critical**. |
| **`exp`, `nbf`, `iat`** | Missing or not enforced; clock skew not handled | Token replay / indefinite validity. **High**. |
| **Audience / issuer** | No `aud` / `iss` check when multiple clients or issuers exist | Wrong-token acceptance. **High**. |
| **Key rotation** | Long-lived single JWK; no kid rollover story | Operational risk; **Medium–High** over time. |
| **Storage in browsers** | JWT in **localStorage/sessionStorage** | XSS → token theft. **High** with any XSS surface. |
| **Refresh tokens** | Long-lived, not rotated, not revocable | Account persistence for attacker. **High**. |

### Passwords and credentials

| Check | What to look for | Risk notes |
|-------|------------------|------------|
| **Hashing** | bcrypt/argon2/scrypt with **per-user salt** and reasonable work factor | MD5/SHA1/plain → **Critical** at rest breach. |
| **Password reset** | Predictable tokens, long-lived, not single-use, logged in URLs | **High** takeover. |
| **MFA** | Sensitive accounts without MFA option; SMS as only factor | Policy/regulatory; phishing risk. |
| **Default / shared credentials** | Seed users in migrations | **Critical** if reachable. |

### Multi-step and “remember me” flows

- **Step-up auth**: Re-auth for sensitive actions (email change, 2FA disable, wire transfer).
- **“Remember this device”**: Cryptographically strong device binding; don’t store raw passwords in cookies.

### Enumeration and UX leaks

| Pattern | Why it matters | Typical severity |
|--------|----------------|------------------|
| Different messages for “unknown user” vs “bad password” | User enumeration | **Low–Medium** |
| Different response times on login | Timing side channel to guess users | **Low–Medium** |
| Registration reveals existing email | Enumeration | **Low–Medium** |

Severity rises if **password reset** or **MFA** flows can be targeted per user.

### Rate limiting and abuse

- **Login**, **OTP**, **reset**, **magic link**: throttle per IP *and* per identifier; lockout or backoff with care (avoid easy DoS).
- **CAPTCHA** only as part of a layered strategy — not a substitute for rate limits.

---

## Part B — Authorization (permission on resources)

### Core mistake: endpoint-only checks

```text
[Router] if (isLoggedIn) → [Service] // no check → DB access
```

Authorization must be enforced where **the resource is resolved** (service/domain layer), not only on the controller.

### IDOR (Insecure Direct Object Reference)

| Pattern | Example | Fix direction |
|--------|---------|----------------|
| **Sequential IDs** | `/api/invoice/1024` | Server checks `invoice.tenant_id == ctx.tenant`. |
| **UUID as secrecy** | Assuming UUID is unguessable | Still enforce ownership; IDs leak in logs/referrers. |
| **Mass assignment** | Client sends `role: "admin"` | DTO allowlists; policy layer. |

### Horizontal vs vertical escalation

- **Horizontal**: Same role, different user’s data (user A reads user B’s order).
- **Vertical**: Lower role gains admin capability (`role` param, hidden field, alternate API version).

### Common code smells

| Smell | What to look for |
|-------|------------------|
| **Trusting client role flags** | JWT claim `admin: true` without server-side mapping to user record |
| **Authz based on URL obscurity** | `/internal/v2` “secret” paths |
| **GraphQL** | Resolver returns object without field/object checks; **batch** queries amplify IDOR |
| **File URLs** | Signed URL without binding to user/session or too-long TTL |
| **“Admin impersonation”** | `X-User-Id` header honored without strong audit + role check | **Critical** if misused |

### Policy models (sanity check)

- Prefer **explicit policies**: RBAC + resource attributes (ReBAC) as needed.
- Centralize **“can(actor, action, resource)”**; avoid scattered `if (role == ...)` copies.

---

## Part C — Token handling (APIs, microservices, SPAs)

### Bearer tokens

- **Transport**: HTTPS only; reject mixed content; HSTS at edge.
- **Replay**: Short-lived access tokens; rotation for refresh; **jti** revocation if required.
- **Audience binding**: Token minted for API A should not work on API B without intent.

### API keys

- Keys are **long-lived secrets** — scope narrowly (read-only where possible), rotate, monitor usage.
- Never log full keys; log **prefix** only.

### Service-to-service

- **mTLS** or **signed service JWTs** with audience; avoid “internal network = trusted.”
- Metadata services (cloud): ensure apps don’t forward attacker-controlled URLs (SSRF + token theft).

---

## Part D — Session vs CSRF (web forms)

- **Stateful session cookies**: use **CSRF tokens** or **SameSite** cookies appropriately; verify for **state-changing** requests.
- **Double-submit cookie** pattern: understand limitations (subdomain attacks, XSS).

---

## Severity mapping (CVSS-aligned, qualitative)

| Scenario | Typical band |
|----------|----------------|
| Unauthenticated access to privileged API affecting many users | **Critical** |
| Authn bypass (accept bad JWT, missing verification) | **Critical** |
| IDOR on sensitive records (health, financial, credentials) | **High–Critical** |
| Missing authz on state-changing resource (non-public data) | **High** |
| Session fixation / weak cookie flags with XSS present | **High** |
| User enumeration, timing leaks | **Low–Medium** |

Always **confirm with concrete code paths** and **deployment exposure** (public internet vs VPN).

---

## Reviewer closing questions

1. Where is **identity** established, and can it be **spoofed** or **replayed**?
2. Where is the **resource** loaded, and is there an **ownership or policy** check?
3. Are **all** entry points covered (REST, GraphQL, gRPC, workers, admin CLI, webhooks)?
4. Is **logout / revocation** real for this token model?
5. What do **logs** contain (tokens, session IDs, PII)?
