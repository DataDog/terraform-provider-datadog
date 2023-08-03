package datadog

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogRoles() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about multiple roles for use in other resources.",
		ReadContext: dataSourceDatadogRolesRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"filter": {
					Description: "Filter all roles by the given string.",
					Type:        schema.TypeString,
					Optional:    true,
				},

				// Computed values
				"roles": {
					Description: "List of Roles",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Description: "ID of the Datadog role",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"name": {
								Description: "Name of the Datadog role",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"user_count": {
								Description: "Number of users that have this role.",
								Type:        schema.TypeInt,
								Computed:    true,
							},
						},
					},
				},
			}
		},
	}
}

func dataSourceDatadogRolesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	var filterPtr *string

	reqParams := datadogV2.NewListRolesOptionalParameters()
	if v, ok := d.GetOk("filter"); ok {
		filter := v.(string)
		filterPtr = &filter
		reqParams.Filter = filterPtr
	}

	pageSize := int64(100)
	pageNumber := int64(0)
	remaining := int64(1)

	var roles []datadogV2.Role
	for remaining > int64(0) {
		rolesResp, httpresp, err := apiInstances.GetRolesApiV2().ListRoles(auth, *reqParams.
			WithPageSize(pageSize).
			WithPageNumber(pageNumber))
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpresp, "error querying roles")
		}

		roles = append(roles, rolesResp.GetData()...)

		if reqParams.Filter != nil {
			remaining = rolesResp.Meta.Page.GetTotalFilteredCount() - pageSize*(pageNumber+1)
		} else {
			remaining = rolesResp.Meta.Page.GetTotalCount() - pageSize*(pageNumber+1)
		}
		pageNumber++
	}

	if len(roles) == 0 {
		return diag.Errorf("your query returned no result, please try a less specific search criteria")
	}

	diags := diag.Diagnostics{}
	tfRoles := make([]map[string]interface{}, 0, len(roles))
	for _, role := range roles {
		if err := utils.CheckForUnparsed(role); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("skipping role with id: %s", role.GetId()),
				Detail:   fmt.Sprintf("role contains unparsed object: %v", err),
			})
			continue
		}

		attributes := role.GetAttributes()
		tfRoles = append(tfRoles, map[string]interface{}{
			"id":         role.GetId(),
			"name":       attributes.GetName(),
			"user_count": attributes.GetUserCount(),
		})
	}
	if err := d.Set("roles", tfRoles); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(computeRolesDataSourceID(filterPtr))

	return diags
}

func computeRolesDataSourceID(filter *string) string {
	// Key for hashing
	var b strings.Builder
	if filter != nil {
		b.WriteString(*filter)
	}

	keyStr := b.String()
	h := sha256.New()
	log.Println("HASHKEY", keyStr)
	h.Write([]byte(keyStr))

	return fmt.Sprintf("%x", h.Sum(nil))
}
