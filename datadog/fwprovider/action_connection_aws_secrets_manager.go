package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const actionConnectionAWSSecretSuffix = "_aws_secrets_manager"

func replaceActionConnectionIfReferenceRemoved(ctx context.Context, request planmodifier.ObjectRequest, response *objectplanmodifier.RequiresReplaceIfFuncResponse) {
	if request.StateValue.IsNull() || request.StateValue.IsUnknown() || !request.PlanValue.IsNull() {
		return
	}
	steps := request.Path.Steps()
	name := strings.TrimSuffix(string(steps[len(steps)-1].(path.PathStepAttributeName)), actionConnectionAWSSecretSuffix)
	var literal types.String
	response.Diagnostics.Append(request.Plan.GetAttribute(ctx, request.Path.ParentPath().AtName(name), &literal)...)
	response.RequiresReplace = literal.IsNull()
}

type actionConnectionExternalSecretsManagerModel struct {
	ConnectionID types.String `tfsdk:"connection_id"`
	Region       types.String `tfsdk:"region"`
}

type actionConnectionAWSSecretReferenceModel struct {
	SecretID     types.String `tfsdk:"secret_id"`
	Key          types.String `tfsdk:"key"`
	VersionID    types.String `tfsdk:"version_id"`
	VersionStage types.String `tfsdk:"version_stage"`
}

func (r *actionConnectionAWSSecretReferenceModel) apiValue() *datadogV2.ActionConnectionAWSSecretsManagerReference {
	if r == nil {
		return nil
	}
	value := datadogV2.NewActionConnectionAWSSecretsManagerReference(r.SecretID.ValueString(), datadogV2.ACTIONCONNECTIONAWSSECRETSMANAGERTYPE_AWS_SECRETS_MANAGER)
	if !r.Key.IsNull() {
		value.SetKey(r.Key.ValueString())
	}
	if !r.VersionID.IsNull() {
		value.SetVersionId(r.VersionID.ValueString())
	}
	if !r.VersionStage.IsNull() {
		value.SetVersionStage(r.VersionStage.ValueString())
	}
	return value
}

func actionConnectionValue(literal types.String, reference *actionConnectionAWSSecretReferenceModel) datadogV2.HTTPTokenValue {
	if reference != nil {
		return datadogV2.ActionConnectionAWSSecretsManagerReferenceAsHTTPTokenValue(reference.apiValue())
	}
	value := literal.ValueString()
	return datadogV2.ActionConnectionLiteralValueAsHTTPTokenValue(&value)
}

func actionConnectionExternalSettings(model *actionConnectionExternalSecretsManagerModel) datadogV2.ActionConnectionAWSSecretsManager {
	return *datadogV2.NewActionConnectionAWSSecretsManager(model.ConnectionID.ValueString(), model.Region.ValueString(), datadogV2.ACTIONCONNECTIONAWSSECRETSMANAGERTYPE_AWS_SECRETS_MANAGER)
}

func actionConnectionReferenceResourceAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description:   "AWS Secrets Manager reference for this field. To use it, set `external_secrets_manager` on the connection and omit the literal field. Available only for supported credential tokens on connections that run in Datadog's cloud.",
		Optional:      true,
		PlanModifiers: []planmodifier.Object{objectplanmodifier.RequiresReplaceIf(replaceActionConnectionIfReferenceRemoved, "Removing a reference without a literal replacement requires replacing the connection.", "Removing a reference without a literal replacement requires replacing the connection.")},
		Attributes: map[string]schema.Attribute{
			"secret_id":     schema.StringAttribute{Description: "Name or ARN of the secret.", Required: true, Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
			"key":           schema.StringAttribute{Description: "Top-level JSON key to extract. Omit to use the entire secret string.", Optional: true},
			"version_id":    schema.StringAttribute{Description: "Secret version ID. Can't be used with `version_stage`. If neither is set, the current version is used.", Optional: true, Validators: []validator.String{stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("version_stage"))}},
			"version_stage": schema.StringAttribute{Description: "Secret version staging label. Can't be used with `version_id`. If neither is set, the current version is used.", Optional: true},
		},
	}
}

