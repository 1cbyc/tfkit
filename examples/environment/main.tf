module "identity" {
  source = "../.."

  environment = "staging"
  service     = "orders"
  additional_tags = {
    CostCenter = "platform"
    Owner      = "commerce"
  }
}

output "name_prefix" {
  description = "Validated prefix that provider-specific resources can consume."
  value       = module.identity.name_prefix
}

output "tags" {
  description = "Validated tags that provider-specific resources can consume."
  value       = module.identity.tags
}
