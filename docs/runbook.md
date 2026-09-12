# Adoption and rollback runbook

## Adopt

1. Pin the module or command source to a reviewed full commit.
2. Run `tfcheck` on the module repository and resolve each finding in source.
3. Run `terraform fmt -check`, `terraform init -backend=false`, `terraform validate`, and `terraform test` before planning.
4. Create a saved plan, render it once with `terraform show -json`, and run `tfplan` on that JSON.
5. Review the human-readable saved plan as well as every destructive address before using `-approve-destructive`.
6. Apply the unchanged saved binary plan through the environment's existing approval boundary.

Never place backend credentials in configuration. Production state should use an authenticated remote backend with encryption, versioning, locking where supported, least-privilege access, audit logs, and a tested restore procedure. Bootstrap and recovery of that backend must remain outside the state it protects.

## Diagnose failures

- Exit 1 from `tfcheck` means the configuration parsed but violated repository policy. The output names each file and rule.
- Exit 1 from `tfplan` means the plan contains deletion or replacement. Inspect the listed addresses and the original plan.
- Exit 2 means invocation or input failed. Regenerate plan JSON using the same supported Terraform version and do not approve malformed input.
- A native Terraform test failure should be fixed in module code or expectations. Do not convert a failing apply test into a plan test merely to bypass behavior.

## Compatibility

The current module requires Terraform 1.16.x or newer within major version 1. Consumers should pin this repository to a full commit. Before changing input types, output names, validation rules, or required tags, test at least one existing consumer and document the migration.

Plan JSON is an integration boundary. Upgrade `tfplan` before upgrading Terraform if a new Terraform release introduces an action sequence or format that the command rejects.

## Roll back

1. Stop before apply if the gate or review result is unexpected.
2. Re-pin the last reviewed repository commit and rerun initialization, validation, tests, and plan generation.
3. Discard stale saved plans. Never apply a plan created from different configuration or state.
4. For already applied infrastructure, use a new reviewed forward plan. Do not edit state manually unless following a provider-specific recovery procedure with a current backup.

Removing the environment module requires replacing its output references with equivalent literal or local values first. Confirm that the resulting plan does not rename or replace provider resources unexpectedly.
