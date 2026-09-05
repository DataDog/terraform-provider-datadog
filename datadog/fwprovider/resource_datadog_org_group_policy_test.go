package fwprovider

import (
	"context"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func orgGroupPolicyValidateConfigSchema() schema.Schema {
	r := &OrgGroupPolicyResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	return resp.Schema
}

// orgGroupPolicyObjectType derives the tftypes object from the resource schema
// itself, so the test fixtures can't drift from the real attribute set.
func orgGroupPolicyObjectType() tftypes.Object {
	s := orgGroupPolicyValidateConfigSchema()
	return s.Type().TerraformType(context.Background()).(tftypes.Object)
}

// assertDiagnosticSummary fails unless diags holds an error with the given summary, so
// tests pin down which guard fired rather than just that something failed.
func assertDiagnosticSummary(t *testing.T, diags diag.Diagnostics, wantSummary string) {
	t.Helper()
	for _, d := range diags {
		if d.Severity() == diag.SeverityError && d.Summary() == wantSummary {
			return
		}
	}
	t.Fatalf("expected an error diagnostic with summary %q, got: %v", wantSummary, diags)
}

func orgGroupPolicyConfigValue(policyType, enforcementTier string) tftypes.Value {
	return tftypes.NewValue(orgGroupPolicyObjectType(), map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, nil),
		"org_group_id":     tftypes.NewValue(tftypes.String, "a1b2c3d4-e5f6-7890-abcd-ef0123456789"),
		"policy_name":      tftypes.NewValue(tftypes.String, "finance_read_only"),
		"content":          tftypes.NewValue(tftypes.String, `{"permissions":["1a2b3c4d-5e6f-7890-abcd-ef0123456789"]}`),
		"enforcement_tier": tftypes.NewValue(tftypes.String, enforcementTier),
		"policy_type":      tftypes.NewValue(tftypes.String, policyType),
	})
}

func orgGroupPolicyStateValue(policyType, enforcementTier string) tftypes.Value {
	return orgGroupPolicyStateValueOrgGroup(tftypes.NewValue(tftypes.String, orgGroupPolicyTestGroupID), policyType, enforcementTier)
}

const orgGroupPolicyTestGroupID = "a1b2c3d4-e5f6-7890-abcd-ef0123456789"

// orgGroupPolicyStateValueOrgGroup is orgGroupPolicyStateValue with an explicit
// org_group_id, so tests can cover a changed or still-unknown group reference.
func orgGroupPolicyStateValueOrgGroup(orgGroupID tftypes.Value, policyType, enforcementTier string) tftypes.Value {
	return tftypes.NewValue(orgGroupPolicyObjectType(), map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "b2c3d4e5-f6a7-8901-bcde-f01234567890"),
		"org_group_id":     orgGroupID,
		"policy_name":      tftypes.NewValue(tftypes.String, "finance_read_only"),
		"content":          tftypes.NewValue(tftypes.String, `{"permissions":["1a2b3c4d-5e6f-7890-abcd-ef0123456789"]}`),
		"enforcement_tier": tftypes.NewValue(tftypes.String, enforcementTier),
		"policy_type":      tftypes.NewValue(tftypes.String, policyType),
	})
}

func TestOrgGroupPolicyDelete_RoleRejectsWhenNotDisabled(t *testing.T) {
	s := orgGroupPolicyValidateConfigSchema()
	r := &OrgGroupPolicyResource{}

	req := resource.DeleteRequest{
		State: tfsdk.State{
			Raw:    orgGroupPolicyStateValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE), "GROUP_MANAGED"),
			Schema: s,
		},
	}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assertDiagnosticSummary(t, resp.Diagnostics, "role policies cannot be deleted")
}

