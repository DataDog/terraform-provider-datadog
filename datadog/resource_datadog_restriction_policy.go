package datadog

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func resourceDatadogRestrictionPolicy() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Restriction Policy resource. This feature lets you assign roles to users based on their SAML attributes.",
		ReadContext:   resourceDatadogRestrictionPolicyRead,
		UpdateContext: resourceDatadogRestrictionPolicyUpdate,
		DeleteContext: resourceDatadogRestrictionPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"bindings": {
				Description: "Bindings of relations to principals.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem: GetBindingSchema(),
			},
		},
	}
}

// GetBindingSchema returns the schema specific to permissions
func GetBindingSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"principals": {
				Type:         schema.TypeList,
				Required:     true,
				Description:  "An array of principals.",
			},
			"relation": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The role/level of access.",
			},
		},
	}
}


func resourceDatadogRestrictionPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TODO replace with restriction policy
	apiInstances := meta.(*ProviderConfiguration).DatadogApiInstances
	auth := meta.(*ProviderConfiguration).Auth

	resp, httpResponse, err := apiInstances.GetRestrictionPoliciesApiV2().GetRestrictionPolicy(auth, d.Id())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId((""))
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting restriction policy")
	}
	restrictionPolicyData := resp.GetData()
	return updateRestrictionPolicyState(d, &restrictionPolicyData)
}

func resourceDatadogRestrictionPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TODO replace with restriction policy
	apiInstances := meta.(*ProviderConfiguration).DatadogApiInstances
	auth := meta.(*ProviderConfiguration).Auth

	req := buildRestrictionPolicyUpdateRequest(d)
	resp, httpResponse, err := apiInstances.GetRestrictionPoliciesApiV2().UpdateRestrictionPolicy(auth, d.Id(), req)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating restriction policy")
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}

	restrictionPolicyData := resp.GetData()
	return updateRestrictionPolicyState(d, &restrictionPolicyData)
}

func resourceDatadogRestrictionPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiInstances := meta.(*ProviderConfiguration).DatadogApiInstances
	auth := meta.(*ProviderConfiguration).Auth

	httpResponse, err := apiInstances.GetRestrictionPoliciesApiV2().DeleteRestrictionPolicy(auth, d.Id())
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting restriction policy")
	}

	return nil
}

func updateRestrictionPolicyState(d *schema.ResourceData, restrictionPolicy *datadogV2.RestrictionPolicy) diag.Diagnostics {
	restrictionPolicyAttributes := restrictionPolicy.GetAttributes()
	restrictionPolicyRelations := restrictionPolicy.GetRelationships()
	restrictionPolicyRoleRelation := restrictionPolicyRelations.GetRole()
	authNRole := restrictionPolicyRoleRelation.GetData()

	if err := d.Set("key", restrictionPolicyAttributes.GetAttributeKey()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("value", restrictionPolicyAttributes.GetAttributeValue()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("role", authNRole.GetId()); err != nil {
		return diag.FromErr(err)
	}
	return nil
}


func buildRestrictionPolicyUpdateRequest(d *schema.ResourceData) datadogV2.RestrictionPolicyUpdateRequest {
	restrictionPolicyUpdateRequest := datadogV2.NewRestrictionPolicyUpdateRequestWithDefaults()
	restrictionPolicy := datadogV2.NewRestrictionPolicyWithDefaults()
	restrictionPolicyAttributes := restrictionPolicy.GetAttributes()

	// Set bindings
	restrictionPolicyAttributes.SetBindings(d.Get("bindings").([]datadogV2.RestrictionPolicyBinding))

	// Set restriction policy update data
	restrictionPolicy.SetAttributes(restrictionPolicyAttributes)
	restrictionPolicy.SetId(d.Id())

	// Set restriction policy update request
	restrictionPolicyUpdateRequest.SetData(*restrictionPolicy)
	return *restrictionPolicyUpdateRequest
}
