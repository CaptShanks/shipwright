---
name: security-awareness
description: >-
  Baseline security knowledge every engineer needs -- input validation, secret
  handling, least privilege, and defensive coding. Use when writing or reviewing
  code that handles user input, secrets, authentication, authorization, or any
  data from external sources.
license: MIT
metadata:
  author: CaptShanks
  version: "1.0.0"
---

# Security Awareness

This skill represents the security baseline that every engineer -- not just security specialists -- must internalize. Security is not a phase or a separate concern; it is a property of well-written code.

## Core Principles

### Defense in Depth
Never rely on a single layer of protection. Validate inputs even if the caller "should" have validated them. Sanitize outputs even if the data "should" be clean. Assume every boundary is a trust boundary.

### Least Privilege
Code should request and hold only the minimum permissions needed for the current operation. Don't open files with write access if you only need to read. Don't use admin credentials for routine queries. Don't grant broad network access when a specific endpoint suffices.

### Zero Trust Data
Treat all data from external sources as untrusted until explicitly validated:
- User input (forms, query params, headers, cookies)
- Data from other services (APIs, databases, message queues)
- Environment variables and configuration files
- File contents read from disk
- Data deserialized from any format (JSON, YAML, protobuf)

### Fail Securely
When an operation fails, the system should default to a secure state:
- Deny access on authentication failure, never grant
- Return generic error messages to users, log details server-side
- Close connections and release resources on error
- Don't leave partial state that could be exploited

## What to Always Check

1. **Inputs are validated** -- type, length, range, format, and character set
2. **Outputs are encoded** -- context-appropriate encoding for HTML, SQL, shell, URLs
3. **Secrets are never exposed** -- not in logs, error messages, URLs, or source code
4. **Authentication is verified** before any privileged operation
5. **Authorization is checked** at the resource level, not just the endpoint level
6. **Dependencies are from trusted sources** and versions are pinned
7. **Error messages don't leak internals** -- no stack traces, file paths, or SQL in user-facing errors
8. **Cryptographic operations use standard libraries** -- never roll your own crypto

## Common Mistakes

- Logging request bodies that may contain passwords or tokens
- Using string concatenation instead of parameterized queries
- Trusting client-side validation as the only validation
- Storing secrets in environment variables without encryption at rest
- Catching and silencing security-relevant exceptions
- Using `http` when `https` is available
- Hardcoding credentials "just for testing" and forgetting to remove them
