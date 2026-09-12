run "normalizes_contract" {
  command = plan

  variables {
    environment = "prod"
    service     = "billing-api"
    additional_tags = {
      CostCenter = "platform"
    }
  }

  assert {
    condition     = output.name_prefix == "billing-api-prod"
    error_message = "name prefix did not preserve the validated contract"
  }

  assert {
    condition     = output.tags.CostCenter == "platform" && output.tags.ManagedBy == "terraform"
    error_message = "required and additional tags were not merged"
  }
}

run "rejects_invalid_environment" {
  command = plan

  variables {
    environment = "Production"
    service     = "billing-api"
  }

  expect_failures = [var.environment]
}

run "rejects_required_tag_override" {
  command = plan

  variables {
    environment = "prod"
    service     = "billing-api"
    additional_tags = {
      ManagedBy = "console"
    }
  }

  expect_failures = [var.additional_tags]
}
