package datadog

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// unrestrictedPermissions is a map of all unrestricted permission IDs to their name
var unrestrictedPermissions map[string]string

// restrictedPermissions is a map of all restricted permission IDs to their name
var restrictedPermissions map[string]string

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
		CustomizeDiff: resourceDatadogRoleCustomizeDiff,
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Name of the role.",
				},
				"default_permissions_opt_out": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "If set to `true`, the role will not have default (restricted) permissions unless they are explicitly set. The `include_restricted` attribute for the datadog_permissions data source must be set to `true` to manage default permissions in Terraform",
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
				"validate": {
					Description: "If set to `false`, skip the validation call done during plan.",
					Type:        schema.TypeBool,
					Optional:    true,
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						// This is never sent to the backend, so it should never generate a diff
						return true
					},
				},
			}
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

func getAllPermissions(ctx context.Context, apiInstances *utils.ApiInstances) (map[string]string, map[string]string, error) {
	// Get a list of all restricted permissions
	if unrestrictedPermissions == nil || restrictedPermissions == nil {
		res, httpResponse, err := apiInstances.GetRolesApiV2().ListPermissions(ctx)
		if err != nil {
			return nil, nil, utils.TranslateClientError(err, httpResponse, "error listing permissions")
		}
		permsList := res.GetData()

		newRestrictedPerms := map[string]string{}
		newUnrestrictedPerms := map[string]string{}

		for _, perm := range permsList {
			if perm.Attributes.GetRestricted() {
				newRestrictedPerms[perm.GetId()] = perm.Attributes.GetName()
			} else {
				newUnrestrictedPerms[perm.GetId()] = perm.Attributes.GetName()
			}
		}
		unrestrictedPermissions = newUnrestrictedPerms
		restrictedPermissions = newRestrictedPerms
	}
	return unrestrictedPermissions, restrictedPermissions, nil
}

func resourceDatadogRoleCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if validate, ok := diff.GetOkExists("validate"); ok && !validate.(bool) {
		// Explicitly skip validation
		return nil
	}

	permissions, ok := diff.GetOkExists("permission")
	if !ok {
		return nil
	}

	defaultPermissionsOptOut, ok := diff.GetOk("default_permissions_opt_out")

	apiInstances := meta.(*ProviderConfiguration).DatadogApiInstances
	auth := meta.(*ProviderConfiguration).Auth

	// Get a list of all valid permissions
	unrestrictedPerms, restrictedPerms, err := getAllPermissions(auth, apiInstances)
	if err != nil {
		return err
	}

	perms := permissions.(*schema.Set)
	for _, permI := range perms.List() {
		perm := permI.(map[string]interface{})
		permID := perm["id"].(string)

		_, isRestrictedPerm := restrictedPerms[permID]
		_, isUnrestrictedPerm := unrestrictedPerms[permID]

		if isRestrictedPerm && !defaultPermissionsOptOut.(bool) {
			return fmt.Errorf(
				"permission with ID %s is a restricted (default) permission and cannot be managed by terraform, set `default_permissions_opt_out` to `true` to manage default permissions, or remove it from your configuration",
				permID,
			)
		}

		if !(isRestrictedPerm || isUnrestrictedPerm) {
			return fmt.Errorf(
				"permission with ID %s does not exist, remove it from your configuration",
				permID,
			)
		}
	}

	return nil
}

func resourceDatadogRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiInstances := meta.(*ProviderConfiguration).DatadogApiInstances
	auth := meta.(*ProviderConfiguration).Auth

	roleReq := buildRoleCreateRequest(d)
	createResp, httpResponse, err := apiInstances.GetRolesApiV2().CreateRole(auth, *roleReq)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating role")
	}
	if err := utils.CheckForUnparsed(createResp); err != nil {
		return diag.FromErr(err)
	}

	var getRoleResponse datadogV2.RoleResponse
	var httpResponseGet *http.Response
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		getRoleResponse, httpResponseGet, err = apiInstances.GetRolesApiV2().GetRole(auth, createResp.Data.GetId())
		if err != nil {
			if httpResponseGet != nil && httpResponseGet.StatusCode == 404 {
				return retry.RetryableError(fmt.Errorf("role not created yet"))
			}

			return retry.NonRetryableError(err)
		}
		if err := utils.CheckForUnparsed(getRoleResponse); err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	roleData := getRoleResponse.GetData()
	d.SetId(roleData.GetId())

	return updateRoleState(auth, d, roleData.Attributes, roleData.Relationships, apiInstances)
}

func updateRoleState(ctx context.Context, d *schema.ResourceData, roleAttrsI interface{}, roleRelations *datadogV2.RoleResponseRelationships, apiInstances *utils.ApiInstances) diag.Diagnostics {
	type namer interface {
		GetName() string
	}
	if roleAttrsI != nil {
		switch roleAttrs := roleAttrsI.(type) {
		case *datadogV2.RoleAttributes:
			if err := d.Set("user_count", roleAttrs.GetUserCount()); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("name", roleAttrs.GetName()); err != nil {
				return diag.FromErr(err)
			}
		case *datadogV2.RoleUpdateAttributes, *datadogV2.RoleCreateAttributes:
			if err := d.Set("name", roleAttrs.(namer).GetName()); err != nil {
				return diag.FromErr(err)
			}
		default:
			return diag.Errorf("unexpected type %s for role attributes", reflect.TypeOf(roleAttrsI).String())
		}
	}

	rolePerms := roleRelations.GetPermissions()
	return updateRolePermissionsState(ctx, d, rolePerms.GetData(), apiInstances)
}

func updateRolePermissionsState(ctx context.Context, d *schema.ResourceData, rolePermsI interface{}, apiInstances *utils.ApiInstances) diag.Diagnostics {

	// Get a list of all valid permissions
	unrestrictedPerms, restrictedPerms, err := getAllPermissions(ctx, apiInstances)
	if err != nil {
		return diag.FromErr(err)
	}

	permsIDToName := make(map[string]string, len(unrestrictedPerms)+len(restrictedPerms))
	for id, name := range unrestrictedPerms {
		permsIDToName[id] = name
	}
	for id, name := range restrictedPerms {
		permsIDToName[id] = name
	}

	var perms []map[string]string
	switch rolePerms := rolePermsI.(type) {
	case []datadogV2.RelationshipToPermissionData:
		for _, perm := range rolePerms {
			perms = appendPerm(perms, perm.GetId(), permsIDToName)
		}
	case []datadogV2.Permission:
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
	apiInstances := meta.(*ProviderConfiguration).DatadogApiInstances
	auth := meta.(*ProviderConfiguration).Auth

	// Get the role
	resp, httpresp, err := apiInstances.GetRolesApiV2().GetRole(auth, d.Id())
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
	return updateRoleState(auth, d, roleData.Attributes, roleData.Relationships, apiInstances)
}

func resourceDatadogRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiInstances := meta.(*ProviderConfiguration).DatadogApiInstances
	auth := meta.(*ProviderConfiguration).Auth

	if d.HasChange("name") || d.HasChange("permission") || d.HasChange("default_permissions_opt_out") {
		roleReq := buildRoleUpdateRequest(d)
		resp, httpResponse, err := apiInstances.GetRolesApiV2().UpdateRole(auth, d.Id(), *roleReq)
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpResponse, "error updating role")
		}
		if err := utils.CheckForUnparsed(resp); err != nil {
			return diag.FromErr(err)
		}
		roleData := resp.GetData()
		if err := updateRoleState(auth, d, roleData.Attributes, roleData.Relationships, apiInstances); err != nil {
			return err
		}
	}

	return nil
}

func resourceDatadogRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiInstances := meta.(*ProviderConfiguration).DatadogApiInstances
	auth := meta.(*ProviderConfiguration).Auth

	httpResponse, err := apiInstances.GetRolesApiV2().DeleteRole(auth, d.Id())
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting role")
	}

	return nil
}

