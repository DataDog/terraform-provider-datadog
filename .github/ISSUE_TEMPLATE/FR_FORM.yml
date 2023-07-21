name: Feature Request
description: File a Feature Request
labels: ["feature-request"]
body:
  - type: markdown
    attributes:
      value: |
        Hi there,
        Thank you for opening an feature request. To expedite the process, we suggest submitting a ticket at `https://help.datadoghq.com/` as well.
  - type: textarea
    id: affected-resources
    attributes:
      label: What resources or data sources are affected?
      description: |
        Please list the resources as a list, for example:
          - opc_instance
          - opc_storage_volume
        If this issue appears to affect multiple resources, it may be an issue with Terraform's core, so please mention this.
      placeholder: ex. resource_datadog_monitor
    validations:
      required: true
  - type: textarea
    id: feature-request
    attributes:
      label: Feature Request
      description: Please describe the requested feature and the use case. ex. new resource, data source, or field in existing resource
  - type: textarea
    id: references
    attributes:
      label: References
      description: |
        Are there any other GitHub issues (open or closed) or Pull Requests that should be linked here?
      placeholder: GH-0000