// Derive eligible fields from the paired model attributes so resource, data
// source, and nested HTTP schemas expose the same supported reference fields.
func addActionConnectionAWSSecretSchema(s *schema.Schema) {
	s.Attributes["external_secrets_manager"] = schema.SingleNestedAttribute{
		Description: "AWS Secrets Manager settings shared by credential references. Required when a credential uses an AWS Secrets Manager reference. Optional for literal credentials. Available only on connections that run in Datadog's cloud.", Optional: true,
		PlanModifiers: []planmodifier.Object{objectplanmodifier.RequiresReplaceIf(func(_ context.Context, request planmodifier.ObjectRequest, response *objectplanmodifier.RequiresReplaceIfFuncResponse) {
			response.RequiresReplace = !request.StateValue.IsNull() && request.PlanValue.IsNull()
		}, "Removing external secrets manager settings requires replacing the connection.", "Removing external secrets manager settings requires replacing the connection.")},
		Attributes: map[string]schema.Attribute{
			"connection_id": schema.StringAttribute{Description: "ID of an AWS connection in the same organization used to retrieve secrets.", Required: true, Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
			"region":        schema.StringAttribute{Description: "AWS region containing the secrets.", Required: true, Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
		},
	}
	addActionConnectionReferenceResourceBlocks(s.Blocks, reflect.TypeOf(connectionModel{}))
}

func actionConnectionModelFieldType(model reflect.Type, name string) reflect.Type {
	for model.Kind() == reflect.Pointer || model.Kind() == reflect.Slice {
		model = model.Elem()
	}
	for i := 0; i < model.NumField(); i++ {
		if model.Field(i).Tag.Get("tfsdk") == name {
			return model.Field(i).Type
		}
	}
	return nil
}

func addActionConnectionReferenceResourceAttributes(attributes map[string]schema.Attribute, model reflect.Type) {
	for name, attribute := range attributes {
		literal, ok := attribute.(schema.StringAttribute)
		if !ok || actionConnectionModelFieldType(model, name+actionConnectionAWSSecretSuffix) == nil {
			continue
		}
		literal.Required = false
		literal.Optional = true
		attributes[name] = literal
		attributes[name+actionConnectionAWSSecretSuffix] = actionConnectionReferenceResourceAttribute()
	}
}

func addActionConnectionReferenceResourceBlocks(blocks map[string]schema.Block, model reflect.Type) {
	for name, block := range blocks {
		fieldType := actionConnectionModelFieldType(model, name)
		if fieldType == nil {
			continue
		}
		switch b := block.(type) {
		case schema.SingleNestedBlock:
			addActionConnectionReferenceResourceAttributes(b.Attributes, fieldType)
			addActionConnectionReferenceResourceBlocks(b.Blocks, fieldType)
		case schema.ListNestedBlock:
			addActionConnectionReferenceResourceAttributes(b.NestedObject.Attributes, fieldType)
			addActionConnectionReferenceResourceBlocks(b.NestedObject.Blocks, fieldType)
		}
	}
}

func actionConnectionReferenceDataSourceBlock() dsschema.SingleNestedBlock {
	attributes := make(map[string]dsschema.Attribute)
	for name, attribute := range actionConnectionReferenceResourceAttribute().Attributes {
		attributes[name] = dsschema.StringAttribute{Computed: true, Description: attribute.GetDescription()}
	}
	return dsschema.SingleNestedBlock{Description: "AWS Secrets Manager reference for this field.", Attributes: attributes}
}

func addActionConnectionAWSSecretDataSourceSchema(s *dsschema.Schema) {
	s.Attributes["external_secrets_manager"] = dsschema.SingleNestedAttribute{Computed: true, Description: "AWS Secrets Manager settings shared by credential references.", Attributes: map[string]dsschema.Attribute{
		"connection_id": dsschema.StringAttribute{Computed: true, Description: "ID of the AWS connection used to retrieve secrets."},
		"region":        dsschema.StringAttribute{Computed: true, Description: "AWS region containing the secrets."},
	}}
	addActionConnectionReferenceDataSourceBlocks(s.Blocks, reflect.TypeOf(connectionModel{}))
}

func addActionConnectionReferenceDataSourceFields(attributes map[string]dsschema.Attribute, blocks map[string]dsschema.Block, model reflect.Type) {
	for name := range attributes {
		if actionConnectionModelFieldType(model, name+actionConnectionAWSSecretSuffix) != nil {
			// Keep outputs within legacy blocks as blocks too: Terraform 1.1.5
			// crashes rendering computed nested attributes in a deferred data source.
			blocks[name+actionConnectionAWSSecretSuffix] = actionConnectionReferenceDataSourceBlock()
		}
	}
}

func addActionConnectionReferenceDataSourceBlocks(blocks map[string]dsschema.Block, model reflect.Type) {
	for name, block := range blocks {
		fieldType := actionConnectionModelFieldType(model, name)
		if fieldType == nil {
			continue
		}
		switch b := block.(type) {
		case dsschema.SingleNestedBlock:
			if b.Blocks == nil {
				b.Blocks = make(map[string]dsschema.Block)
			}
			addActionConnectionReferenceDataSourceFields(b.Attributes, b.Blocks, fieldType)
			addActionConnectionReferenceDataSourceBlocks(b.Blocks, fieldType)
			blocks[name] = b
		case dsschema.ListNestedBlock:
			if b.NestedObject.Blocks == nil {
				b.NestedObject.Blocks = make(map[string]dsschema.Block)
			}
			addActionConnectionReferenceDataSourceFields(b.NestedObject.Attributes, b.NestedObject.Blocks, fieldType)
			addActionConnectionReferenceDataSourceBlocks(b.NestedObject.Blocks, fieldType)
			blocks[name] = b
		}
	}
}

func actionConnectionHasReference(value reflect.Value, name string) bool {
	field, ok := lookupTFSDKField(value, name+actionConnectionAWSSecretSuffix)
	return ok && !field.IsNil()
}

// A literal and reference share one API field. Traverse the model to validate
// their exclusivity at every nesting level, including list elements.
func walkActionConnectionReferences(value reflect.Value, p path.Path, visit func(reflect.Value, types.String, *actionConnectionAWSSecretReferenceModel, path.Path)) {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return
		}
		value = value.Elem()
	}
	if value.Kind() == reflect.Slice {
		for i := 0; i < value.Len(); i++ {
			walkActionConnectionReferences(value.Index(i), p.AtListIndex(i), visit)
		}
		return
	}
	if value.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		name := value.Type().Field(i).Tag.Get("tfsdk")
		if name == "" || strings.HasSuffix(name, actionConnectionAWSSecretSuffix) || name == "external_secrets_manager" {
			continue
		}
		if literal, ok := field.Interface().(types.String); ok {
			refField, ok := lookupTFSDKField(value, name+actionConnectionAWSSecretSuffix)
			if !ok {
				continue
			}
			var ref *actionConnectionAWSSecretReferenceModel
			if !refField.IsNil() {
				ref = refField.Interface().(*actionConnectionAWSSecretReferenceModel)
			}
			visit(refField, literal, ref, p.AtName(name))
		} else {
			walkActionConnectionReferences(field, p.AtName(name), visit)
		}
	}
}

