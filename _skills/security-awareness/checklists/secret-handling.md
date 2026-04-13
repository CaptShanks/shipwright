# Secret Handling Checklist

## What Counts as a Secret

- API keys, tokens, passwords, passphrases
- Database connection strings with credentials
- Private keys (SSH, TLS, signing)
- OAuth client secrets and refresh tokens
- Encryption keys and initialization vectors
- Service account credentials
- Webhook signing secrets

## Rules

- [ ] **Never commit secrets to source control** -- use `.gitignore`, pre-commit hooks, or tools like `gitleaks`
- [ ] **Never log secrets** -- redact or mask sensitive values in log output
- [ ] **Never include secrets in error messages** -- return generic errors, log details with redaction
- [ ] **Never pass secrets via URL query parameters** -- they appear in server logs, browser history, and referrer headers
- [ ] **Never hardcode secrets** -- not even "temporarily" or "just for testing"
- [ ] **Use environment variables or secret managers** -- AWS Secrets Manager, HashiCorp Vault, 1Password, or equivalent
- [ ] **Rotate secrets regularly** -- and ensure the application handles rotation gracefully
- [ ] **Use short-lived credentials** -- prefer tokens with expiration over long-lived keys
- [ ] **Encrypt secrets at rest** -- don't store plaintext secrets in config files or databases
- [ ] **Limit secret scope** -- each service gets its own credentials with minimum required permissions

## Detection Patterns

Watch for these in code reviews:
- String literals that look like base64-encoded keys (40+ chars of `[A-Za-z0-9+/=]`)
- Variables named `password`, `secret`, `token`, `key`, `credential`, `apikey` assigned to string literals
- Connection strings with embedded `user:pass@host` patterns
- `BEGIN RSA PRIVATE KEY` or `BEGIN OPENSSH PRIVATE KEY` blocks
- AWS patterns: `AKIA[0-9A-Z]{16}`, `[0-9a-zA-Z/+]{40}`
- JWT tokens: `eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`
