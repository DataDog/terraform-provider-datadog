package fwprovider

import (
	"context"
	"encoding/json"
	"testing"
)

func TestIntegrationAwsAccountCcmConfigRequestIncludesTerraformSetupMethod(t *testing.T) {
	r := &integrationAwsAccountCcmConfigResource{}

	req, diags := r.buildIntegrationAwsAccountCcmConfigRequestBody(
		context.Background(),
		&integrationAwsAccountCcmConfigModel{},
	)
	if diags.HasError() {
		t.Fatalf("building request returned diagnostics: %v", diags)
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshaling request: %v", err)
	}

	var payload struct {
		Meta map[string]string `json:"meta"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshaling request: %v", err)
	}
	if got := payload.Meta["setup_method"]; got != "terraform" {
		t.Fatalf("expected setup method %q, got %q in %s", "terraform", got, body)
	}
}
