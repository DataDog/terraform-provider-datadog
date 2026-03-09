variable "datadog_api_key" {
  type      = string
  sensitive = true
}

variable "datadog_app_key" {
  type      = string
  sensitive = true
}

variable "connection_name" {
  type    = string
  default = "My AWS Connection"
}

variable "aws_account_id" {
  type    = string
  default = "087496745774"
}

variable "aws_role_name" {
  type    = string
  default = "datadog-aws-integration-role-zeina"
}
