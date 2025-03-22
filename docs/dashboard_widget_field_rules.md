# Rules for Adding Fields to Dashboard Widgets

## Before Making Changes
1. Check if the field is already implemented in the widget's schema and build functions
2. Review the dashboard.yml API spec to understand the field's structure and requirements
3. Look at similar implementations in other widgets for consistent patterns

## Implementation Steps
1. Start Small: Make changes one at a time and review each change independently
2. Follow Existing Patterns: Use the same field implementation pattern as seen in other widgets
3. Required Components:
   - Add field to widget's schema definition function (e.g., `getTreemapDefinitionSchema`)
   - Add field handling in Terraform-to-Datadog build function (e.g., `buildDatadogTreemapDefinition`)
   - Add field handling in Datadog-to-Terraform build function (e.g., `buildTerraformTreemapDefinition`)
 
## Common Gotchas
1. Don't modify existing working code (e.g., title fields that are already implemented)
2. Ensure field names in schema match the API spec
3. Use existing helper functions when available (e.g., `getWidgetCustomLinkSchema` for custom links)
4. Maintain consistent naming between schema and build functions

## Testing
1. Test both directions of conversion (Terraform to Datadog and vice versa)
2. Verify optional vs required field behavior
3. Test with and without the field present

## Documentation
1. Include field descriptions in schema
2. Follow existing patterns for validation and types
3. Document any special handling or requirements 

## Example Configuration
1. Provide a working example of the Terraform configuration using the new field
2. Include comments explaining each part of the configuration
3. Show both required and optional fields
4. If applicable, show different variations of how the field can be used 

```hcl
resource "datadog_dashboard" "example_dashboard" {
  title       = "Treemap Custom Links Test"
  layout_type = "ordered"

  widget {
    treemap_definition {
      request {
        query {
          metric_query {
            query = "avg:system.cpu.user{*} by {host}"
            name  = "cpu_usage"
          }
        }
        formula {
          formula = "cpu_usage"
        }
      }
      title = "Host CPU Usage"
      custom_links {
        label = "View Host Details"
        link  = "https://app.datadoghq.com/infrastructure/host/{{host}}"  # Correct template variable syntax
      }
      custom_links {
        label = "View Documentation"
        link  = "https://docs.example.com/monitoring"  # Static URL example
      }
    }
  }
}
```

## Pull Request Description
1. Clear title describing the change (e.g., "Add custom_links support to treemap widget")
2. Summary of changes made
3. Link to relevant documentation or specs
4. Impact assessment
   - Breaking changes (if any)
   - Backward compatibility considerations
5. Testing instructions
   - Example configuration
   - Steps to verify functionality
   - Expected behavior
6. Screenshots or examples (if applicable) 