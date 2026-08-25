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

### Secrets in Nested Blocks

The three attributes live at the resource root by default. Set `ParentBlocks` to scope them under static nested blocks, so that both the generated validators and the handler lookups target the right paths:

```go
var secretConfig = fwutils.WriteOnlySecretConfig{
    OriginalAttr:  "password",
    WriteOnlyAttr: "password_wo",
    TriggerAttr:   "password_wo_version",
    ParentBlocks:  []string{"authentication", "basic"},
    // ...descriptions
}
```

The attributes are then addressed as `authentication.basic.password`. Only single-nested blocks are supported: attributes inside list- or set-nested blocks need index-aware paths, which a list of block names cannot express.

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
