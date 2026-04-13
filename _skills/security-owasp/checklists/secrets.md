# Secrets in Code & Config — Review Checklist

Secrets in repositories, images, logs, or client bundles are a **leading source of incident**. This checklist covers **entropy-based heuristics**, **known formats**, **`.env` exposure**, and **common code patterns** — for manual review and for tuning automated scanners.

---

## Principles

1. **Assume leakage**: git history, forks, CI logs, crash reports, browser bundles, mobile apps, and screenshots all outlive “temporary” testing keys.
2. **High entropy ≠ proof of secrecy**: random IDs, hashes, and compressed data can false-positive; **correlate with context** (variable names, neighboring strings, file path).
3. **Context is decisive**: `AWS_SECRET_ACCESS_KEY=` adjacent to 40 hex chars is stronger evidence than bare hex in a test vector.

---

## Entropy-based detection (conceptual)

High-signal secret material often:

- Uses **base64**, **base32**, **hex**, or **ASCII armoring** (PEM blocks).
- Has **length above typical IDs** (e.g., 32+ random bytes when encoded).
- Appears **without** dictionary structure (unlike prose or UUIDs with fixed hyphen pattern — though **UUIDs can be capability tokens** in some designs).

**Practical reviewer approach:**

- Grep for **assignment** patterns (`=`, `:`, JSON key-value) with **long right-hand strings**.
- Flag **repeated** “random” strings across files (copy-paste keys).
- Treat **short “secrets”** (8–10 chars) as weak regardless of entropy.

False positives to expect: **Git SHAs**, **content hashes**, **compressed assets**, **JWTs in tests** (still consider hygiene).

---

## Known secret formats (representative patterns)

> Use these as **signals**, not legal proofs. Validate with provider revocation and ownership.

### AWS

| Kind | Pattern notes | Notes |
|------|----------------|------|
| **Access Key ID** | Often starts with `AKIA` (long-term IAM); also `ASIA` for temporary | Pair with secret. |
| **Secret Access Key** | 40 chars **base64-ish** (A–Z, a–z, 0–9, +/); historically described as 40 bytes | **Rotate** if leaked. |
| **Session token** | Long string with temp keys | Short-lived but still sensitive. |

### Google

| Kind | Pattern notes |
|------|----------------|
| **API key** | Often `AIza...` for some Google APIs; many formats exist |
| **OAuth client secret** | High entropy string in config |
| **Service account JSON** | `"private_key": "-----BEGIN PRIVATE KEY-----\n..."` |

### GitHub / GitLab / Bitbucket tokens

- **GitHub classic PAT**: `ghp_`, `gho_`, `ghu_`, `ghs_`, `ghr_` prefixes (evolve over time — verify current docs).
- **Fine-grained tokens**: distinct patterns; treat **any** long GitHub-looking token in non-GitHub code as suspect.
- **GitLab**: `glpat-` personal access tokens (prefix may change).

### Slack / Discord / messaging

- Slack tokens often start with `xoxb-`, `xoxp-`, `xapp-` (verify current documentation).
- Discord bot tokens are **high-entropy**; often adjacent to `"token":` in JSON.

### Stripe

- **Live secret key**: `sk_live_...` — **Critical** if valid.
- **Test keys**: `sk_test_...` — lower *production* risk but bad hygiene and may reveal integration details.

### Private keys (PEM)

```
-----BEGIN RSA PRIVATE KEY-----
-----BEGIN OPENSSH PRIVATE KEY-----
-----BEGIN EC PRIVATE KEY-----
-----BEGIN PRIVATE KEY-----   (PKCS#8)
```

**Any** PEM private block in app source, Terraform state in repo, or Docker layers is **Critical** until proven a **dedicated test key** with **no** production trust.

### JWT (as secret material)

- Three **base64url** segments separated by `.` — **often not a secret to protect the server** (public claims) but may **grant access** if stolen.
- **Hs256** with weak `secret` in code is a **crypto failure** (A02) as well as a secret leak.

### Database URLs

- `postgres://user:password@host:5432/db` in **source**, **compose**, or **Helm values** checked into git.
- Password often high-entropy; **context** (`DATABASE_URL`, `connectionString`) is unambiguous.

### Generic API tokens

