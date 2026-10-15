module "identity" {
  source = "../.."

  environment = "staging"
  service     = "checkout"
  additional_tags = {
    CostCenter = "platform"
    Owner      = "commerce"
  }
}

resource "terraform_data" "service_contract" {
  input = {
    name = module.identity.name_prefix
    tags = module.identity.tags
  }
}

output "service_contract" {
  description = "Provider-free resource input proving the module's consumer interface."
  value       = terraform_data.service_contract.output
}