func TestOrgGroupPolicyDelete_RoleAllowsWhenAlreadyDisabled(t *testing.T) {
	s := orgGroupPolicyValidateConfigSchema()
	r := &OrgGroupPolicyResource{}

	req := resource.DeleteRequest{
		State: tfsdk.State{
			Raw:    orgGroupPolicyStateValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE), "DELEGATE"),
			Schema: s,
		},
	}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("did not expect an error deleting an already-disabled role policy: %v", resp.Diagnostics)
	}
}

func TestOrgGroupPolicyValidateConfig_RoleRejectsOverrideAllowed(t *testing.T) {
	s := orgGroupPolicyValidateConfigSchema()
	r := &OrgGroupPolicyResource{}

	req := resource.ValidateConfigRequest{
		Config: tfsdk.Config{
			Raw:    orgGroupPolicyConfigValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE), "OVERRIDE_ALLOWED"),
			Schema: s,
		},
	}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(context.Background(), req, resp)

	assertDiagnosticSummary(t, resp.Diagnostics, `Invalid enforcement_tier for policy_type "role"`)
}

func TestOrgGroupPolicyValidateConfig_RoleAllowsGroupManagedAndDelegate(t *testing.T) {
	s := orgGroupPolicyValidateConfigSchema()
	r := &OrgGroupPolicyResource{}

	for _, tier := range []string{"GROUP_MANAGED", "DELEGATE"} {
		req := resource.ValidateConfigRequest{
			Config: tfsdk.Config{
				Raw:    orgGroupPolicyConfigValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE), tier),
				Schema: s,
			},
		}
		resp := &resource.ValidateConfigResponse{}
		r.ValidateConfig(context.Background(), req, resp)

		if resp.Diagnostics.HasError() {
			t.Errorf("did not expect an error for policy_type=role with enforcement_tier=%s: %v", tier, resp.Diagnostics)
		}
	}
}

func TestOrgGroupPolicyValidateConfig_OrgConfigAllowsOverrideAllowed(t *testing.T) {
	s := orgGroupPolicyValidateConfigSchema()
	r := &OrgGroupPolicyResource{}

	req := resource.ValidateConfigRequest{
		Config: tfsdk.Config{
			Raw:    orgGroupPolicyConfigValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ORG_CONFIG), "OVERRIDE_ALLOWED"),
			Schema: s,
		},
	}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("did not expect an error for policy_type=org_config with enforcement_tier=OVERRIDE_ALLOWED: %v", resp.Diagnostics)
	}
}

func orgGroupPolicyConfigValueNullTier(policyType string) tftypes.Value {
	return tftypes.NewValue(orgGroupPolicyObjectType(), map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, nil),
		"org_group_id":     tftypes.NewValue(tftypes.String, "a1b2c3d4-e5f6-7890-abcd-ef0123456789"),
		"policy_name":      tftypes.NewValue(tftypes.String, "finance_read_only"),
		"content":          tftypes.NewValue(tftypes.String, `{"permissions":["1a2b3c4d-5e6f-7890-abcd-ef0123456789"]}`),
		"enforcement_tier": tftypes.NewValue(tftypes.String, nil),
		"policy_type":      tftypes.NewValue(tftypes.String, policyType),
	})
}

func TestOrgGroupPolicyUpdate_RoleRejectsReEnableFromDelegate(t *testing.T) {
	s := orgGroupPolicyValidateConfigSchema()
	r := &OrgGroupPolicyResource{}

	req := resource.UpdateRequest{
		State: tfsdk.State{
			Raw:    orgGroupPolicyStateValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE), "DELEGATE"),
			Schema: s,
		},
		Plan: tfsdk.Plan{
			Raw:    orgGroupPolicyStateValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE), "GROUP_MANAGED"),
			Schema: s,
		},
	}
	resp := &resource.UpdateResponse{}
	r.Update(context.Background(), req, resp)

	assertDiagnosticSummary(t, resp.Diagnostics, "Cannot re-enable a disabled role policy")
}

