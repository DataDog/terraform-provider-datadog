variable "datadog_api_key" {
  type        = string
  sensitive   = true
  description = "Datadog API key. Set via TF_VAR_datadog_api_key or terraform.tfvars."
}

variable "datadog_app_key" {
  type        = string
  sensitive   = true
  description = "Datadog app key (must be registered for Actions API). Set via TF_VAR_datadog_app_key or terraform.tfvars."
}

variable "connection_name" {
  type    = string
  default = "My AWS Connection"
}

variable "aws_account_id" {
  type = string
}

variable "aws_role_name" {
  type = string
}
