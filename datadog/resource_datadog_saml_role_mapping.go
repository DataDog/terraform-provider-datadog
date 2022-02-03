package datadog

import (
	"context"
	"fmt"
	"net/http"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func resourceDatadogRoleMapping() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog SAML Role Mappings resource. This can be used to create and manage Datadog SAML Role Mappings.",
		CreateContext: resourceDatadogSamlRoleMappingCreate,
		ReadContext:   resourceDatadogSamlRoleMappingRead,
		UpdateContext: resourceDatadogSamlRoleMappingUpdate,
		DeleteContext: resourceDatadogSamlRoleMappingDelete,

		Schema: map[string]*schema.Schema{
			"key": {
				Description: "Identity provider key.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"value": {
				Description: "Identity provider value.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"role": {
				Description: "The role to assign for key:value mapping.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceDatadogSamlRoleMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2
	samlRoleMapReq := buildSamlRoleMappingCreateRequest(d)
	createResp, httpResponse, err := client.AuthNMappingsApi.CreateAuthNMapping(auth, samlRoleMapReq)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating role")
	}
	if err := utils.CheckForUnparsed(createResp); err != nil {
		return diag.FromErr(err)
	}

	var getSamlMappingResponse datadog.AuthNMappingResponse
	var httpResponseGet *http.Response

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		getSamlMappingResponse, httpResponseGet, err = client.AuthNMappingsApi.GetAuthNMapping(auth, createResp.Data.GetId())
		if err != nil {
			if httpResponseGet != nil && httpResponseGet.StatusCode == 404 {
				return resource.RetryableError(fmt.Errorf("SAML role mapping not created yet"))
			}

			return resource.NonRetryableError(err)
		}
		if err := utils.CheckForUnparsed(getSamlMappingResponse); err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	samlRoleMappingData := getSamlMappingResponse.GetData()
	d.SetId(samlRoleMappingData.GetId())
	return nil
}

func resourceDatadogSamlRoleMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceDatadogSamlRoleMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceDatadogSamlRoleMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func buildSamlRoleMappingCreateRequest(d *schema.ResourceData) datadog.AuthNMappingCreateRequest{
	samlRoleMappingCreateRequest := datadog.NewAuthNMappingCreateRequestWithDefaults()
	samlRoleMappingCreateData := datadog.NewAuthNMappingCreateDataWithDefaults()
	samlRoleMappingCreateAttrs := datadog.NewAuthNMappingCreateAttributesWithDefaults()
	samlRoleMappingRelations := datadog.NewAuthNMappingCreateRelationshipsWithDefaults()

	// Set SAML role mapping Attributes
	samlRoleMappingCreateAttrs.SetAttributeKey(d.Get("key").(string))
	samlRoleMappingCreateAttrs.SetAttributeValue(d.Get("value").(string))

	// Set SAML role mapping Relationships
	roleRelations := datadog.NewRelationshipToRoleWithDefaults()
	roleRelationsData := datadog.NewRelationshipToRoleDataWithDefaults()
	roleRelationsData.SetId(d.Get("role").(string))
	roleRelations.SetData(*roleRelationsData)
	samlRoleMappingRelations.SetRole(*roleRelations)
	
	// Set SAML role mapping create data
	samlRoleMappingCreateData.SetAttributes(*samlRoleMappingCreateAttrs)
	samlRoleMappingCreateData.SetRelationships(*samlRoleMappingRelations)

	// Set SAML role mapping create request
	samlRoleMappingCreateRequest.SetData(*samlRoleMappingCreateData)
	return *samlRoleMappingCreateRequest
}

/*func updateSamlRoleMapState(ctx context.Context, d *schema.ResourceData, roleAttrsI interface{}, roleRelations *datadog.RoleResponseRelationships, client *datadog.APIClient) diag.Diagnostics {
	type namer interface {
		GetName() string
	}
	if roleAttrsI != nil {
		switch roleAttrs := roleAttrsI.(type) {
		case *datadog.RoleAttributes:
			if err := d.Set("user_count", roleAttrs.GetUserCount()); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("name", roleAttrs.GetName()); err != nil {
				return diag.FromErr(err)
			}
		case *datadog.RoleUpdateAttributes, *datadog.RoleCreateAttributes:
			if err := d.Set("name", roleAttrs.(namer).GetName()); err != nil {
				return diag.FromErr(err)
			}
		default:
			return diag.Errorf("unexpected type %s for role attributes", reflect.TypeOf(roleAttrsI).String())
		}
	}

	rolePerms := roleRelations.GetPermissions()
	return updateRolePermissionsState(ctx, d, rolePerms.GetData(), client)
}*/
