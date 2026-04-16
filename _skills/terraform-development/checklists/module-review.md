# Terraform Module Review Checklist

## Principle: A Module Is an API Contract

Every variable is a public input. Every output is a public return value. Review a module with the same rigor you review a function signature — because changing it later is a breaking change for every consumer.

## 1. Variable Design

- [ ] Every variable has a `description` that explains its purpose to a caller who has never read the module source
- [ ] Every variable has a concrete `type` constraint — no `type = any` unless genuinely required
- [ ] Sensitive inputs are marked with `sensitive = true`
- [ ] Variables with known valid values use `validation` blocks

```hcl
variable "environment" {
  type        = string
  description = "Deployment environment. Controls naming, sizing, and feature flags."
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}
```

- [ ] Related inputs are grouped into typed objects with `optional()` defaults rather than 15+ flat variables
- [ ] Default values are safe and sensible (not permissive — e.g., `publicly_accessible = false`)
- [ ] Variable count is under 15 — if more, the module is doing too much
- [ ] No variable has a default that is environment-specific (no hardcoded account IDs, region names, or CIDR blocks)

## 2. Output Design

- [ ] Every output has a `description`
- [ ] Outputs expose what consumers need, not internal resource structure (`database_endpoint` not `aws_rds_instance_main_endpoint`)
- [ ] Sensitive outputs are marked with `sensitive = true`
- [ ] Outputs do not leak internal implementation details that would break if the module is refactored
- [ ] Outputs needed for cross-module wiring are present (IDs, ARNs, endpoints, security group IDs)

## 3. Resource Patterns

- [ ] `for_each` with maps/sets is used instead of `count` with lists (avoids index-shift cascading destroys)
- [ ] `count` is used only for simple conditional creation (`count = var.enabled ? 1 : 0`)
- [ ] Resources have descriptive names: `aws_iam_role.lambda_execution` not `aws_iam_role.role1`
- [ ] No inline `provisioner` blocks (use dedicated tools for config management)
- [ ] `depends_on` is used only for side-effect dependencies that Terraform cannot infer

## 4. Lifecycle and Safety

- [ ] Stateful resources (databases, S3 buckets with data) have `prevent_destroy = true`
- [ ] Resources that cannot tolerate downtime use `create_before_destroy = true`
- [ ] `ignore_changes` is used sparingly and each usage has a comment explaining why
- [ ] `moved` blocks exist for any resource address changes (prevent destroy-and-recreate)

```hcl
moved {
  from = aws_s3_bucket.data
  to   = module.storage.aws_s3_bucket.main
}
```

- [ ] No use of `-auto-approve` in documented usage examples

## 5. Provider Configuration

- [ ] Module declares `required_providers` with source and version constraints
- [ ] Module declares `required_version` for minimum Terraform version
- [ ] Provider versions use pessimistic constraints (`~> 5.0`, not `>= 5.0`)
- [ ] Child modules do not hardcode provider configuration (region, credentials) — accept via `providers` map
- [ ] Multi-region patterns use provider aliases passed from the root module

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

## 6. State and Backend

- [ ] Root modules configure a remote backend with state locking (S3 + DynamoDB, GCS, Terraform Cloud)
- [ ] State backend has encryption at rest enabled
- [ ] State file scope is appropriate (not a monolithic state for an entire account)
- [ ] `terraform_remote_state` is avoided — prefer data sources or parameter store for cross-stack references
- [ ] No sensitive values stored in state without backend encryption

## 7. Naming and Style

- [ ] All identifiers use `snake_case`
- [ ] Resource names are descriptive and prefixed by what they represent
- [ ] `locals` are used to name computed expressions — no repeated complex inline expressions
- [ ] Files are organized logically: `main.tf`, `variables.tf`, `outputs.tf`, `providers.tf`, `versions.tf`
- [ ] No commented-out resource blocks left in the code

## 8. Documentation

- [ ] Module has a `README.md` with usage examples
- [ ] Required vs optional variables are clear from types and defaults
- [ ] Any non-obvious design decisions are documented (why `ignore_changes` is used, why a resource uses `create_before_destroy`)
- [ ] Example `terraform.tfvars` or usage block is provided

## 9. Testing and Validation

- [ ] `terraform validate` passes without errors
- [ ] `terraform fmt -check` passes (consistent formatting)
- [ ] Policy-as-code scans pass (`checkov`, `tfsec`, or OPA/Rego)
- [ ] Module has `.tftest.hcl` contract tests (Terraform 1.6+) or Terratest/kitchen-terraform tests
- [ ] Plan output has been reviewed for expected changes before any apply

## Anti-Patterns

- **Mega-module** — A single module creating VPCs, databases, compute, DNS, and monitoring with 50+ variables. Split by domain boundary.
- **`type = any` everywhere** — Disables the type system. Errors surface at apply time with cryptic provider messages instead of at plan time.
- **Hardcoded identifiers** — Account IDs, AMI IDs, and CIDR blocks pasted inline. Use data sources and variables.
- **No backend** — Running with local state in a team. State diverges, concurrent applies corrupt resources.
- **Undocumented outputs** — Outputs without descriptions are unusable by consumers who do not read the module source.
- **`count` with lists** — Index shifts cause cascading resource destruction. Use `for_each` with maps.
