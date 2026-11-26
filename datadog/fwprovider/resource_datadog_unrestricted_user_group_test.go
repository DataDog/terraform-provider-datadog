package fwprovider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const testUnrestrictedUserGroupPrincipal = "role:00000000-0000-1111-0000-000000000000"

func TestBuildUnrestrictedUserGroupRequest(t *testing.T) {
	ctx := context.Background()
	state := unrestrictedUserGroupModel{
		PrincipalID: types.StringValue(testUnrestrictedUserGroupPrincipal),
		Products: types.SetValueMust(types.StringType, []attr.Value{
			types.StringValue("monitor"),
			types.StringValue("logs"),
		}),
	}

	body, diags := buildUnrestrictedUserGroupRequest(ctx, &state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	got, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	// Products are sorted so the body is deterministic.
	want := `{"data":{"attributes":{"products":[{"name":"logs"},{"name":"monitor"}]},"id":"role:00000000-0000-1111-0000-000000000000","type":"unrestricted_user_groups"}}`
	if string(got) != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestFindUnrestrictedUserGroup(t *testing.T) {
	groups := []datadogV2.UnrestrictedUserGroup{
		*datadogV2.NewUnrestrictedUserGroup(datadogV2.UnrestrictedUserGroupAttributes{}, "team:00000000-0000-2222-0000-000000000000", datadogV2.UNRESTRICTEDUSERGROUPTYPE_UNRESTRICTED_USER_GROUPS),
		*datadogV2.NewUnrestrictedUserGroup(datadogV2.UnrestrictedUserGroupAttributes{}, testUnrestrictedUserGroupPrincipal, datadogV2.UNRESTRICTEDUSERGROUPTYPE_UNRESTRICTED_USER_GROUPS),
	}

	if group := findUnrestrictedUserGroup(groups, testUnrestrictedUserGroupPrincipal); group == nil || group.GetId() != testUnrestrictedUserGroupPrincipal {
		t.Errorf("expected to find %s, got %v", testUnrestrictedUserGroupPrincipal, group)
	}
	if group := findUnrestrictedUserGroup(groups, "user:00000000-0000-3333-0000-000000000000"); group != nil {
		t.Errorf("expected no group, got %s", group.GetId())
	}
}

func TestUpdateUnrestrictedUserGroupState(t *testing.T) {
	ctx := context.Background()
	group := datadogV2.NewUnrestrictedUserGroup(
		*datadogV2.NewUnrestrictedUserGroupAttributes([]datadogV2.UnrestrictedUserGroupProduct{
			*datadogV2.NewUnrestrictedUserGroupProduct("apm"),
			*datadogV2.NewUnrestrictedUserGroupProduct("logs"),
		}),
		testUnrestrictedUserGroupPrincipal,
		datadogV2.UNRESTRICTEDUSERGROUPTYPE_UNRESTRICTED_USER_GROUPS,
	)

	// An imported resource only has its ID in state.
	state := unrestrictedUserGroupModel{ID: types.StringValue(testUnrestrictedUserGroupPrincipal)}
	if diags := updateUnrestrictedUserGroupState(ctx, &state, group); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if state.PrincipalID.ValueString() != testUnrestrictedUserGroupPrincipal {
		t.Errorf("principal_id = %s, want %s", state.PrincipalID.ValueString(), testUnrestrictedUserGroupPrincipal)
	}
	want := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("apm"), types.StringValue("logs")})
	if !state.Products.Equal(want) {
		t.Errorf("products = %s, want %s", state.Products, want)
	}
}
