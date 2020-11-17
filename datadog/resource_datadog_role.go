package datadog

import (
	"github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogRole() *schema.Resource {
	return &schema.Resource{
		Exists: resourceDatadogRoleExists,
		Create: resourceDatadogRoleCreate,
		Read:   resourceDatadogRoleRead,
		Update: resourceDatadogRoleUpdate,
		Delete: resourceDatadogRoleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogRoleImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the role.",
			},
			"permissions": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of permission IDs to give to this role.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"user_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of users that have this role.",
			},
		},
	}
}

func resourceDatadogRoleExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2
	_, httpresp, err := client.RolesApi.GetRole(auth, d.Id()).Execute()
	if err != nil {
		if httpresp.StatusCode == 404 {
			return false, nil
		}
		return false, translateClientError(err, "error checking if role exists")
	}
	return true, nil
}

func resourceDatadogRoleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2

	roleReq := buildRoleCreateRequest(d)
	resp, _, err := client.RolesApi.CreateRole(auth).Body(roleReq).Execute()
	if err != nil {
		return translateClientError(err, "error creating role")
	}
	roleData := resp.GetData()
	d.SetId(roleData.GetId())

	return resourceDatadogRoleRead(d, meta)
}

func resourceDatadogRoleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2
	resp, _, err := client.RolesApi.GetRole(auth, d.Id()).Execute()
	if err != nil {
		return translateClientError(err, "error getting role")
	}
	roleData := resp.GetData()
	roleAttrs := roleData.GetAttributes()
	d.Set("user_count", roleAttrs.GetUserCount())
	d.Set("name", roleAttrs.GetName())

	roleRelations := roleData.GetRelationships()
	rolePerms := roleRelations.GetPermissions()
	rolePermsData := rolePerms.GetData()
	perms := make([]string, len(rolePermsData))
	for i, perm := range rolePermsData {
		perms[i] = perm.GetId()
	}
	d.Set("permissions", perms)

	return nil
}

func resourceDatadogRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderConfiguration).DatadogClientV2
	auth := meta.(*ProviderConfiguration).AuthV2

	if d.HasChange("name") {
		roleReq := buildRoleUpdateRequest(d)
		_, _, err := client.RolesApi.UpdateRole(auth, d.Id()).Body(roleReq).Execute()
		if err != nil {
			return translateClientError(err, "error updating role")
		}
	}
	if d.HasChange("permissions") {
		oldPermsI, newPermsI := d.GetChange("permissions")
		oldPerms := oldPermsI.(*schema.Set)
		newPerms := newPermsI.(*schema.Set)
		permsToRemove := oldPerms.Difference(newPerms)
		permsToAdd := newPerms.Difference(oldPerms)
		for _, permI := range permsToRemove.List() {
			permRelation := datadog.NewRelationshipToPermissionWithDefaults()
			permRelationData := datadog.NewRelationshipToPermissionDataWithDefaults()
			permRelationData.SetId(permI.(string))
			permRelation.SetData(*permRelationData)
			_, _, err := client.RolesApi.RemovePermissionFromRole(auth, d.Id()).Body(*permRelation).Execute()
			if err != nil {
				return translateClientError(err, "error removing permission from role")
			}
		}
		for _, permI := range permsToAdd.List() {
			permRelation := datadog.NewRelationshipToPermissionWithDefaults()
			permRelationData := datadog.NewRelationshipToPermissionDataWithDefaults()
			permRelationData.SetId(permI.(string))
			permRelation.SetData(*permRelationData)
			_, _, err := client.RolesApi.AddPermissionToRole(auth, d.Id()).Body(*permRelation).Execute()
			if err != nil {
				return translateClientError(err, "error adding permission to role")
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
		return translateClientError(err, "error deleting role")
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
	if permsI, ok := d.GetOk("permissions"); ok {
		perms := permsI.([]string)
		rolePermRelations := datadog.NewRelationshipToPermissionsWithDefaults()
		rolePermRelationsData := make([]datadog.RelationshipToPermissionData, len(perms))
		for i, perm := range perms {
			roleRelationshipToPerm := datadog.NewRelationshipToPermissionDataWithDefaults()
			roleRelationshipToPerm.SetId(perm)
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
