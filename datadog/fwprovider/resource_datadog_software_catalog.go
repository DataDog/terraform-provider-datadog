package fwprovider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/Masterminds/semver/v3"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/customtypes"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/path"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gopkg.in/yaml.v3"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

const catalogPath = "/api/v2/catalog/entity"

var (
	_ resource.ResourceWithConfigure   = &catalogEntityResource{}
	_ resource.ResourceWithImportState = &catalogEntityResource{}
)

type EntityResponse struct {
	Included []Included `json:"included"`
}

type Included struct {
	ID         string              `json:"id"`
	Type       string              `json:"type"`
	Attributes *IncludedAttributes `json:"attributes"`
}

type IncludedAttributes struct {
	Schema    *Entity `json:"schema,omitempty"`
	RawSchema string  `json:"rawSchema,omitempty"`
}

func entityFromYAML(inYAML string) (Entity, error) {
	var e Entity
	err := yaml.Unmarshal([]byte(inYAML), &e)
	return e, err
}

type Reference struct {
	Kind      string
	Name      string
	Namespace string
}

func (r Reference) equal(o Reference) bool {
	return r.Kind == o.Kind && r.Name == o.Name && r.Namespace == o.Namespace
}

func (r Reference) String() string {
	return string(r.Kind) + ":" + r.Namespace + "/" + r.Name
}

type Entity struct {
	APIVersion   string         `yaml:"apiVersion" json:"apiVersion"`
	Kind         string         `yaml:"kind" json:"kind"`
	Metadata     Metadata       `yaml:"metadata" json:"metadata"`
	Spec         map[string]any `yaml:"spec,omitempty" json:"spec,omitempty"`
	Integrations map[string]any `yaml:"integrations,omitempty" json:"integrations,omitempty"`
	Extensions   map[string]any `yaml:"extensions,omitempty" json:"extensions,omitempty"`
	Datadog      map[string]any `yaml:"datadog,omitempty" json:"datadog,omitempty"`
}

func (e *Entity) reference() *Reference {
	if e.Kind == "" {
		return nil
	}
	if e.Metadata == nil {
		return nil
	}
	if e.Metadata.name() == "" {
		return nil
	}
	return &Reference{
		Kind:      e.Kind,
		Name:      e.Metadata.name(),
		Namespace: e.Metadata.namespace(),
	}
}

func (e *Entity) toYAML() (string, error) {
	result, err := yaml.Marshal(e)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

type Metadata map[string]any

func (m Metadata) name() string {
	name, _ := m["name"].(string)
	return name
}

func (m Metadata) namespace() string {
	if namespace, ok := m["namespace"].(string); ok {
		return namespace
	}
	return "default"
}

type catalogEntityResource struct {
	Api  *datadog.APIClient
	Auth context.Context
}

type entityTFState struct {
	EntityYAML customtypes.YAMLStringValue `tfsdk:"entity"`
	ID         types.String                `tfsdk:"id"`
}

func (e *entityTFState) entityYAML() string {
	return e.EntityYAML.ValueString()
}

func (e *entityTFState) update(entityYAML string, ref *Reference) {
	e.EntityYAML = customtypes.YAMLStringValue{StringValue: types.StringValue(entityYAML)}

	if ref != nil {
		e.ID = types.StringValue(ref.String())
	}
}

func NewCatalogEntityResource() resource.Resource {
	return &catalogEntityResource{}
}

func (r *catalogEntityResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.HttpClient
	r.Auth = providerData.Auth
}

func (r *catalogEntityResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "software_catalog"
}

func (r *catalogEntityResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	const modifierDesc = "new entity if ref is updated"
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Software Catalog Entity resource. This can be used to create and manage entities in Datadog Software Catalog using the YAML/JSON definition.",
		Attributes: map[string]schema.Attribute{
			"entity": schema.StringAttribute{
				Description:   "The catalog entity definition.",
				Required:      true,
				Validators:    []validator.String{validEntityYAMLValidator{}},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplaceIf(replacePlanModifier, modifierDesc, modifierDesc)},
				CustomType:    customtypes.YAMLStringType{},
			},
			// Resource ID
			"id": utils.ResourceIDAttribute(),
		},
	}
}

