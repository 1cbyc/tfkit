locals {
  normalized_environment = lower(trimspace(var.environment))
  normalized_service     = lower(trimspace(var.service))
  name_prefix            = "${local.normalized_service}-${local.normalized_environment}"
  required_tags = {
    Environment = local.normalized_environment
    ManagedBy   = "terraform"
    Service     = local.normalized_service
  }
  tags = merge(local.required_tags, var.additional_tags)
}

check "required_tags_cannot_be_overridden" {
  assert {
    condition     = alltrue([for key, value in local.required_tags : lookup(local.tags, key, "") == value])
    error_message = "additional_tags must not override Environment, ManagedBy, or Service."
  }
}
