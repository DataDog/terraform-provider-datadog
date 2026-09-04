package fwprovider

import (
	"context"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func orgGroupPolicyValidateConfigSchema() schema.Schema {
	r := &OrgGroupPolicyResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	return resp.Schema
}

func orgGroupPolicyConfigValue(policyType, enforcementTier string) tftypes.Value {
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"org_group_id":     tftypes.String,
			"policy_name":      tftypes.String,
			"content":          tftypes.String,
			"enforcement_tier": tftypes.String,
			"policy_type":      tftypes.String,
		},
	}
	return tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, nil),
		"org_group_id":     tftypes.NewValue(tftypes.String, "a1b2c3d4-e5f6-7890-abcd-ef0123456789"),
		"policy_name":      tftypes.NewValue(tftypes.String, "finance_read_only"),
		"content":          tftypes.NewValue(tftypes.String, `{"permissions":["1a2b3c4d-5e6f-7890-abcd-ef0123456789"]}`),
		"enforcement_tier": tftypes.NewValue(tftypes.String, enforcementTier),
		"policy_type":      tftypes.NewValue(tftypes.String, policyType),
	})
}

func orgGroupPolicyStateValue(policyType, enforcementTier string) tftypes.Value {
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"org_group_id":     tftypes.String,
			"policy_name":      tftypes.String,
			"content":          tftypes.String,
			"enforcement_tier": tftypes.String,
			"policy_type":      tftypes.String,
		},
	}
	return tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, "b2c3d4e5-f6a7-8901-bcde-f01234567890"),
		"org_group_id":     tftypes.NewValue(tftypes.String, "a1b2c3d4-e5f6-7890-abcd-ef0123456789"),
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

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error deleting a role policy with enforcement_tier != DELEGATE")
	}
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

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error for policy_type=role with enforcement_tier=OVERRIDE_ALLOWED")
	}
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
			Raw:    orgGroupPolicyConfigValue("org_config", "OVERRIDE_ALLOWED"),
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
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":               tftypes.String,
			"org_group_id":     tftypes.String,
			"policy_name":      tftypes.String,
			"content":          tftypes.String,
			"enforcement_tier": tftypes.String,
			"policy_type":      tftypes.String,
		},
	}
	return tftypes.NewValue(objType, map[string]tftypes.Value{
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

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error re-enabling a disabled role policy (DELEGATE -> GROUP_MANAGED)")
	}
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
			Raw:    orgGroupPolicyStateValue("org_config", "OVERRIDE_ALLOWED"),
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
