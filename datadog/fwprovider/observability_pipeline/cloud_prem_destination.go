package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CloudPremDestinationModel represents the Terraform model for cloud_prem destination configuration
type CloudPremDestinationModel struct {
}

// ExpandCloudPremDestination converts the Terraform model to the Datadog API model
func ExpandCloudPremDestination(ctx context.Context, id string, inputs types.List, src *CloudPremDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	d := datadogV2.NewObservabilityPipelineCloudPremDestinationWithDefaults()
	d.SetId(id)

	var inputsList []string
	inputs.ElementsAs(ctx, &inputsList, false)
	d.SetInputs(inputsList)

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineCloudPremDestination: d,
	}
}

// FlattenCloudPremDestination converts the Datadog API model to the Terraform model
func FlattenCloudPremDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineCloudPremDestination) *CloudPremDestinationModel {
	if src == nil {
		return nil
	}

	out := &CloudPremDestinationModel{}

	return out
}

// CloudPremDestinationSchema returns the schema for cloud_prem destination
func CloudPremDestinationSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `cloud_prem` destination sends logs to Datadog CloudPrem.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{},
			Blocks:     map[string]schema.Block{},
		},
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(1),
		},
	}
}
