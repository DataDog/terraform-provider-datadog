package datadog

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

func resourceDatadogAuthnMapping() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog AuthN Mappings resource. This feature lets you automatically assign roles to users based on their SAML attributes.",
		CreateContext: resourceDatadogAuthnMappingCreate,
		ReadContext:   resourceDatadogAuthnMappingRead,
		UpdateContext: resourceDatadogAuthnMappingUpdate,
		DeleteContext: resourceDatadogAuthnMappingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
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
					Description:  "The ID of a role to attach to all users with the corresponding key and value. Cannot be used in conjunction with `team`.",
					Type:         schema.TypeString,
					Optional:     true,
					ExactlyOneOf: []string{"role", "team"},
				},
				"team": {
					Description:  "The ID of a team to add all users with the corresponding key and value to. Cannot be used in conjunction with `role`.",
					Type:         schema.TypeString,
					Optional:     true,
					ExactlyOneOf: []string{"role", "team"},
				},
			}
		},
	}
}

func resourceDatadogAuthnMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiInstances := meta.(*ProviderConfiguration).DatadogApiInstances
	auth := meta.(*ProviderConfiguration).Auth
	authNMapReq := buildAuthNMappingCreateRequest(d)

	createResp, httpResponse, err := apiInstances.GetAuthNMappingsApiV2().CreateAuthNMapping(auth, *authNMapReq)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating authn mapping")
	}
	if err := utils.CheckForUnparsed(createResp); err != nil {
		return diag.FromErr(err)
	}

	var getAuthNMappingResponse datadogV2.AuthNMappingResponse
	var httpResponseGet *http.Response

	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		getAuthNMappingResponse, httpResponseGet, err = apiInstances.GetAuthNMappingsApiV2().GetAuthNMapping(auth, createResp.Data.GetId())
		if err != nil {
			if httpResponseGet != nil && httpResponseGet.StatusCode == 404 {
				return retry.RetryableError(fmt.Errorf("SAML role mapping not created yet"))
			}

			return retry.NonRetryableError(err)
		}
		if err := utils.CheckForUnparsed(getAuthNMappingResponse); err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	authNMappingData := getAuthNMappingResponse.GetData()
	d.SetId(authNMappingData.GetId())
	return updateAuthNMappingState(d, &authNMappingData)
}

func resourceDatadogAuthnMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiInstances := meta.(*ProviderConfiguration).DatadogApiInstances
	auth := meta.(*ProviderConfiguration).Auth

	resp, httpResponse, err := apiInstances.GetAuthNMappingsApiV2().GetAuthNMapping(auth, d.Id())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId((""))
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting authn mapping")
	}
	authNMappingData := resp.GetData()
	return updateAuthNMappingState(d, &authNMappingData)
}

func resourceDatadogAuthnMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiInstances := meta.(*ProviderConfiguration).DatadogApiInstances
	auth := meta.(*ProviderConfiguration).Auth

	req := buildAuthNMappingUpdateRequest(d)
	resp, httpResponse, err := apiInstances.GetAuthNMappingsApiV2().UpdateAuthNMapping(auth, d.Id(), *req)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating role mapping")
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}

	authNMappingData := resp.GetData()
	return updateAuthNMappingState(d, &authNMappingData)
}

func resourceDatadogAuthnMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiInstances := meta.(*ProviderConfiguration).DatadogApiInstances
	auth := meta.(*ProviderConfiguration).Auth

	httpResponse, err := apiInstances.GetAuthNMappingsApiV2().DeleteAuthNMapping(auth, d.Id())
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting authn mapping")
	}

	return nil
}

