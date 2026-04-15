---
name: terraform-development
description: >-
  Terraform and HCL idioms, patterns, and conventions for writing production-grade
  infrastructure as code. Use when writing, reviewing, or refactoring Terraform
  configurations, designing module architectures, managing state, or applying
  provider best practices.
license: MIT
metadata:
  author: CaptShanks
  version: "1.0.0"
---

# Terraform Development

This skill encodes how experienced infrastructure teams write Terraform that is readable, composable, safe to apply, and stable across upgrades. Terraform favors **declarative intent, explicit dependencies, and small blast-radius changes** over clever abstractions.

## HCL Style and Variable Design

### Variables are your module's API contract

Every `variable` block is a public input. Treat it the same way you treat a function signature in application code.

- Use `type` constraints to prevent misconfiguration at plan time, not apply time.
- Prefer concrete types (`string`, `number`, `bool`, `list(string)`, `map(string)`) over `any`. The `any` type silences the type checker and passes garbage through to the provider where errors are opaque.
- Use `optional()` with defaults in object types to keep caller ergonomics clean while maintaining strong contracts:

```hcl
variable "settings" {
  type = object({
    retention_days = optional(number, 30)
    encrypt        = optional(bool, true)
    tags           = optional(map(string), {})
  })
}
```

### Validation blocks

Use `validation` blocks to encode business rules that types alone cannot express:

```hcl
variable "environment" {
  type = string
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}
```

Validation is cheap, runs at plan time, and saves a failed apply plus a state rollback.

### Naming conventions

- Variables: `snake_case`, descriptive, no abbreviations. `vpc_cidr_block` not `cidr`.
- Resources: `snake_case`, prefixed by what they represent. `aws_iam_role.lambda_execution` not `aws_iam_role.role1`.
- Locals: use `locals` to name computed expressions; do not inline complex expressions repeatedly.
- Outputs: name by what the consumer needs, not internal resource structure. `database_endpoint` not `aws_rds_instance_main_endpoint`.

## Module Architecture

### Root vs child modules

- **Root modules** are deployable units tied to a state file. They compose child modules, set variable values, and configure backends.
- **Child modules** are reusable building blocks. They should be stateless, backend-agnostic, and provider-agnostic (accept provider configuration via `required_providers`, not hardcoded).

### Composition over inheritance

Terraform has no inheritance. Compose small, focused modules:

- A `vpc` module creates networking primitives.
- A `cluster` module takes VPC outputs and creates compute.
- A root module wires them together.

Avoid "god modules" that create an entire environment in one resource graph. They are slow to plan, hard to test, and risky to apply.

### Module versioning and pinning

- Pin module sources to exact versions or narrow ranges: `source = "git::https://...?ref=v2.1.0"` or `version = "~> 2.1"`.
- Never use `ref=main` in production. A broken commit on main becomes a broken infrastructure apply with no rollback path.
- For registry modules, use the `version` constraint. For git modules, use tag refs.

### Module interface design

- Keep the variable count under 15. If a module needs more, it is doing too much.
- Group related inputs into typed objects rather than flat variables.
- Every output should have a `description`. Outputs are the module's return values; undocumented outputs are unusable.

## State Management

### Remote backends

Always use a remote backend in shared environments. Local state files are a single point of failure and merge conflict source.

- S3 + DynamoDB (locking) is the AWS standard.
- GCS, Azure Blob, Terraform Cloud, and Consul are alternatives.
- Enable encryption at rest on the backend storage.

### State locking

State locking prevents concurrent applies from corrupting state. Never disable locking (`-lock=false`) in automation. If a lock is stuck, investigate before force-unlocking.

### Workspaces

Workspaces multiplex a single configuration across environments. They work for simple cases but break down when environments differ structurally (different regions, different account shapes). Prefer separate root modules per environment for non-trivial divergence.

### `import` and `moved` blocks

- Use `import` blocks (Terraform 1.5+) to bring existing infrastructure under management without recreating it.
- Use `moved` blocks to refactor resource addresses without destroying and recreating. This is critical for renaming resources or restructuring modules.

```hcl
moved {
  from = aws_s3_bucket.data
  to   = module.storage.aws_s3_bucket.main
}
```

### `terraform_remote_state` vs data sources

Prefer **data sources** over `terraform_remote_state` for cross-stack references. `terraform_remote_state` exposes the entire output map and creates a tight coupling to the source state structure. Data sources (e.g., `aws_vpc`, `aws_ssm_parameter`) are narrowly scoped and resilient to upstream refactors.

## Provider Configuration

### Version constraints

Pin providers in `required_providers` with pessimistic constraints:

```hcl
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  required_version = ">= 1.5"
}
```

Run `terraform init -upgrade` deliberately, not accidentally.

### Provider aliases

Use aliases for multi-region or multi-account patterns:

```hcl
provider "aws" {
  alias  = "us_west_2"
  region = "us-west-2"
}

provider "aws" {
  alias  = "eu_west_1"
  region = "eu-west-1"
}
```

Pass aliased providers to child modules via `providers` map. Never hardcode regions inside child modules.

### `assume_role`

