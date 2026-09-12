# Design

## Users and problem

This repository is for engineers who maintain shared Terraform modules or review plans in CI. Terraform already validates language and provider schemas. The missing layer is a small organization-level contract that catches repository drift and makes destructive plan actions visible before apply.

## Policy boundary

`tfcheck` parses HCL rather than searching text. It checks the root module and each direct child under `modules/` that contains Terraform files. A reusable module must have the standard documentation and interface files, a Terraform version constraint, descriptions for variables and outputs, and explicit variable types. Backends belong to consumer root configurations, not reusable modules. Provisioners are rejected because they introduce host-dependent side effects outside provider lifecycle semantics.

The checker does not enforce provider-specific opinions. Network exposure, encryption, backup, and deletion-protection rules belong in `awsbox`, `gcpbox`, or another provider-aware policy layer where resource semantics are available.

## Plan boundary

`tfplan` accepts the stable machine-readable output produced by `terraform show -json`. Input is bounded to 64 MiB, unknown fields are rejected, and unknown action sequences fail closed. A delete or either replacement order is destructive. The approval flag acknowledges review but does not apply the plan.

This gate is not authorization. A delivery workflow must still constrain credentials, require environment approval, retain the original binary plan, and ensure the JSON and applied plan came from the same run.

## Environment contract

The provider-free root module centralizes only values with cross-provider meaning: a validated environment, service identity, name prefix, and required tags. It creates no resources. Provider modules can consume the outputs without inheriting credentials or backend choices.

## Explicit non-goals

The repository does not estimate cost, configure state storage, wrap Terraform execution, provision cloud resources, or claim that static checks prove a plan is safe. Remote-state design is documented in the runbook because its availability and recovery controls depend on the selected platform.
