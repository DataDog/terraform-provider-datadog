# AWS Action Connection

Create a Datadog action connection for AWS (e.g. Workflow Automation). Uses account `087496745774` and role `datadog-aws-integration-role-zeina`.

**Requires:** Registered application key for the Actions API.

## Steps

1. **Set credentials via env** (no secrets in files)
   ```bash
   export TF_VAR_datadog_api_key="your_api_key"
   export TF_VAR_datadog_app_key="your_registered_app_key"
   ```
   Terraform reads variables from `TF_VAR_<name>` automatically.

2. **Optional:** copy `terraform.tfvars.example` to `terraform.tfvars` to override connection name, account ID, or role. Omit if defaults are fine.

3. **Apply**
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

4. **Get IAM values**
   ```bash
   terraform output external_id
   terraform output principal_id
   ```
   Add these to the IAM role trust policy for `datadog-aws-integration-role-zeina` in AWS.

5. **Use in workflow**  
   In Datadog: Workflow Automation → your workflow → step → Connection → select this connection.
