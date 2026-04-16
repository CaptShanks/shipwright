# Terraform Security Hardening Checklist

## Principle: Secure by Default, Permissive by Exception

Every resource should start locked down. Permissions, network access, and encryption are opened only as explicitly required and documented. The default path through your module should produce a secure deployment.

## 1. IAM and Access Control

- [ ] IAM policies use explicit `Action` lists — no `"Action": "*"`
- [ ] `Resource` is scoped to specific ARNs or ARN patterns — no `"Resource": "*"` except for genuinely account-level actions (`sts:GetCallerIdentity`)
- [ ] Use `aws_iam_policy_document` data source for policy construction (catches JSON syntax errors at plan time)
- [ ] IAM roles follow least-privilege: only the permissions needed for the specific workload
- [ ] Service roles use condition keys to restrict assumption (`sts:ExternalId`, `aws:SourceAccount`, `aws:SourceArn`)
- [ ] No inline IAM policies on resources — use managed policies attached to roles for auditability
- [ ] Cross-account access uses `assume_role` with scoped trust policies, not shared credentials

```hcl
data "aws_iam_policy_document" "lambda_execution" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = [
      "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${var.function_name}:*"
    ]
  }
}
```

- [ ] No wildcard (`*`) in `Principal` for trust policies unless intentionally public
- [ ] IAM roles have permission boundaries applied where organizational policy requires them

## 2. Encryption

- [ ] S3 buckets have server-side encryption enabled (`aws_s3_bucket_server_side_encryption_configuration`)
- [ ] RDS instances have `storage_encrypted = true`
- [ ] EBS volumes have `encrypted = true`
- [ ] SNS topics and SQS queues use KMS encryption for sensitive data
- [ ] Secrets Manager and SSM Parameter Store entries use KMS CMKs, not default AWS-managed keys, for cross-account or compliance scenarios
- [ ] KMS key policies follow least-privilege (key admins ≠ key users)
- [ ] KMS keys have automatic rotation enabled where supported
- [ ] Data in transit uses TLS — ALB listeners on 443, RDS `require_ssl`, ElastiCache in-transit encryption

## 3. Network Security

- [ ] Security groups follow least-privilege: no `0.0.0.0/0` ingress except for intentionally public-facing load balancers
- [ ] Security group rules specify exact ports, not ranges like `0-65535`
- [ ] Egress rules are restricted where possible (default allows all — tighten for sensitive workloads)
- [ ] Private subnets are used for databases, internal services, and compute that does not need direct internet access
- [ ] Public subnets contain only load balancers, NAT gateways, and bastion hosts
- [ ] VPC flow logs are enabled for audit and troubleshooting
- [ ] No resources have `publicly_accessible = true` unless explicitly required and documented

```hcl
resource "aws_security_group_rule" "app_ingress" {
  type                     = "ingress"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.alb.id
  security_group_id        = aws_security_group.app.id
  description              = "HTTPS from ALB only"
}
```

- [ ] NACLs are used as a secondary layer where compliance requires defense-in-depth
- [ ] VPC endpoints are used for AWS service access from private subnets (S3, DynamoDB, ECR, etc.) to avoid NAT gateway costs and exposure

## 4. Secrets Management

- [ ] No secrets in Terraform variables, `.tfvars` files, or source control
- [ ] Secrets are retrieved via data sources from Secrets Manager, SSM Parameter Store, or Vault
- [ ] Variables that accept sensitive values are marked `sensitive = true`
- [ ] State backend has encryption at rest enabled (secrets appear in state files even when marked sensitive)
- [ ] CI/CD pipeline injects secrets via environment variables sourced from a secrets manager
- [ ] No hardcoded AWS account IDs, access keys, or credentials anywhere in `.tf` files

```hcl
data "aws_secretsmanager_secret_version" "db_password" {
  secret_id = var.db_password_secret_arn
}

resource "aws_db_instance" "main" {
  # ...
  password = data.aws_secretsmanager_secret_version.db_password.secret_string
}
```

- [ ] `.gitignore` excludes `*.tfvars`, `*.tfstate`, `*.tfstate.backup`, and `.terraform/`

## 5. Logging and Monitoring

- [ ] CloudTrail is enabled for API audit logging
- [ ] S3 access logging is enabled for sensitive buckets
- [ ] ALB access logs are enabled and shipped to a secured, separate bucket
- [ ] CloudWatch log groups have retention policies set (not indefinite)
- [ ] Alarms exist for security-relevant events (root login, IAM policy changes, security group modifications)

## 6. Resource-Specific Hardening

### S3

- [ ] Public access block is enabled at the bucket level (`aws_s3_bucket_public_access_block`)
- [ ] Bucket policies deny unencrypted uploads (`aws:SecureTransport`)
- [ ] Versioning is enabled for data integrity and recovery
- [ ] Lifecycle rules expire old versions and incomplete multipart uploads

### RDS

- [ ] `publicly_accessible = false`
- [ ] `deletion_protection = true` for production databases
- [ ] `skip_final_snapshot = false` with a named final snapshot
- [ ] Automated backups are enabled with appropriate retention
- [ ] Database is in a private subnet group

### Lambda

- [ ] Function runs in a VPC if it accesses private resources
- [ ] Execution role has only the permissions the function needs
- [ ] Environment variables do not contain plaintext secrets (use Secrets Manager references)
- [ ] Reserved concurrency or provisioned concurrency is set to prevent runaway invocations

## 7. Terraform Operational Security

- [ ] `terraform apply -auto-approve` is not used in CI without a prior plan review gate
- [ ] State locking is enabled and never bypassed with `-lock=false`
- [ ] State file access is restricted via backend IAM policies (not every developer needs state access)
- [ ] Plan output is reviewed for unexpected resource deletions or replacements before every apply
- [ ] `checkov`, `tfsec`, or OPA/Rego policy scans run in CI on every PR

```bash
# Run security scans in CI
checkov -d . --framework terraform
tfsec .
```

- [ ] Terraform version and provider versions are pinned to prevent supply-chain drift

## Anti-Patterns

- **`"Action": "*"` policies** — Grants full service access. Always enumerate specific actions.
- **`0.0.0.0/0` ingress on port 22** — SSH open to the internet. Use bastion hosts, Systems Manager Session Manager, or VPN.
- **Secrets in `.tfvars`** — Even if the file is gitignored, it exists on developer machines and CI runners unencrypted. Use a secrets manager.
- **Unencrypted state backend** — State files contain every attribute of every resource, including passwords and keys. Encrypt the backend.
- **`publicly_accessible = true` by default** — Databases and services should be private by default. Public access is an explicit, documented exception.
- **Disabled deletion protection** — Production databases without `deletion_protection = true` are one `terraform destroy` away from data loss.
