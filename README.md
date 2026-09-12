# tfkit

`tfkit` provides two review gates for reusable Terraform and one small provider-free module that demonstrates them.

- `tfcheck` checks module layout and rejects undocumented interfaces, missing type declarations, embedded backends, and provisioners.
- `tfplan` reads `terraform show -json` output, prints a compact resource-change summary, and blocks deletes or replacements until a reviewer explicitly acknowledges them.
- The root module validates service and environment identity, produces a stable name prefix, and prevents callers from overriding required ownership tags.

The gates are deliberately separate from `terraform apply`. They do not obtain credentials, change infrastructure, or hide a plan from reviewers.

## Requirements

- Go 1.26 or later
- Terraform 1.16.x
- GNU Make

## Use the commands

Build both commands:

```text
make build
./bin/tfcheck .
terraform show -json saved.tfplan > plan.json
./bin/tfplan plan.json
```

`tfplan` exits with status 1 when a valid plan includes deletion or replacement. After reviewing the named addresses and the original saved plan, acknowledge that specific review step explicitly:

```text
./bin/tfplan -approve-destructive plan.json
```

Malformed input, unsupported action sequences, missing files, and parser errors exit with status 2. Policy findings exit with status 1. This distinction lets CI separate invalid execution from an actionable policy failure.

## Consume the module

The [environment example](examples/environment/main.tf) shows a local adoption path. In another repository, pin the source to a reviewed commit:

```hcl
module "identity" {
  source = "git::ssh://git@github.com/1cbyc/tfkit.git?ref=<full-commit>"

  environment = "prod"
  service     = "orders"
}
```

Feed `module.identity.name_prefix` and `module.identity.tags` into provider-specific resources. The module creates no resources and needs no provider credentials.

## Validate

```text
make ci
```

This runs Go formatting, vet, race-enabled tests, command-level success and rejection checks, Terraform formatting, initialization, validation, and native module tests. It never runs `terraform apply`.

## Design boundaries

See [docs/design.md](docs/design.md) for tradeoffs and [docs/runbook.md](docs/runbook.md) for adoption, diagnosis, compatibility, and rollback.

## License

MIT. See [LICENSE](LICENSE).
