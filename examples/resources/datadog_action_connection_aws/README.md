# AWS Action Connection

Example for creating a Datadog AWS action connection (e.g. for Workflow Automation). Requires an app key [registered for the Actions API](https://registry.terraform.io/providers/DataDog/datadog/latest/docs/resources/app_key_registration).

**Run:**

```bash
export TF_VAR_datadog_api_key="your_api_key"
export TF_VAR_datadog_app_key="your_registered_app_key"
terraform init
terraform plan
terraform apply
```

Override connection name, account ID, or role via `terraform.tfvars` or `TF_VAR_*` if needed (defaults are in `variables.tf`).

**After apply:** Get `external_id` for the IAM role trust policy from the connection in Datadog (Workflow Automation → Connections → open the connection) 

Then use the connection in your workflow (Workflow Automation → workflow → step → Connection).
