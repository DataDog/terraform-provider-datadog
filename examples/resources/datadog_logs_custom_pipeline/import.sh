# For the previously created custom pipelines, you can include them in Terraform with the import operation. Currently, Terraform requires you to explicitly create resources that match the existing pipelines to import them.
terraform import <resource.name> <pipelineID>