func updateAuthNMappingState(d *schema.ResourceData, authNMapping *datadogV2.AuthNMapping) diag.Diagnostics {
	authNMappingAttributes := authNMapping.GetAttributes()
	authNMappingRelations := authNMapping.GetRelationships()
	authNMappingRoleRelation := authNMappingRelations.GetRole()
	authNMappingTeamRelation := authNMappingRelations.GetTeam()
	authNRole := authNMappingRoleRelation.GetData()
	authNTeam := authNMappingTeamRelation.GetData()

	if err := d.Set("key", authNMappingAttributes.GetAttributeKey()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("value", authNMappingAttributes.GetAttributeValue()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("role", authNRole.GetId()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("team", authNTeam.GetId()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func buildAuthNMappingCreateRequest(d *schema.ResourceData) *datadogV2.AuthNMappingCreateRequest {
	authNMappingCreateRequest := datadogV2.NewAuthNMappingCreateRequestWithDefaults()
	authNMappingCreateData := datadogV2.NewAuthNMappingCreateDataWithDefaults()
	authNMappingCreateAttrs := datadogV2.NewAuthNMappingCreateAttributesWithDefaults()

	// Set AuthN mapping Attributes
	authNMappingCreateAttrs.SetAttributeKey(d.Get("key").(string))
	authNMappingCreateAttrs.SetAttributeValue(d.Get("value").(string))

	// Set AuthN mapping Relationships
	relationships := datadogV2.AuthNMappingCreateRelationships{}
	roleRelation := buildRoleRelations(d)
	teamRelation := buildTeamRelations(d)
	if roleRelation != nil {
		relationshipToRole := datadogV2.NewAuthNMappingRelationshipToRoleWithDefaults()
		relationshipToRole.SetRole(*roleRelation)
		relationships.AuthNMappingRelationshipToRole = relationshipToRole
	}
	if teamRelation != nil {
		relationshipToTeam := datadogV2.NewAuthNMappingRelationshipToTeamWithDefaults()
		relationshipToTeam.SetTeam(*teamRelation)
		relationships.AuthNMappingRelationshipToTeam = relationshipToTeam
	}

	// Set AuthN mapping create data
	authNMappingCreateData.SetAttributes(*authNMappingCreateAttrs)
	authNMappingCreateData.SetRelationships(relationships)

	// Set AuthN mapping create request
	authNMappingCreateRequest.SetData(*authNMappingCreateData)
	return authNMappingCreateRequest
}

func buildAuthNMappingUpdateRequest(d *schema.ResourceData) *datadogV2.AuthNMappingUpdateRequest {
	authNMappingUpdateRequest := datadogV2.NewAuthNMappingUpdateRequestWithDefaults()
	authNMappingUpdateData := datadogV2.NewAuthNMappingUpdateDataWithDefaults()
	authNMappingUpdateAttrs := datadogV2.NewAuthNMappingUpdateAttributesWithDefaults()

	// Set AuthN mapping Attributes
	authNMappingUpdateAttrs.SetAttributeKey(d.Get("key").(string))
	authNMappingUpdateAttrs.SetAttributeValue(d.Get("value").(string))

	// Set AuthN mapping Relationships
	relationships := datadogV2.AuthNMappingUpdateRelationships{}
	roleRelation := buildRoleRelations(d)
	teamRelation := buildTeamRelations(d)
	if roleRelation != nil {
		relationshipToRole := datadogV2.NewAuthNMappingRelationshipToRoleWithDefaults()
		relationshipToRole.SetRole(*roleRelation)
		relationships.AuthNMappingRelationshipToRole = relationshipToRole
	}
	if teamRelation != nil {
		relationshipToTeam := datadogV2.NewAuthNMappingRelationshipToTeamWithDefaults()
		relationshipToTeam.SetTeam(*teamRelation)
		relationships.AuthNMappingRelationshipToTeam = relationshipToTeam
	}

	// Set AuthN mapping update data
	authNMappingUpdateData.SetAttributes(*authNMappingUpdateAttrs)
	authNMappingUpdateData.SetRelationships(relationships)
	authNMappingUpdateData.SetId(d.Id())

	// Set AuthN mapping update request
	authNMappingUpdateRequest.SetData(*authNMappingUpdateData)
	return authNMappingUpdateRequest
}

func buildRoleRelations(d *schema.ResourceData) *datadogV2.RelationshipToRole {
	role := d.Get("role")
	if role == nil || role == "" {
		return nil
	}

	roleRelations := datadogV2.NewRelationshipToRoleWithDefaults()
	roleRelationsData := datadogV2.NewRelationshipToRoleDataWithDefaults()

	roleRelationsData.SetId(role.(string))
	roleRelations.SetData(*roleRelationsData)
	return roleRelations
}

func buildTeamRelations(d *schema.ResourceData) *datadogV2.RelationshipToTeam {
	team := d.Get("team")
	if team == nil || team == "" {
		return nil
	}

	teamRelations := datadogV2.NewRelationshipToTeamWithDefaults()
	teamRelationsData := datadogV2.NewRelationshipToTeamDataWithDefaults()

	teamRelationsData.SetId(team.(string))
	teamRelations.SetData(*teamRelationsData)
	return teamRelations
}
