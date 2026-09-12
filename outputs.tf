output "name_prefix" {
  description = "Validated service and environment prefix for provider-specific resource names."
  value       = local.name_prefix
}

output "tags" {
  description = "Required identity tags merged with caller-provided non-authoritative tags."
  value       = local.tags
}