func TestOrgGroupPolicyUpdate_RoleInvalidTierReportedBeforeReEnable(t *testing.T) {
	s := orgGroupPolicyValidateConfigSchema()
	r := &OrgGroupPolicyResource{}

	// Both guards match here; the tier one must win so the message names the bad value.
	req := resource.UpdateRequest{
		State: tfsdk.State{
			Raw:    orgGroupPolicyStateValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE), "DELEGATE"),
			Schema: s,
		},
		Plan: tfsdk.Plan{
			Raw:    orgGroupPolicyStateValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE), "OVERRIDE_ALLOWED"),
			Schema: s,
		},
	}
	resp := &resource.UpdateResponse{}
	r.Update(context.Background(), req, resp)

	assertDiagnosticSummary(t, resp.Diagnostics, "Invalid enforcement_tier for policy_type \"role\"")
}

func TestOrgGroupPolicyNameRequiresReplace_RoleDoesNotForceReplace(t *testing.T) {
	s := orgGroupPolicyValidateConfigSchema()

	req := planmodifier.StringRequest{
		State: tfsdk.State{
			Raw:    orgGroupPolicyStateValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE), "GROUP_MANAGED"),
			Schema: s,
		},
	}
	resp := &stringplanmodifier.RequiresReplaceIfFuncResponse{}
	orgGroupPolicyNameRequiresReplaceIf(context.Background(), req, resp)

	if resp.RequiresReplace {
		t.Fatal("role policy rename should NOT force replace")
	}
}

func TestOrgGroupPolicyNameRequiresReplace_OrgConfigForcesReplace(t *testing.T) {
	s := orgGroupPolicyValidateConfigSchema()

	req := planmodifier.StringRequest{
		State: tfsdk.State{
			Raw:    orgGroupPolicyStateValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ORG_CONFIG), "OVERRIDE_ALLOWED"),
			Schema: s,
		},
	}
	resp := &stringplanmodifier.RequiresReplaceIfFuncResponse{}
	orgGroupPolicyNameRequiresReplaceIf(context.Background(), req, resp)

	if !resp.RequiresReplace {
		t.Fatal("org_config policy rename should force replace")
	}
}

func TestOrgGroupPolicyValidateConfig_NullEnforcementTierNoDiagnostic(t *testing.T) {
	s := orgGroupPolicyValidateConfigSchema()
	r := &OrgGroupPolicyResource{}

	req := resource.ValidateConfigRequest{
		Config: tfsdk.Config{
			Raw:    orgGroupPolicyConfigValueNullTier(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE)),
			Schema: s,
		},
	}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("did not expect an error for policy_type=role with a null enforcement_tier: %v", resp.Diagnostics)
	}
}

func TestOrgGroupPolicyModifyPlan_RoleRejectsReEnableFromDelegate(t *testing.T) {
	s := orgGroupPolicyValidateConfigSchema()
	r := &OrgGroupPolicyResource{}

	req := resource.ModifyPlanRequest{
		State: tfsdk.State{
			Raw:    orgGroupPolicyStateValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE), "DELEGATE"),
			Schema: s,
		},
		Plan: tfsdk.Plan{
			Raw:    orgGroupPolicyStateValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE), "GROUP_MANAGED"),
			Schema: s,
		},
	}
	resp := &resource.ModifyPlanResponse{}
	r.ModifyPlan(context.Background(), req, resp)

	assertDiagnosticSummary(t, resp.Diagnostics, "Cannot re-enable a disabled role policy")
}