func buildRoleCreateRequest(d *schema.ResourceData) *datadogV2.RoleCreateRequest {
	roleCreateRequest := datadogV2.NewRoleCreateRequestWithDefaults()
	roleCreateData := datadogV2.NewRoleCreateDataWithDefaults()
	roleCreateAttrs := datadogV2.NewRoleCreateAttributesWithDefaults()
	roleCreateRelations := datadogV2.NewRoleRelationshipsWithDefaults()

	// Set attributes
	roleCreateAttrs.SetName(d.Get("name").(string))
	roleCreateAttrs.AdditionalProperties = map[string]any{
		"default_permissions_opt_out": d.Get("default_permissions_opt_out"),
	}
	roleCreateData.SetAttributes(*roleCreateAttrs)

	// Set permission relationships
	if permsI, ok := d.GetOk("permission"); ok {
		perms := permsI.(*schema.Set).List()
		rolePermRelations := datadogV2.NewRelationshipToPermissionsWithDefaults()
		rolePermRelationsData := make([]datadogV2.RelationshipToPermissionData, len(perms))
		for i, permI := range perms {
			perm := permI.(map[string]interface{})
			roleRelationshipToPerm := datadogV2.NewRelationshipToPermissionDataWithDefaults()
			roleRelationshipToPerm.SetId(perm["id"].(string))
			rolePermRelationsData[i] = *roleRelationshipToPerm
		}
		rolePermRelations.SetData(rolePermRelationsData)
		roleCreateRelations.SetPermissions(*rolePermRelations)
	}
	roleCreateData.SetRelationships(*roleCreateRelations)

	roleCreateRequest.SetData(*roleCreateData)
	return roleCreateRequest
}

func buildRoleUpdateRequest(d *schema.ResourceData) *datadogV2.RoleUpdateRequest {
	roleUpdateRequest := datadogV2.NewRoleUpdateRequestWithDefaults()
	roleUpdateData := datadogV2.NewRoleUpdateDataWithDefaults()
	roleUpdateAttributes := datadogV2.NewRoleUpdateAttributesWithDefaults()
	roleUpdateRelations := datadogV2.NewRoleRelationshipsWithDefaults()

	if name, ok := d.GetOk("name"); ok {
		roleUpdateAttributes.SetName(name.(string))
	}

	roleUpdateData.SetId(d.Id())

	roleUpdateAttributes.AdditionalProperties = map[string]any{
		"default_permissions_opt_out": d.Get("default_permissions_opt_out"),
	}
	roleUpdateData.SetAttributes(*roleUpdateAttributes)

	// Set permission relationships
	rolePermRelations := datadogV2.NewRelationshipToPermissionsWithDefaults()
	if permsI, ok := d.GetOk("permission"); ok {
		perms := permsI.(*schema.Set).List()
		rolePermRelationsData := make([]datadogV2.RelationshipToPermissionData, len(perms))
		for i, permI := range perms {
			perm := permI.(map[string]interface{})
			roleRelationshipToPerm := datadogV2.NewRelationshipToPermissionDataWithDefaults()
			roleRelationshipToPerm.SetId(perm["id"].(string))
			rolePermRelationsData[i] = *roleRelationshipToPerm
		}
		rolePermRelations.SetData(rolePermRelationsData)
	} else {
		// Must set permissions to empty slice if there are none so that all
		// unrestricted permissions are removed instead of being left unchanged
		rolePermRelationsData := []datadogV2.RelationshipToPermissionData{}
		rolePermRelations.SetData(rolePermRelationsData)
	}
	roleUpdateRelations.SetPermissions(*rolePermRelations)
	roleUpdateData.SetRelationships(*roleUpdateRelations)

	roleUpdateRequest.SetData(*roleUpdateData)
	return roleUpdateRequest
}
