package datadog

import (
	"context"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestResourceDatadogDashboardCreateDisallowed(t *testing.T) {
	t.Setenv("DD_TERRAFORM_DISALLOW_DASHBOARD_V1_CREATE", "true")

	diags := resourceDatadogDashboardCreate(context.Background(), nil, nil)
	if !diags.HasError() {
		t.Fatal("expected v1 dashboard creation to be rejected")
	}

	if got := diags[0].Summary; !strings.Contains(got, "use datadog_dashboard_v2 instead") {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestDisallowDashboardV1CreateDefaultsToFalse(t *testing.T) {
	t.Setenv("DD_TERRAFORM_DISALLOW_DASHBOARD_V1_CREATE", "")

	if utils.DisallowDashboardV1Create() {
		t.Fatal("expected v1 dashboard creation to be allowed by default")
	}
}

func TestDisallowDashboardV1CreateEnabled(t *testing.T) {
	t.Setenv("DD_TERRAFORM_DISALLOW_DASHBOARD_V1_CREATE", "true")

	if !utils.DisallowDashboardV1Create() {
		t.Fatal("expected v1 dashboard creation to be disabled")
	}
}

func TestValidateDashboardV1OperationAllowsExistingUpdate(t *testing.T) {
	t.Setenv("DD_TERRAFORM_DISALLOW_DASHBOARD_V1_CREATE", "true")

	if err := validateDashboardV1Operation("abc-123", "ordered", "ordered"); err != nil {
		t.Fatalf("expected an existing v1 dashboard update to be allowed: %s", err)
	}
}

func TestValidateDashboardV1OperationRejectsReplacement(t *testing.T) {
	t.Setenv("DD_TERRAFORM_DISALLOW_DASHBOARD_V1_CREATE", "true")

	err := validateDashboardV1Operation("abc-123", "ordered", "free")
	if err == nil {
		t.Fatal("expected a v1 dashboard replacement to be rejected")
	}
	if !strings.Contains(err.Error(), "migrate to datadog_dashboard_v2") {
		t.Fatalf("unexpected error: %s", err)
	}
}
