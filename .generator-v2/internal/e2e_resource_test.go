package internal_test

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/emit"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/parser"
)

// TestSparseUpdatePreservesPlannedOptionalComputedValue drives the real
// incident-type fixture through the complete offline generation pipeline. It
// pins the three generated-code contracts that make a sparse PATCH response
// safe: Terraform carries the prior Optional+Computed value into the plan,
// Update decodes that complete plan, and updateState assigns the field only
// when the API returned it. Consequently, an omitted description leaves the
// planned model value untouched when the response is overlaid.
func TestSparseUpdatePreservesPlannedOptionalComputedValue(t *testing.T) {
	specPath := filepath.Join("testdata", "fixtures", "resource_incident_type", "openapi.yaml")
	spec, err := parser.LoadSpec(specPath)
	if err != nil {
		t.Fatalf("load incident-type fixture: %v", err)
	}

	var operation *model.Operation
	for _, candidate := range spec.Operations {
		if candidate.Tracking != nil &&
			candidate.Tracking.ArtifactKind == model.ArtifactKindResource &&
			candidate.Tracking.ArtifactName == "incident_type" {
			operation = candidate
			break
		}
	}
	if operation == nil {
		t.Fatal("incident-type fixture has no tracked resource operation")
	}

	artifact, err := model.BuildArtifact(operation)
	if err != nil {
		t.Fatalf("build incident-type artifact: %v", err)
	}
	view, err := emit.BuildResourceView(artifact)
	if err != nil {
		t.Fatalf("build incident-type resource view: %v", err)
	}
	sourceBytes, err := emit.RenderResource(view)
	if err != nil {
		t.Fatalf("render incident-type resource: %v", err)
	}
	source := string(sourceBytes)

	descriptionSchema := section(t, source,
		`"description": schema.StringAttribute{`, "\n\t\t\t},")
	for _, want := range []string{
		`Optional:\s+true`,
		`Computed:\s+true`,
		`stringplanmodifier\.UseStateForUnknown\(\)`,
	} {
		if !regexp.MustCompile(want).MatchString(descriptionSchema) {
			t.Errorf("description schema does not match %q:\n%s", want, descriptionSchema)
		}
	}

	update := section(t, source,
		`func (r *datadogIncidentTypeResource) Update(`,
		`func (r *datadogIncidentTypeResource) Delete(`)
	if !strings.Contains(update, `response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)`) ||
		!strings.Contains(update, `response.Diagnostics.Append(plan.As(ctx, &state, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})...)`) {
		t.Errorf("Update does not seed state from the complete unknown-capable planned value:\n%s", update)
	}

	create := section(t, source,
		`func (r *datadogIncidentTypeResource) Create(`,
		`func (r *datadogIncidentTypeResource) Read(`)
	if !strings.Contains(create, `if state.Description.IsUnknown() {
		state.Description = types.StringNull()
	}`) {
		t.Errorf("Create does not normalize a computed value omitted by a sparse response:\n%s", create)
	}

	updateState := section(t, source,
		`func (r *datadogIncidentTypeResource) updateState(`, "")
	guardedAssignment := `if description, ok := attributes.GetDescriptionOk(); ok && description != nil {
		state.Description = types.StringValue(*description)
	}`
	if !strings.Contains(updateState, guardedAssignment) {
		t.Errorf("updateState does not guard the sparse description overlay:\n%s", updateState)
	}
	if count := strings.Count(updateState, `state.Description =`); count != 1 {
		t.Errorf("updateState has %d description assignments, want only the guarded overlay:\n%s", count, updateState)
	}
}

func section(t *testing.T, source, start, end string) string {
	t.Helper()
	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("generated source missing section start %q", start)
	}
	section := source[startIndex:]
	if end == "" {
		return section
	}
	endIndex := strings.Index(section, end)
	if endIndex < 0 {
		t.Fatalf("generated source section %q missing end %q", start, end)
	}
	return section[:endIndex]
}
