package validators

import (
	"context"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"gopkg.in/yaml.v3"
)

// the entity type is a struct from the datadog_software_catalog
// it was duplicated here to avoid looping dependencies
type entity struct {
	APIVersion   string         `yaml:"apiVersion" json:"apiVersion"`
	Kind         string         `yaml:"kind" json:"kind"`
	Metadata     map[string]any `yaml:"metadata" json:"metadata"`
	Spec         map[string]any `yaml:"spec,omitempty" json:"spec,omitempty"`
	Integrations map[string]any `yaml:"integrations,omitempty" json:"integrations,omitempty"`
	Extensions   map[string]any `yaml:"extensions,omitempty" json:"extensions,omitempty"`
	Datadog      map[string]any `yaml:"datadog,omitempty" json:"datadog,omitempty"`
}

// entityFromYAML is a function from the datadog_software_catalog
// it was duplicated here to avoid looping dependencies
func entityFromYAML(inYAML string) (entity, error) {
	var e entity
	err := yaml.Unmarshal([]byte(inYAML), &e)
	return e, err
}

var _ validator.String = validEntityYAMLValidator{}

type validEntityYAMLValidator struct {
}

func (v validEntityYAMLValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v validEntityYAMLValidator) MarkdownDescription(_ context.Context) string {
	return "Entity must be a valid entity YAML/JSON structure"
}

func (v validEntityYAMLValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	inYAML := req.ConfigValue.ValueString()
	e, err := entityFromYAML(inYAML)
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "", "entity must be a valid entity YAML/JSON structure")
		return
	}

	// verify apiVersion is v3 or above, or backstage.io/v1alpha1
	if e.APIVersion == "" {
		resp.Diagnostics.AddAttributeError(req.Path, "", "apiVersion must be non empty (v3 or above, or backstage.io/v1alpha1)")
		return
	}
	if e.APIVersion == "backstage.io/v1alpha1" {
		return
	}

	// validate as semantic version (v3 or above)
	version, err := semver.NewVersion(e.APIVersion)
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "", "apiVersion must be a valid version (v3 or above, or backstage.io/v1alpha1)")
		return
	}
	if version.Major() < 3 {
		resp.Diagnostics.AddAttributeError(req.Path, "", "apiVersion must be v3 or above, or backstage.io/v1alpha1")
	}
}

func ValidEntityYAMLValidator() validator.String {
	return validEntityYAMLValidator{}
}
