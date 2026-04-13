# Error Message Security Checklist

## Principle: Separate Internal and External Error Information

Users get a generic, helpful message. Engineers get the full details in logs.

## What to Never Expose to Users

- [ ] **Stack traces** -- reveals internal structure, file paths, library versions
- [ ] **SQL queries** -- reveals schema, table names, column names
- [ ] **File system paths** -- reveals server structure and deployment layout
- [ ] **Internal IP addresses or hostnames** -- reveals network topology
- [ ] **Library/framework versions** -- enables targeted exploit searches
- [ ] **Database error messages** -- reveals query structure and schema details
- [ ] **Configuration details** -- reveals security settings and architecture
- [ ] **Diff between valid/invalid states** -- e.g., "user not found" vs "wrong password" reveals valid usernames

## Pattern: Structured Error Handling

```
User sees:   "Something went wrong. Please try again. (ref: abc123)"
Log contains: "ref=abc123 err=pq: relation \"users\" does not exist query=SELECT..."
```

The reference ID links the user-facing message to the detailed log entry without exposing internals.

## Authentication Errors

Use identical messages for all authentication failure modes:
- Wrong username → "Invalid credentials"
- Wrong password → "Invalid credentials"
- Account locked → "Invalid credentials"
- Account disabled → "Invalid credentials"

Never reveal which part of the authentication failed. This prevents username enumeration.

## Rate Limiting Errors

Don't reveal rate limit thresholds in error messages. "Too many requests, try again later" is sufficient. Revealing "You have 3 attempts remaining" helps attackers calibrate brute-force attacks.
