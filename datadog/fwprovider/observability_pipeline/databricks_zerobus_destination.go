package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DatabricksZerobusAuthModel represents the OAuth client credentials for the DatabricksZerobus destination.
type DatabricksZerobusAuthModel struct {
	ClientId        types.String `tfsdk:"client_id"`
	ClientSecretKey types.String `tfsdk:"client_secret_key"`
}

// DatabricksZerobusDestinationModel represents the Terraform model for the DatabricksZerobus destination.
type DatabricksZerobusDestinationModel struct {
	IngestionEndpoint    types.String                 `tfsdk:"ingestion_endpoint"`
	TableName            types.String                 `tfsdk:"table_name"`
	UnityCatalogEndpoint types.String                 `tfsdk:"unity_catalog_endpoint"`
	Auth                 []DatabricksZerobusAuthModel `tfsdk:"auth"`
}

// DatabricksZerobusDestinationSchema returns the schema for the DatabricksZerobus destination.
func DatabricksZerobusDestinationSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `databricks_zerobus` destination sends logs to Databricks via the Zerobus ingestion API.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"ingestion_endpoint": schema.StringAttribute{
					Required:    true,
					Description: "The Databricks Zerobus ingestion endpoint URL.",
				},
				"table_name": schema.StringAttribute{
					Required:    true,
					Description: "The name of the Databricks table to ingest logs into.",
				},
				"unity_catalog_endpoint": schema.StringAttribute{
					Required:    true,
					Description: "The Databricks Unity Catalog endpoint URL.",
				},
			},
			Blocks: map[string]schema.Block{
				"auth": schema.ListNestedBlock{
					Description: "OAuth client credentials used to authenticate with Databricks.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"client_id": schema.StringAttribute{
								Required:    true,
								Description: "The OAuth client ID used to authenticate with Databricks.",
							},
							"client_secret_key": schema.StringAttribute{
								Optional:    true,
								Description: "The name of the secret or environment variable holding the OAuth client secret. Defaults to `DESTINATION_DATABRICKS_ZEROBUS_OAUTH_CLIENT_SECRET`.",
							},
						},
					},
				},
			},
		},
	}
}

// ExpandDatabricksZerobusDestination converts the Terraform model to the API model.
func ExpandDatabricksZerobusDestination(ctx context.Context, id string, inputs types.List, src *DatabricksZerobusDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineDatabricksZerobusDestinationWithDefaults()
	dest.SetId(id)

	var inputsList []string
	inputs.ElementsAs(ctx, &inputsList, false)
	dest.SetInputs(inputsList)

	dest.SetIngestionEndpoint(src.IngestionEndpoint.ValueString())
	dest.SetTableName(src.TableName.ValueString())
	dest.SetUnityCatalogEndpoint(src.UnityCatalogEndpoint.ValueString())

	if len(src.Auth) > 0 {
		auth := datadogV2.NewObservabilityPipelineDatabricksZerobusDestinationAuthWithDefaults()
		auth.SetClientId(src.Auth[0].ClientId.ValueString())
		if !src.Auth[0].ClientSecretKey.IsNull() {
			auth.SetClientSecretKey(src.Auth[0].ClientSecretKey.ValueString())
		}
		dest.SetAuth(*auth)
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineDatabricksZerobusDestination: dest,
	}
}

// FlattenDatabricksZerobusDestination converts the API model to the Terraform model.
func FlattenDatabricksZerobusDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineDatabricksZerobusDestination) *DatabricksZerobusDestinationModel {
	if src == nil {
		return nil
	}

	model := &DatabricksZerobusDestinationModel{
		IngestionEndpoint:    types.StringValue(src.GetIngestionEndpoint()),
		TableName:            types.StringValue(src.GetTableName()),
		UnityCatalogEndpoint: types.StringValue(src.GetUnityCatalogEndpoint()),
	}

	if auth, ok := src.GetAuthOk(); ok {
		authModel := DatabricksZerobusAuthModel{
			ClientId: types.StringValue(auth.GetClientId()),
		}
		if v, ok := auth.GetClientSecretKeyOk(); ok {
			authModel.ClientSecretKey = types.StringValue(*v)
		}
		model.Auth = []DatabricksZerobusAuthModel{authModel}
	}

	return model
}
