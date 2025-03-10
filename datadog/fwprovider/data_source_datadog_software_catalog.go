package fwprovider

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ datasource.DataSource = &datadogSoftwareCatalogDataSource{}
)

type CatalogEntityModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Namespace   types.String `tfsdk:"namespace"`
	Owner       types.String `tfsdk:"owner"`
	Kind        types.String `tfsdk:"kind"`
	Tags        types.List   `tfsdk:"tags"`
}

type datadogSoftwareCatalogDataSourceModel struct {
	// Query Parameters
	ID                    types.String `tfsdk:"id"`
	FilterID              types.String `tfsdk:"filter_id"`
	FilterName            types.String `tfsdk:"filter_name"`
	FilterRef             types.String `tfsdk:"filter_ref"`
	FilterExcludeSnapshot types.String `tfsdk:"filter_exclude_snapshot"`
	FilterKind            types.String `tfsdk:"filter_kind"`
	FilterOwner           types.String `tfsdk:"filter_owner"`
	FilterRelationType    types.String `tfsdk:"filter_relation_type"`

	// Results
	Entities []*CatalogEntityModel `tfsdk:"entities"`
}

func NewDatadogSoftwareCatalogDataSource() datasource.DataSource {
	return &datadogSoftwareCatalogDataSource{}
}

type datadogSoftwareCatalogDataSource struct {
	Api  *datadogV2.SoftwareCatalogApi
	Auth context.Context
}

func (d *datadogSoftwareCatalogDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetSoftwareCatalogApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogSoftwareCatalogDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "software_catalog"
}

func (d *datadogSoftwareCatalogDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to list software catalog entities to use in other resources.",
		Attributes: map[string]schema.Attribute{
			// Datasource Parameters
			"id": utils.ResourceIDAttribute(),
			"filter_id": schema.StringAttribute{
				Description: "Filter entities by UUID.",
				Optional:    true,
			},
			"filter_name": schema.StringAttribute{
				Description: "Filter entities by name.",
				Optional:    true,
			},
			"filter_ref": schema.StringAttribute{
				Description: "Filter entities by reference.",
				Optional:    true,
			},
			"filter_exclude_snapshot": schema.StringAttribute{
				Description: "Filter entities by excluding snapshotted entities.",
				Optional:    true,
			},
			"filter_kind": schema.StringAttribute{
				Description: "Filter entities by kind.",
				Optional:    true,
			},
			"filter_owner": schema.StringAttribute{
				Description: "Filter entities by owner.",
				Optional:    true,
			},
			"filter_relation_type": schema.StringAttribute{
				Description: "Filter entities by relation type.",
				Optional:    true,
				Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV2.NewRelationTypeFromValue)},
			},

			// Computed values
			"entities": schema.ListAttribute{
				Description: "List of entities",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":           types.StringType,
						"name":         types.StringType,
						"display_name": types.StringType,
						"namespace":    types.StringType,
						"owner":        types.StringType,
						"kind":         types.StringType,
						"tags":         types.ListType{ElemType: types.StringType},
					},
				},
			},
		},
	}
}

func (d *datadogSoftwareCatalogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datadogSoftwareCatalogDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	optionalParams := datadogV2.NewListCatalogEntityOptionalParameters()
	if !state.FilterID.IsNull() {
		optionalParams.WithFilterId(state.FilterID.ValueString())
	}
	if !state.FilterName.IsNull() {
		optionalParams.WithFilterName(state.FilterName.ValueString())
	}
	if !state.FilterExcludeSnapshot.IsNull() {
		optionalParams.WithFilterExcludeSnapshot(state.FilterExcludeSnapshot.ValueString())
	}
	if !state.FilterKind.IsNull() {
		optionalParams.WithFilterKind(state.FilterKind.ValueString())
	}
	if !state.FilterOwner.IsNull() {
		optionalParams.WithFilterOwner(state.FilterOwner.ValueString())
	}
	if !state.FilterRelationType.IsNull() {
		rel, err := datadogV2.NewRelationTypeFromValue(state.FilterRelationType.ValueString())
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("relation type value is incorrect, please refer to allowed values in specification")))
			return
		}
		optionalParams.WithFilterRelationType(*rel)
	}
	if !state.FilterRef.IsNull() {
		optionalParams.WithFilterRef(state.FilterRef.ValueString())
	}

	offset := int64(0)
	limit := int64(100)

	var entities []datadogV2.EntityData
	for {
		optionalParams.WithPageOffset(offset)
		optionalParams.WithPageLimit(limit)

		ddResp, _, err := d.Api.ListCatalogEntity(d.Auth, *optionalParams)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error querying software catalog"))
			return
		}

		entities = append(entities, ddResp.GetData()...)
		if len(ddResp.GetData()) < int(limit) {
			break
		}
		offset += limit
	}

	d.updateState(ctx, resp, &state, &entities)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *datadogSoftwareCatalogDataSource) updateState(ctx context.Context, resp *datasource.ReadResponse, state *datadogSoftwareCatalogDataSourceModel, entities *[]datadogV2.EntityData) {
	var softwareEntities []*CatalogEntityModel

	for _, entity := range *entities {
		attributes := entity.GetAttributes()
		tags, diags := types.ListValueFrom(ctx, types.StringType, attributes.GetTags())
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		softwareEntity := CatalogEntityModel{
			ID:          types.StringValue(entity.GetId()),
			Name:        types.StringValue(attributes.GetName()),
			DisplayName: types.StringValue(attributes.GetDisplayName()),
			Namespace:   types.StringValue(attributes.GetNamespace()),
			Owner:       types.StringValue(attributes.GetOwner()),
			Kind:        types.StringValue(attributes.GetKind()),
			Tags:        tags,
		}

		softwareEntities = append(softwareEntities, &softwareEntity)
	}

	// Making sure that the ordering is stable
	sort.Slice(softwareEntities, func(i, j int) bool {
		return softwareEntities[i].Name.String() < softwareEntities[j].Name.String()
	})

	idHash := fmt.Sprintf("%x", sha256.Sum256([]byte(
		state.FilterID.ValueString()+state.FilterName.ValueString()+state.FilterExcludeSnapshot.ValueString()+
			state.FilterKind.ValueString()+state.FilterOwner.ValueString()+state.FilterRelationType.ValueString()+
			state.FilterRef.ValueString(),
	)))

	state.ID = types.StringValue(idHash)
	state.Entities = softwareEntities
}
