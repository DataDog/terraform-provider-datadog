package datadog

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogRoles() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about multiple roles for use in other resources.",
		ReadContext: dataSourceDatadogRolesRead,
		Schema: map[string]*schema.Schema{
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
		},
	}
}

func dataSourceDatadogRolesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	var filterPtr *string

	reqParams := datadog.NewListRolesOptionalParameters()
	if v, ok := d.GetOk("filter"); ok {
		filter := v.(string)
		filterPtr = &filter
		reqParams.Filter = filterPtr
	}

	pageSize := int64(100)
	pageNumber := int64(0)
	remaining := int64(1)

	var roles []datadog.Role
	for remaining > int64(0) {
		rolesResp, httpresp, err := datadogClientV2.RolesApi.ListRoles(authV2, *reqParams.
			WithPageSize(pageSize).
			WithPageNumber(pageNumber))
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpresp, "error querying roles")
		}
		if err := utils.CheckForUnparsed(rolesResp); err != nil {
			return diag.FromErr(err)
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

	tfRoles := make([]map[string]interface{}, 0, len(roles))
	for _, role := range roles {
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

	return nil
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
