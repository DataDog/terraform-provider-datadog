package fwprovider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestAwsCurConfigUpdateRequestClearsOmittedAccountFilters(t *testing.T) {
	r := &awsCurConfigResource{}

	req, diags := r.buildAwsCurConfigUpdateRequestBody(
		context.Background(),
		&awsCurConfigModel{},
	)
	if diags.HasError() {
		t.Fatalf("building update request returned diagnostics: %v", diags)
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshaling update request: %v", err)
	}

	var payload struct {
		Data struct {
			Attributes map[string]json.RawMessage `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshaling update request: %v", err)
	}

	accountFilters, ok := payload.Data.Attributes["account_filters"]
	if !ok {
		t.Fatalf("expected account_filters to be sent explicitly in %s", body)
	}
	if got, want := string(accountFilters), `{}`; got != want {
		t.Fatalf("account_filters = %s, want %s", got, want)
	}
}

func TestAwsCurConfigUpdateRequestPreservesAccountFilters(t *testing.T) {
	tests := []struct {
		name    string
		filters *accountFiltersModel
		want    string
	}{
		{
			name: "excluded accounts",
			filters: &accountFiltersModel{
				IncludeNewAccounts: types.BoolValue(true),
				ExcludedAccounts:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("123456789012")}),
				IncludedAccounts:   types.ListNull(types.StringType),
			},
			want: `{"include_new_accounts":true,"excluded_accounts":["123456789012"]}`,
		},
		{
			name: "included accounts",
			filters: &accountFiltersModel{
				IncludeNewAccounts: types.BoolValue(false),
				ExcludedAccounts:   types.ListNull(types.StringType),
				IncludedAccounts:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("123456789013")}),
			},
			want: `{"include_new_accounts":false,"included_accounts":["123456789013"]}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := &awsCurConfigResource{}
			req, diags := r.buildAwsCurConfigUpdateRequestBody(context.Background(), &awsCurConfigModel{AccountFilters: test.filters})
			require.False(t, diags.HasError(), "building update request: %v", diags)
			body, err := json.Marshal(req)
			require.NoError(t, err)
			require.JSONEq(t, `{"data":{"type":"aws_cur_config_patch_request","attributes":{"account_filters":`+test.want+`}}}`, string(body))
		})
	}
}