func TestOrgGroupPolicyModifyPlan_SkipsReEnableGuardWhenReplaceRequired(t *testing.T) {
	s := orgGroupPolicyValidateConfigSchema()
	r := &OrgGroupPolicyResource{}

	// policy_type changing away from "role" forces a replace; the new resource's
	// enforcement_tier belongs to the replacement, not an update of the DELEGATE one.
	req := resource.ModifyPlanRequest{
		State: tfsdk.State{
			Raw:    orgGroupPolicyStateValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE), "DELEGATE"),
			Schema: s,
		},
		Plan: tfsdk.Plan{
			Raw:    orgGroupPolicyStateValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ORG_CONFIG), "GROUP_MANAGED"),
			Schema: s,
		},
	}
	// The framework always hands ModifyPlan an empty RequiresReplace, so the guard has to
	// infer the replace from the values themselves.
	resp := &resource.ModifyPlanResponse{RequiresReplace: frameworkPath.Paths{}}
	r.ModifyPlan(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("did not expect the re-enable guard to fire when a replace is already required: %v", resp.Diagnostics)
	}
}

func TestOrgGroupPolicyModifyPlan_SkipsReEnableGuardOnOrgGroupChange(t *testing.T) {
	s := orgGroupPolicyValidateConfigSchema()
	r := &OrgGroupPolicyResource{}

	// org_group_id also carries RequiresReplace, so moving the policy to another group is
	// a destroy/create whose enforcement_tier belongs to the replacement.
	req := resource.ModifyPlanRequest{
		State: tfsdk.State{
			Raw:    orgGroupPolicyStateValueOrgGroup(tftypes.NewValue(tftypes.String, orgGroupPolicyTestGroupID), string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE), "DELEGATE"),
			Schema: s,
		},
		Plan: tfsdk.Plan{
			Raw:    orgGroupPolicyStateValueOrgGroup(tftypes.NewValue(tftypes.String, "ffffffff-e5f6-7890-abcd-ef0123456789"), string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE), "GROUP_MANAGED"),
			Schema: s,
		},
	}
	resp := &resource.ModifyPlanResponse{RequiresReplace: frameworkPath.Paths{}}
	r.ModifyPlan(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("did not expect the re-enable guard to fire on an org_group_id change: %v", resp.Diagnostics)
	}
}

func TestOrgGroupPolicyModifyPlan_SkipsReEnableGuardOnUnknownOrgGroup(t *testing.T) {
	s := orgGroupPolicyValidateConfigSchema()
	r := &OrgGroupPolicyResource{}

	// org_group_id has no UseStateForUnknown, so it is unknown at plan time while it
	// interpolates a not-yet-created org group; that may still be a replace.
	req := resource.ModifyPlanRequest{
		State: tfsdk.State{
			Raw:    orgGroupPolicyStateValueOrgGroup(tftypes.NewValue(tftypes.String, orgGroupPolicyTestGroupID), string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE), "DELEGATE"),
			Schema: s,
		},
		Plan: tfsdk.Plan{
			Raw:    orgGroupPolicyStateValueOrgGroup(tftypes.NewValue(tftypes.String, tftypes.UnknownValue), string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE), "GROUP_MANAGED"),
			Schema: s,
		},
	}
	resp := &resource.ModifyPlanResponse{RequiresReplace: frameworkPath.Paths{}}
	r.ModifyPlan(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("did not expect the re-enable guard to fire on an unknown org_group_id: %v", resp.Diagnostics)
	}
}

func TestOrgGroupPolicyType_ChangeRequiresReplace(t *testing.T) {
	s := orgGroupPolicyValidateConfigSchema()
	modifier := stringplanmodifier.RequiresReplace()

	stateValue := types.StringValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE))
	planValue := types.StringValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ORG_CONFIG))

	req := planmodifier.StringRequest{
		State: tfsdk.State{
			Raw:    orgGroupPolicyStateValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE), "GROUP_MANAGED"),
			Schema: s,
		},
		Plan: tfsdk.Plan{
			Raw:    orgGroupPolicyConfigValue(string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ORG_CONFIG), "OVERRIDE_ALLOWED"),
			Schema: s,
		},
		StateValue: stateValue,
		PlanValue:  planValue,
	}
	resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
	modifier.PlanModifyString(context.Background(), req, resp)

	if !resp.RequiresReplace {
		t.Fatal("changing policy_type should force a replace")
	}
}