- **Authorization: Bearer** followed by long opaque string **logged** or **checked into tests**.
- **Basic** auth with static `base64(user:pass)` in code.

---

## `.env`, `.env.*`, and config files

### Exposure vectors

| Vector | What to look for |
|--------|------------------|
| **Committed `.env`** | File in repo root or `config/.env` |
| **Example files with real values** | `.env.example` containing live-looking keys |
| **Docker** | `ENV` instructions echoing secrets; `docker history` reveals layers |
| **Shell history** | Docs telling devs to `export AWS_SECRET_ACCESS_KEY=...` |
| **Frontend bundling** | `REACT_APP_*`, `NEXT_PUBLIC_*`, `VITE_*` — **client-visible** by design |

### Review rules

- **Never** put live secrets in **public** env vars for front-end frameworks.
- `.env.local`, `.env.production` should be **gitignored**; verify **templates** only contain placeholders.
- **Secrets Manager / vault** references in code: ensure **actual retrieval** is not mocked with real values in `docker-compose.override.yml` committed by mistake.

### Patterns in code referencing `.env`

- `process.env.SECRET`, `os.environ["API_KEY"]`, `ENV.fetch("AWS_SECRET_ACCESS_KEY")`
- **Default values** in code: `os.getenv("KEY", "hardcoded-fallback")` — treat fallback as **hardcoded secret**.

---

## Common secret patterns in code (regex-level mental model)

Use targeted search (ripgrep, GitHub secret scanning, trufflehog, gitleaks). Illustrative classes:

| Class | Example shapes | Hint |
|-------|----------------|------|
| **Assignment** | `api_key\s*=\s*['"][A-Za-z0-9/+]{20,}['"]` | Language-adjust. |
| **JSON/YAML** | `"client_secret": "..."`, `password: ...` | Especially in **Kubernetes Secrets** committed as plain YAML — should use sealed-secrets/external secrets. |
| **Headers** | `'Authorization': 'Bearer '` + long literal | |
| **Cloud provider SDKs** | `aws_access_key_id`, `AWS_SECRET_ACCESS_KEY`, `GOOGLE_APPLICATION_CREDENTIALS` pointing to **checked-in JSON** | |
| **Terraform / Pulumi** | `password = "..."` in `.tf`; **state files** with secrets | Often **Critical**. |
| **Mobile** | Keys in `Info.plist`, `google-services.json` — assume **extractable** | Use **attestation** / backend proxying for sensitive calls. |

---

## Low-entropy and structural “secrets” (still dangerous)

- **Default passwords** (`admin/admin`, `postgres/postgres`) — catastrophic if exposed to network.
- **Shared team passwords** in wiki or Slack-export archives (out of repo but in scope for org risk).
- **Webhook signing secrets** short or reused across environments — allows forging callbacks.

---

## When you find a candidate

1. **Classify**: live vs test; prod vs staging; scope of access (IAM policy, Stripe live, etc.).
2. **Rotate** (credentials) or **re-encrypt** (keys) per incident response — **do not** only delete the line in HEAD; **history** matters.
3. **Score** with CVSS: e.g., valid **cloud root-like** key in public repo → often **Critical**; **test Stripe** key → **Low–Medium** for prod impact but still fix.
4. **Prevent recurrence**: pre-commit hooks, secret scanning in CI, least-privilege keys, short TTL.

---

## Positive patterns to recommend

- **Vault / cloud secret manager** with IAM-bound retrieval; **no** plaintext in git.
- **OIDC** from CI to cloud (**no** long-lived `AWS_ACCESS_KEY_ID` in CI env vars where avoidable).
- **Environment injection** at deploy time (K8s Secret, ECS secrets), not baked into images.
- **Git history scanning** on each PR plus **allowlist** only with written justification.

---

## Quick triage table

| Finding | Typical severity |
|---------|------------------|
| Live cloud **root** or broad IAM user keys in repo | **Critical** |
| Live **database admin** password in repo | **Critical** |
| **PEM private key** trusted by production | **Critical** |
| Third-party **live** API key with spend/data impact | **High–Critical** |
| **Read-only** scoped key with limited data | **Medium–High** |
| Test key in **tests/** with sandbox-only scope | **Low** (still remove from public repos when possible) |

Always adjust for **reachability** (public repo, npm package, mobile binary) and **detectability** (search engines, archive.org).
