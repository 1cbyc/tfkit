variable "environment" {
  description = "Short deployment environment identifier used in names and required tags."
  type        = string
  nullable    = false

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{1,14}$", var.environment))
    error_message = "environment must be 2 to 15 lowercase letters, numbers, or hyphens and start with a letter."
  }
}

variable "service" {
  description = "Short service identifier used in names and required tags."
  type        = string
  nullable    = false

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{1,29}$", var.service))
    error_message = "service must be 2 to 30 lowercase letters, numbers, or hyphens and start with a letter."
  }
}

variable "additional_tags" {
  description = "Extra non-authoritative tags merged with the required environment, service, and ownership tags."
  type        = map(string)
  default     = {}
  nullable    = false

  validation {
    condition     = length(setintersection(toset(keys(var.additional_tags)), toset(["Environment", "ManagedBy", "Service"]))) == 0
    error_message = "additional_tags must not contain Environment, ManagedBy, or Service."
  }
}
