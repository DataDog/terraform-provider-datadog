package datadog

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// validPermissions is a map of all unrestricted permission IDs to their name
var validPermissions map[string]string

func resourceDatadogRole() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Datadog role resource. This can be used to create and manage Datadog roles.",
		Create:      resourceDatadogRoleCreate,
		Read:        resourceDatadogRoleRead,
		Update:      resourceDatadogRoleUpdate,
		Delete:      resourceDatadogRoleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogRoleImport,
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

func getValidPermissions(client *datadog.APIClient, auth context.Context) (map[string]string, error) {
	// Get a list of all permissions, to ignore restricted perms
	if validPermissions == nil {
		res, _, err := client.RolesApi.ListPermissions(auth).Execute()
		if err != nil {
			return nil, utils.TranslateClientError(err, "error listing permissions")
		}
		permsList := res.GetData()
		permsNameToId := make(map[string]string, len(permsList))
		for _, perm := range permsList {
			if !perm.Attributes.GetRestricted() {
				permsNameToId[perm.GetId()] = perm.Attributes.GetName()
			}
		}
		validPermissions = permsNameToId
	}
	return validPermissions, nil
}

func validatePermissionsUnrestricted(value interface{}, meta interface{}) error {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2

	// Get a list of all valid permissions
	validPerms, err := getValidPermissions(client, auth)
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

func resourceDatadogRoleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2

	roleReq := buildRoleCreateRequest(d)
	resp, _, err := client.RolesApi.CreateRole(auth).Body(roleReq).Execute()
	if err != nil {
		return utils.TranslateClientError(err, "error creating role")
	}
	roleData := resp.GetData()
	d.SetId(roleData.GetId())

	return resourceDatadogRoleRead(d, meta)
}

func resourceDatadogRoleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2

	// Get the role
	resp, httpresp, err := client.RolesApi.GetRole(auth, d.Id()).Execute()
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientError(err, "error getting role")
	}
	roleData := resp.GetData()
	roleAttrs := roleData.GetAttributes()
	if err := d.Set("user_count", roleAttrs.GetUserCount()); err != nil {
		return err
	}
	if err := d.Set("name", roleAttrs.GetName()); err != nil {
		return err
	}

	// Get a list of all valid permissions, to ignore restricted perms
	permsIDToName, err := getValidPermissions(client, auth)
	if err != nil {
		return err
	}

	roleRelations := roleData.GetRelationships()
	rolePerms := roleRelations.GetPermissions()
	rolePermsData := rolePerms.GetData()
	var perms []map[string]string
	for _, perm := range rolePermsData {
		permID := perm.GetId()
		// If perm ID is not restricted, add it to the state
		if permName, ok := permsIDToName[permID]; ok {
			permR := map[string]string{
				"id":   permID,
				"name": permName,
			}
			perms = append(perms, permR)
		}
	}

	if err := d.Set("permission", perms); err != nil {
		return err
	}

	return nil
}

func resourceDatadogRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2

	if d.HasChange("name") {
		roleReq := buildRoleUpdateRequest(d)
		_, _, err := client.RolesApi.UpdateRole(auth, d.Id()).Body(roleReq).Execute()
		if err != nil {
			return utils.TranslateClientError(err, "error updating role")
		}
	}
	if d.HasChange("permission") {
		oldPermsI, newPermsI := d.GetChange("permission")
		oldPerms := oldPermsI.(*schema.Set)
		newPerms := newPermsI.(*schema.Set)
		permsToRemove := oldPerms.Difference(newPerms)
		permsToAdd := newPerms.Difference(oldPerms)
		for _, permI := range permsToRemove.List() {
			perm := permI.(map[string]interface{})
			permRelation := datadog.NewRelationshipToPermissionWithDefaults()
			permRelationData := datadog.NewRelationshipToPermissionDataWithDefaults()
			permRelationData.SetId(perm["id"].(string))
			permRelation.SetData(*permRelationData)
			_, _, err := client.RolesApi.RemovePermissionFromRole(auth, d.Id()).Body(*permRelation).Execute()
			if err != nil {
				return utils.TranslateClientError(err, "error removing permission from role")
			}
		}
		for _, permI := range permsToAdd.List() {
			perm := permI.(map[string]interface{})
			permRelation := datadog.NewRelationshipToPermissionWithDefaults()
			permRelationData := datadog.NewRelationshipToPermissionDataWithDefaults()
			permRelationData.SetId(perm["id"].(string))
			permRelation.SetData(*permRelationData)
			_, _, err := client.RolesApi.AddPermissionToRole(auth, d.Id()).Body(*permRelation).Execute()
			if err != nil {
				return utils.TranslateClientError(err, "error adding permission to role")
			}
		}
	}

	return resourceDatadogRoleRead(d, meta)
}

func resourceDatadogRoleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2

	_, err := client.RolesApi.DeleteRole(auth, d.Id()).Execute()
	if err != nil {
		return utils.TranslateClientError(err, "error deleting role")
	}

	return nil
}

func resourceDatadogRoleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogRoleRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
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