func validateActionConnectionSecretReferences(conn connectionResourceModel, response *resource.ValidateConfigResponse) {
	walkActionConnectionReferences(reflect.ValueOf(conn.connectionModel), path.Empty(), func(_ reflect.Value, literal types.String, ref *actionConnectionAWSSecretReferenceModel, p path.Path) {
		if ref == nil {
			return
		}
		if !literal.IsNull() && !literal.IsUnknown() {
			response.Diagnostics.AddAttributeError(p, "Conflicting credential values", "Specify either the literal field or its AWS Secrets Manager reference.")
		}
		if conn.ExternalSecretsManager == nil {
			response.Diagnostics.AddAttributeError(p, "Missing AWS Secrets Manager settings", "An AWS Secrets Manager reference requires external_secrets_manager on the connection.")
		}
		if !ref.VersionID.IsNull() && !ref.VersionID.IsUnknown() && !ref.VersionStage.IsNull() && !ref.VersionStage.IsUnknown() {
			response.Diagnostics.AddAttributeError(p, "Conflicting secret version selectors", "Specify either version_id or version_stage.")
		}
	})
}

func additionalActionConnectionIntegrationData(model connectionResourceModel) (map[string]interface{}, error) {
	data, err := additionalActionConnectionLiteralIntegrationData(model)
	if err != nil {
		return nil, err
	}
	credential := data["credentials"].(map[string]interface{})
	walkActionConnectionReferences(reflect.ValueOf(model.connectionModel), path.Empty(), func(_ reflect.Value, _ types.String, ref *actionConnectionAWSSecretReferenceModel, p path.Path) {
		if ref != nil {
			steps := p.Steps()
			credential[string(steps[len(steps)-1].(path.PathStepAttributeName))] = ref.apiValue()
		}
	})
	return data, nil
}

func actionConnectionAWSSecretReferenceFromAPI(reference *datadogV2.ActionConnectionAWSSecretsManagerReference) *actionConnectionAWSSecretReferenceModel {
	if reference == nil {
		return nil
	}
	return &actionConnectionAWSSecretReferenceModel{
		SecretID:     types.StringValue(reference.SecretId),
		Key:          types.StringPointerValue(reference.Key),
		VersionID:    types.StringPointerValue(reference.VersionId),
		VersionStage: types.StringPointerValue(reference.VersionStage),
	}
}

// Non-HTTP credentials are flat objects whose API field names match the model's
// tfsdk tags. Reuse the credentials decoded by the literal reader; HTTP tokens
// are read directly from the SDK's typed unions alongside their names and types.
func readActionConnectionAWSSecretReferences(model *connectionResourceModel, credentials map[string]json.RawMessage) error {
	var readErr error
	walkActionConnectionReferences(reflect.ValueOf(&model.connectionModel), path.Empty(), func(field reflect.Value, _ types.String, _ *actionConnectionAWSSecretReferenceModel, p path.Path) {
		if readErr != nil {
			return
		}
		steps := p.Steps()
		name := string(steps[len(steps)-1].(path.PathStepAttributeName))
		value := credentials[name]
		if len(value) == 0 || value[0] != '{' {
			return
		}
		var reference datadogV2.ActionConnectionAWSSecretsManagerReference
		if err := json.Unmarshal(value, &reference); err != nil {
			readErr = fmt.Errorf("could not read AWS Secrets Manager reference for %s: %w", p, err)
			return
		}
		if reference.UnparsedObject != nil {
			readErr = fmt.Errorf("invalid AWS Secrets Manager reference for %s", p)
			return
		}
		field.Set(reflect.ValueOf(actionConnectionAWSSecretReferenceFromAPI(&reference)))
	})
	return readErr
}
