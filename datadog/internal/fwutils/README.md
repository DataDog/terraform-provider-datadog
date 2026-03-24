# fwutils

Shared utilities for Terraform Plugin Framework resources.

## Write-Only Secret Helpers (`writeonly_helpers.go`)

Helpers for adding write-only attributes (Terraform 1.11+) to resources while maintaining backwards compatibility with plaintext attributes.

### Schema

`CreateWriteOnlySecretAttributes` generates the three-attribute pattern (`<attr>`, `<attr>_wo`, `<attr>_wo_version`) with proper validators (ExactlyOneOf, AlsoRequires, PreferWriteOnlyAttribute). Use `MergeAttributes` to combine with your other attributes:

```go
var secretConfig = fwutils.WriteOnlySecretConfig{
    OriginalAttr:         "value",
    WriteOnlyAttr:        "value_wo",
    TriggerAttr:          "value_wo_version",
    OriginalDescription:  "The secret value.",
    WriteOnlyDescription: "Write-only secret value (not stored in state).",
    TriggerDescription:   "Version trigger for value_wo rotation.",
}

Attributes: fwutils.MergeAttributes(
    fwutils.CreateWriteOnlySecretAttributes(secretConfig),
    map[string]schema.Attribute{
        "id":   schema.StringAttribute{Computed: true},
        "name": schema.StringAttribute{Required: true},
    },
)
```

### CRUD Operations

`WriteOnlySecretHandler` retrieves the secret from whichever mode the user chose:

```go
var secretHandler = &fwutils.WriteOnlySecretHandler{
    Config:                 secretConfig,
    SecretRequiredOnUpdate: false, // true if API requires secret on every update
}

// In Create:
result := secretHandler.GetSecretForCreate(ctx, &req.Config)

// In Update (checks version trigger to detect rotation):
result := secretHandler.GetSecretForUpdate(ctx, &req.Config, &req)

if result.ShouldSetValue {
    body.SetValue(result.Value)
}
```

See `resource_datadog_synthetics_global_variable.go` for a complete working example.
