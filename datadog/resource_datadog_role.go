package datadog

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// validPermissions is a map of all unrestricted permission IDs to their name
var validPermissions map[string]string

func resourceDatadogRole() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog role resource. This can be used to create and manage Datadog roles.",
		CreateContext: resourceDatadogRoleCreate,
		ReadContext:   resourceDatadogRoleRead,
		UpdateContext: resourceDatadogRoleUpdate,
		DeleteContext: resourceDatadogRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customdiff.ValidateValue("permission", validatePermissionsUnrestricted),
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the role.",
			},
			"permission": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Set of objects containing the permission ID and the name of the permissions granted to this role.",
				Elem:        GetRolePermissionSchema(),
			},
			"user_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of users that have this role.",
			},
		},
	}
}

// GetRolePermissionSchema returns the schema specific to permissions
func GetRolePermissionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "ID of the permission to assign.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the permission.",
			},
		},
	}
}

func getValidPermissions(ctx context.Context, client *datadog.APIClient) (map[string]string, error) {
	// Get a list of all permissions, to ignore restricted perms
	if validPermissions == nil {
		res, httpResponse, err := client.RolesApi.ListPermissions(ctx)
		if err != nil {
			return nil, utils.TranslateClientError(err, httpResponse, "error listing permissions")
		}
		permsList := res.GetData()
		permsNameToID := make(map[string]string, len(permsList))
		for _, perm := range permsList {
			if !perm.Attributes.GetRestricted() {
				permsNameToID[perm.GetId()] = perm.Attributes.GetName()
			}
		}
		validPermissions = permsNameToID
	}
	return validPermissions, nil
}

func validatePermissionsUnrestricted(ctx context.Context, value interface{}, meta interface{}) error {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2

	// Get a list of all valid permissions
	validPerms, err := getValidPermissions(auth, client)
	if err != nil {
		return err
	}

	perms := value.(*schema.Set)
	for _, permI := range perms.List() {
		perm := permI.(map[string]interface{})
		permID := perm["id"].(string)
		if _, ok := validPerms[permID]; !ok {
			return fmt.Errorf(
				"permission with ID %s is restricted and cannot be managed by terraform or does not exist, remove it from your configuration",
				permID,
			)
		}
	}

	return nil
}

func resourceDatadogRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2

	roleReq := buildRoleCreateRequest(d)
	resp, httpResponse, err := client.RolesApi.CreateRole(auth, roleReq)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating role")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	roleData := resp.GetData()
	d.SetId(roleData.GetId())

	return updateRoleState(auth, d, roleData.Attributes, roleData.Relationships, client)
}

func updateRoleState(ctx context.Context, d *schema.ResourceData, roleAttrsI interface{}, roleRelations *datadog.RoleResponseRelationships, client *datadog.APIClient) diag.Diagnostics {
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
}

func updateRolePermissionsState(ctx context.Context, d *schema.ResourceData, rolePermsI interface{}, client *datadog.APIClient) diag.Diagnostics {

	// Get a list of all valid permissions, to ignore restricted perms
	permsIDToName, err := getValidPermissions(ctx, client)
	if err != nil {
		return diag.FromErr(err)
	}

	var perms []map[string]string
	switch rolePerms := rolePermsI.(type) {
	case []datadog.RelationshipToPermissionData:
		for _, perm := range rolePerms {
			perms = appendPerm(perms, perm.GetId(), permsIDToName)
		}
	case []datadog.Permission:
		for _, perm := range rolePerms {
			perms = appendPerm(perms, perm.GetId(), permsIDToName)
		}
	default:
		return diag.Errorf("unexpected type %s for permissions list", reflect.TypeOf(rolePermsI).String())
	}

	if err := d.Set("permission", perms); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func appendPerm(perms []map[string]string, permID string, permsIDToName map[string]string) []map[string]string {
	// If perm ID is not restricted, add it to the state
	if permName, ok := permsIDToName[permID]; ok {
		permR := map[string]string{
			"id":   permID,
			"name": permName,
		}
		perms = append(perms, permR)
	}
	return perms
}

func resourceDatadogRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2

	// Get the role
	resp, httpresp, err := client.RolesApi.GetRole(auth, d.Id())
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting role")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	roleData := resp.GetData()
	return updateRoleState(auth, d, roleData.Attributes, roleData.Relationships, client)
}

func resourceDatadogRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2

	if d.HasChange("name") {
		roleReq := buildRoleUpdateRequest(d)
		resp, httpResponse, err := client.RolesApi.UpdateRole(auth, d.Id(), roleReq)
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpResponse, "error updating role")
		}
		if err := utils.CheckForUnparsed(resp); err != nil {
			return diag.FromErr(err)
		}
		roleData := resp.GetData()
		if err := updateRoleState(auth, d, roleData.Attributes, roleData.Relationships, client); err != nil {
			return err
		}
	}
	if d.HasChange("permission") {
		oldPermsI, newPermsI := d.GetChange("permission")
		oldPerms := oldPermsI.(*schema.Set)
		newPerms := newPermsI.(*schema.Set)
		permsToRemove := oldPerms.Difference(newPerms)
		permsToAdd := newPerms.Difference(oldPerms)
		var (
			permsResponse datadog.PermissionsResponse
			err           error
			httpResponse  *http.Response
		)
		for _, permI := range permsToRemove.List() {
			perm := permI.(map[string]interface{})
			permRelation := datadog.NewRelationshipToPermissionWithDefaults()
			permRelationData := datadog.NewRelationshipToPermissionDataWithDefaults()
			permRelationData.SetId(perm["id"].(string))
			permRelation.SetData(*permRelationData)
			permsResponse, httpResponse, err = client.RolesApi.RemovePermissionFromRole(auth, d.Id(), *permRelation)
			if err != nil {
				return utils.TranslateClientErrorDiag(err, httpResponse, "error removing permission from role")
			}
			if err := utils.CheckForUnparsed(permsResponse); err != nil {
				return diag.FromErr(err)
			}

		}
		for _, permI := range permsToAdd.List() {
			perm := permI.(map[string]interface{})
			permRelation := datadog.NewRelationshipToPermissionWithDefaults()
			permRelationData := datadog.NewRelationshipToPermissionDataWithDefaults()
			permRelationData.SetId(perm["id"].(string))
			permRelation.SetData(*permRelationData)
			permsResponse, httpResponse, err = client.RolesApi.AddPermissionToRole(auth, d.Id(), *permRelation)
			if err != nil {
				return utils.TranslateClientErrorDiag(err, httpResponse, "error adding permission to role")
			}
			if err := utils.CheckForUnparsed(permsResponse); err != nil {
				return diag.FromErr(err)
			}
		}
		// Only need to update once all the permissions have been added/revoked, with the last call response
		if err := updateRolePermissionsState(auth, d, permsResponse.GetData(), client); err != nil {
			return err
		}
	}

	return nil
}

func resourceDatadogRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2

	httpResponse, err := client.RolesApi.DeleteRole(auth, d.Id())
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting role")
	}

	return nil
}

func buildRoleCreateRequest(d *schema.ResourceData) datadog.RoleCreateRequest {
	roleCreateRequest := datadog.NewRoleCreateRequestWithDefaults()
	roleCreateData := datadog.NewRoleCreateDataWithDefaults()
	roleCreateAttrs := datadog.NewRoleCreateAttributesWithDefaults()
	roleCreateRelations := datadog.NewRoleRelationshipsWithDefaults()

	// Set attributes
	roleCreateAttrs.SetName(d.Get("name").(string))
	roleCreateData.SetAttributes(*roleCreateAttrs)

	// Set permission relationships
	if permsI, ok := d.GetOk("permission"); ok {
		perms := permsI.(*schema.Set).List()
		rolePermRelations := datadog.NewRelationshipToPermissionsWithDefaults()
		rolePermRelationsData := make([]datadog.RelationshipToPermissionData, len(perms))
		for i, permI := range perms {
			perm := permI.(map[string]interface{})
			roleRelationshipToPerm := datadog.NewRelationshipToPermissionDataWithDefaults()
			roleRelationshipToPerm.SetId(perm["id"].(string))
			rolePermRelationsData[i] = *roleRelationshipToPerm
		}
		rolePermRelations.SetData(rolePermRelationsData)
		roleCreateRelations.SetPermissions(*rolePermRelations)
	}
	roleCreateData.SetRelationships(*roleCreateRelations)

	roleCreateRequest.SetData(*roleCreateData)
	return *roleCreateRequest
}

func buildRoleUpdateRequest(d *schema.ResourceData) datadog.RoleUpdateRequest {
	roleUpdateRequest := datadog.NewRoleUpdateRequestWithDefaults()
	roleUpdateData := datadog.NewRoleUpdateDataWithDefaults()
	roleUpdateAttributes := datadog.NewRoleUpdateAttributesWithDefaults()

	roleUpdateAttributes.SetName(d.Get("name").(string))

	roleUpdateData.SetId(d.Id())
	roleUpdateData.SetAttributes(*roleUpdateAttributes)

	roleUpdateRequest.SetData(*roleUpdateData)
	return *roleUpdateRequest
}
