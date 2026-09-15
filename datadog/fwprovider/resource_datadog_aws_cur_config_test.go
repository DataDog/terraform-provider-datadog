package fwprovider

import (
	"context"
	"encoding/json"
	"testing"
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
