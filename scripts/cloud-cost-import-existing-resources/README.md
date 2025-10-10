# Terraform Import Helper Scripts

This directory contains utility scripts to help migrate existing Datadog resources into Terraform.

## generate_import_config.sh

Generates Terraform configuration and import blocks for existing Cloud Cost Management and Tag Pipeline resources.

### Supported Resources

- **datadog_aws_cur_config** - AWS Cost and Usage Report configurations
- **datadog_azure_uc_config** - Azure Usage Cost configurations
- **datadog_gcp_uc_config** - GCP Usage Cost configurations
- **datadog_custom_allocation_rule** - Custom cost allocation rules (when available)
- **datadog_tag_pipeline_ruleset** - Tag pipeline rulesets (when available)

### Prerequisites

1. **Terraform Version**:
   - Terraform 1.5 or later (required for native import blocks)

2. **Environment Variables**:
   ```bash
   export DD_API_KEY="your_api_key"
   export DD_APP_KEY="your_app_key"
   export DD_SITE="datadoghq.com"  # Optional, defaults to datadoghq.com
   ```

3. **Required Tools**:
   - `curl` - for API calls
   - `jq` - for JSON parsing
     - macOS: `brew install jq`
     - Linux: `apt-get install jq` or `yum install jq`

4. **Permissions**:
   - Your API and App keys must have permissions to read:
     - Cloud Cost Management configurations
     - Tag Pipeline configurations

### Usage

```bash
# Basic usage - generates files in current directory
./scripts/generate_import_config.sh

# Specify output directory
./scripts/generate_import_config.sh ./terraform/imports
```

### Output Files

The script generates two files:

1. **generated_resources.tf** - Terraform resource definitions
   ```hcl
   resource "datadog_aws_cur_config" "aws_cur_123456789012" {
     account_id    = "123456789012"
     bucket_name   = "my-cur-bucket"
     # ...
   }
   ```

2. **imports.tf** - Terraform import blocks (Terraform 1.5+)
   ```hcl
   import {
     to = datadog_aws_cur_config.aws_cur_123456789012
     id = "123"
   }

   import {
     to = datadog_azure_uc_config.azure_uc_456789012345
     id = "456"
   }
   ```

### Workflow

1. **Generate the configuration**:
   ```bash
   cd /path/to/terraform/project
   /path/to/terraform-provider-datadog/scripts/generate_import_config.sh .
   ```

2. **Review and customize** `generated_resources.tf`:
   - Rename resources to meaningful names
   - Adjust configurations as needed
   - Remove any resources you don't want to manage
   - If you remove resources, also remove corresponding import blocks from `imports.tf`

3. **Initialize Terraform** (if not already done):
   ```bash
   terraform init
   ```

4. **Import all resources** using Terraform's native import blocks:
   ```bash
   terraform apply
   ```

   This will process all `import` blocks in `imports.tf` and import the resources into your Terraform state. Terraform will ask for confirmation before importing.

5. **Verify the import**:
   ```bash
   terraform plan
   ```

   You should see "No changes" if everything matches.

6. **(Optional) Clean up**:
   ```bash
   rm imports.tf
   ```

   After successful import, you can delete the `imports.tf` file as it's no longer needed.

### Example Session

```bash
$ export DD_API_KEY="***"
$ export DD_APP_KEY="***"
$ ./scripts/generate_import_config.sh ./my-terraform-project

Checking prerequisites...
Fetching existing resources from Datadog...
Checking for AWS CUR configurations...
  Found 2 AWS CUR configuration(s)
Checking for Azure UC configurations...
  Found 1 Azure UC configuration(s)
Checking for GCP UC configurations...
  Found 1 GCP UC configuration(s)
Checking for custom allocation rules...
  Found 3 custom allocation rule(s)
Checking for tag pipeline rulesets...
  Found 2 tag pipeline ruleset(s)

========================================
Generation complete!
========================================

Total resources found: 8

Generated files:
  - ./my-terraform-project/generated_resources.tf
  - ./my-terraform-project/imports.tf

Next steps:
  1. Review and customize ./my-terraform-project/generated_resources.tf
  2. Run: terraform init (if not already done)
  3. Run: terraform apply (to import all resources)
  4. Run: terraform plan (to verify state matches)

$ cd my-terraform-project
$ terraform init
$ terraform apply

Terraform will perform the following actions:

  # datadog_aws_cur_config.aws_cur_123456789012 will be imported
  # datadog_azure_uc_config.azure_uc_987654321098 will be imported
  # ... (and so on)

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

datadog_aws_cur_config.aws_cur_123456789012: Importing...
datadog_aws_cur_config.aws_cur_123456789012: Import complete!
...

Apply complete! Resources: 8 imported, 0 added, 0 changed, 0 destroyed.

$ terraform plan
No changes. Your infrastructure matches the configuration.

$ rm imports.tf  # Clean up after successful import
```

### Troubleshooting

**"Error: DD_API_KEY and DD_APP_KEY environment variables must be set"**
- Make sure both environment variables are exported in your shell

**"Error: jq is not installed"**
- Install jq using your package manager (see Prerequisites)

**"No resources were found"**
- Verify your API keys have the correct permissions
- Check that you're using the correct DD_SITE
- Confirm you have resources in your Datadog account

**Import fails with "resource not found"**
- The resource may have been deleted between generation and import
- Regenerate the configuration files

**Terraform plan shows differences after import**
- After a successful import, `terraform plan` should show "No changes"
- If you see differences, this may indicate:
  - The resource was modified between generation and import - regenerate the configuration
  - A bug in the provider - please report it with the plan output

**"Import blocks are not supported" or similar error**
- You need Terraform 1.5 or later for native import blocks
- Run: `terraform version` to check your version
- Upgrade Terraform or use the legacy `terraform import` commands instead

### Notes

- The script focuses on Cloud Cost Management and Tag Pipeline resources since they lack Terraformer support
- This script uses Terraform's native import blocks (Terraform 1.5+) for efficient parallel imports
- Generated resource names follow the pattern `{resource_type}_{identifier}` (e.g., `aws_cur_123456789012`)
- Always review generated configurations before importing to ensure they match your expectations
- For resources with complex nested structures (like Azure UC configs), verify all nested blocks are correct
- After successful import, the `imports.tf` file can be deleted as it's no longer needed

### Extending the Script

To add support for additional resources:

1. Add a new section following the existing pattern
2. Make the API call to list resources
3. Generate Terraform configuration using jq
4. Add import commands
5. Update the resource counter

Example:
```bash
# ==========================================
# New Resource Type
# ==========================================
echo -e "${YELLOW}Checking for new resources...${NC}"

NEW_RESPONSE=$(api_call "/api/v2/new_resource")
NEW_COUNT=$(echo "$NEW_RESPONSE" | jq '.data | length')

if [ "$NEW_COUNT" -gt 0 ]; then
    echo "  Found $NEW_COUNT new resource(s)"
    # Add generation logic here
    RESOURCE_COUNT=$((RESOURCE_COUNT + NEW_COUNT))
fi
```