For cross-account access, use `assume_role` in the provider block. The executing identity should have minimal permissions; the assumed role carries the deployment permissions. This is standard for multi-account AWS landing zones.

## Resource Patterns

### `for_each` vs `count`

- **`for_each`** with maps or sets is preferred. Resources are keyed by meaningful identifiers, so adding or removing an element does not shift indices and force recreation of unrelated resources.
- **`count`** is acceptable for simple conditional creation (`count = var.enabled ? 1 : 0`). Avoid `count` with lists—index shifts cause cascading destroys.

### Lifecycle meta-arguments

- `create_before_destroy`: use for resources that cannot tolerate downtime during replacement (load balancers, DNS records, launch templates).
- `prevent_destroy`: use for stateful resources you never want accidentally deleted (databases, S3 buckets with data). Pair with documentation explaining why.
- `ignore_changes`: use sparingly for attributes managed outside Terraform (e.g., ASG desired count managed by autoscaling). Document what is ignored and why. Overuse of `ignore_changes` turns Terraform into a partial source of truth.

### `depends_on`

Terraform infers most dependencies from reference expressions. Use explicit `depends_on` only for **side-effect dependencies** that Terraform cannot see (e.g., an IAM policy must exist before a Lambda can assume a role, but the Lambda resource does not reference the policy directly).

### Data sources for discovery

Use data sources to look up existing infrastructure rather than passing IDs as variables:

```hcl
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
```

This eliminates hardcoded account IDs and region strings, which are the most common source of copy-paste drift across environments.

## Security

### Least-privilege IAM

- Write IAM policies with explicit `Action` lists, not `"Action": "*"`.
- Scope `Resource` to specific ARNs or ARN patterns. `"Resource": "*"` is acceptable only for actions that genuinely operate at the account level (e.g., `sts:GetCallerIdentity`).
- Use `aws_iam_policy_document` data source for programmatic policy construction; it catches JSON syntax errors at plan time.

### Sensitive variables

Mark variables containing secrets as `sensitive = true`. Terraform redacts them from plan output and logs. But remember: sensitive values still appear in state files—encrypt your state backend.

### No hardcoded identifiers

Never hardcode AWS account IDs, ARNs, or environment-specific values. Use data sources (`aws_caller_identity`, `aws_region`), variables, or SSM Parameter Store lookups. Hardcoded values are the fastest path to "it works in dev but destroys prod."

### Secrets management

Do not store secrets in Terraform variables or `.tfvars` files committed to source control. Use:
- AWS Secrets Manager or SSM Parameter Store with `data` source lookups.
- Vault provider for HashiCorp Vault.
- Environment variables for CI/CD pipelines, injected from a secrets manager.

## Testing and Validation

### `terraform validate`

`terraform validate` checks syntax and internal consistency without provider access. Run it as the first CI gate—it is fast and catches typos, missing required arguments, and type mismatches.

### `terraform plan` as a test

A successful `terraform plan` against a real backend is the most reliable test. Parse the plan output or use `terraform show -json` for programmatic assertions on expected changes. In CI, run plan on every PR and post the summary as a PR comment.

### `terraform test` framework

Terraform 1.6+ includes a native test framework (`.tftest.hcl` files). Use it for module-level contract tests: apply a module with known inputs and assert outputs match expectations. Tests run real applies in an ephemeral workspace and destroy afterwards.

### Policy-as-code

Use `checkov`, `tfsec`, or OPA/Rego policies to enforce organizational guardrails (e.g., "all S3 buckets must have encryption enabled"). Run these in CI alongside `terraform plan`. They catch compliance violations before apply, not after an audit.

## Common Anti-Patterns

### Mega-modules

A single module that creates VPCs, databases, compute, DNS, and monitoring. It has 50+ variables, takes 10 minutes to plan, and a change to a tag triggers review of 200 resources. Split by domain boundary.

### `type = any` everywhere

Disables the type system. Errors surface at apply time with cryptic provider messages instead of at plan time with clear Terraform diagnostics. Use concrete types; use `optional()` for flexibility.

### State file coupling

Using `terraform_remote_state` to chain five stacks creates a fragile dependency graph where a state format change in stack A breaks stacks B through E. Prefer data sources or parameter store for cross-stack communication.

### Hardcoded values

Account IDs, region names, AMI IDs, and CIDR blocks pasted inline. These work in one environment and break everywhere else. Parameterize through variables and data sources.

### No backend configuration

Running with local state in a team. State diverges, concurrent applies corrupt resources, and there is no lock. Configure a remote backend on day one.

### Monolithic state files

A single state file for an entire AWS account with hundreds of resources. Plan takes minutes, blast radius is unbounded, and state lock contention blocks the team. Split by service, environment, or domain.

### `terraform apply -auto-approve` in CI without plan review

Skipping the plan review step in CI means changes go to production without human or automated verification. Always separate plan and apply stages; gate apply on plan approval.

## When in Doubt

Prefer **small, focused modules** with typed interfaces, remote state with locking, and `for_each` over `count`. Run `validate` and `plan` in CI on every change. Treat your Terraform code with the same rigor as application code: review it, test it, version it.