var replacePlanModifier = func(ctx context.Context, request planmodifier.StringRequest, response *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	// we use this function to compute what's considered a new entity, therefore requiring replacement.
	// if the entity's reference is changed, then it's a new entity. eg. from service:myservice to service:otherservice
	// else if the entity's attributes, such as owner, are changed, then we are updating it.
	if request.State.Raw.IsNull() {
		return
	}
	if request.Plan.Raw.IsNull() {
		return
	}
	var oldEntityType customtypes.YAMLStringValue
	request.State.GetAttribute(ctx, path.Root("entity"), &oldEntityType)
	oldEntity, errO := entityFromYAML(oldEntityType.ValueString())
	if errO != nil {
		return
	}
	var newEntityType customtypes.YAMLStringValue
	request.Plan.GetAttribute(ctx, path.Root("entity"), &newEntityType)
	newEntity, errN := entityFromYAML(newEntityType.ValueString())
	if errN != nil {
		return
	}
	// reference is a unique key. if it's changed, then we force to create a new entity.
	if oldEntity.reference() != nil && newEntity.reference() != nil {
		oldRef := oldEntity.reference()
		newRef := newEntity.reference()
		response.RequiresReplace = !oldRef.equal(*newRef)
	}
}

type validEntityYAMLValidator struct {
}

func (v validEntityYAMLValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v validEntityYAMLValidator) MarkdownDescription(_ context.Context) string {
	return "entity must be a valid entity YAML/JSON structure"
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

	// verify apiVersion is v3 or above
	if e.APIVersion == "" {
		resp.Diagnostics.AddAttributeError(req.Path, "", "apiVersion must be non empty (v3 or above)")
		return
	}

	version, err := semver.NewVersion(e.APIVersion)
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "", "apiVersion must be a valid version (v3 or above)")
	}
	if version.Major() < 3 {
		resp.Diagnostics.AddAttributeError(req.Path, "", "apiVersion must be v3 or above")
	}
}

func (r *catalogEntityResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func decodeBase64String(data string, response *resource.ReadResponse) ([]byte, error) {
	bytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (r *catalogEntityResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state entityTFState
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	path := catalogPath + "?include=raw_schema&filter[ref]=" + id
	httpRespByte, httpResp, err := utils.SendRequest(r.Auth, r.Api, "GET", path, nil)

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting entity"))
		return
	}
	var entityResp EntityResponse
	err = json.Unmarshal(httpRespByte, &entityResp)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error unmarshalling entity"))
	}

	if len(entityResp.Included) != 1 || entityResp.Included[0].Attributes == nil || entityResp.Included[0].Attributes.RawSchema == "" {
		err := fmt.Errorf("no entity is found in the response, path=%v response=%v", path, string(httpRespByte))
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving entity"))
		return
	}

	var e Entity
	rawSchema := entityResp.Included[0].Attributes.RawSchema
	encodedBytes, err := decodeBase64String(rawSchema, response)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error unmarshalling entity"))
		return
	}

	err = yaml.Unmarshal(encodedBytes, &e)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error unmarshalling entity"))
		return
	}

	entityYAML, err := e.toYAML()
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error marshalling entity"))
		return
	}
	state.update(entityYAML, e.reference())
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *catalogEntityResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state entityTFState
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	err := r.resourceEntityUpsert(&state, "create")
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error while creating entity"))
		return
	}
	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *catalogEntityResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state entityTFState
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	err := r.resourceEntityUpsert(&state, "update")
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error while updating entity"))
		return
	}
	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *catalogEntityResource) resourceEntityUpsert(state *entityTFState, action string) error {
	entityYAML := state.entityYAML()
	respByte, resp, err := utils.SendRequest(r.Auth, r.Api, "POST", catalogPath, &entityYAML)
	if err != nil || resp.StatusCode != 202 {
		return fmt.Errorf("error while calling Software Catalog to %s entity", action)
	}

	var response EntityResponse
	err = json.Unmarshal(respByte, &response)
	if err != nil {
		return err
	}

	if len(response.Included) != 1 || response.Included[0].Attributes == nil || response.Included[0].Attributes.Schema == nil {
		return errors.New("missing entity in the response")
	}
	e := response.Included[0].Attributes.Schema

	state.update(entityYAML, e.reference())
	return nil
}

func (r *catalogEntityResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state entityTFState
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	_, httpResp, err := utils.SendRequest(r.Auth, r.Api, "DELETE", catalogPath+"/"+id, nil)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting entity"))
		return
	}
}
