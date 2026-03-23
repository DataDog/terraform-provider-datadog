package datadog

import (
	"testing"

	"github.com/google/uuid"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

func TestBuildTerraformTabs_IntegerWidgetIds(t *testing.T) {
	// Simulate API response with arbitrary integer widget IDs
	widgets := []datadogV1.Widget{
		{Id: Ptr[int64](1234567890123456)},
		{Id: Ptr[int64](9876543210987654)},
		{Id: Ptr[int64](5555555555555555)},
	}

	tabs := []datadogV1.DashboardTab{
		{
			Id:        uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"),
			Name:      "Overview",
			WidgetIds: []int64{1234567890123456, 9876543210987654},
		},
		{
			Id:        uuid.MustParse("11111111-2222-3333-4444-555555555555"),
			Name:      "Details",
			WidgetIds: []int64{5555555555555555},
		},
	}

	result := buildTerraformTabs(tabs, &widgets)

	if len(result) != 2 {
		t.Fatalf("expected 2 tabs, got %d", len(result))
	}

	// Tab 0: Overview with widgets @1 and @2
	if result[0]["name"] != "Overview" {
		t.Errorf("tab 0 name: expected Overview, got %v", result[0]["name"])
	}
	refs0 := result[0]["widget_ids"].([]string)
	if len(refs0) != 2 || refs0[0] != "@1" || refs0[1] != "@2" {
		t.Errorf("tab 0 widget_ids: expected [@1 @2], got %v", refs0)
	}

	// Tab 1: Details with widget @3
	if result[1]["name"] != "Details" {
		t.Errorf("tab 1 name: expected Details, got %v", result[1]["name"])
	}
	refs1 := result[1]["widget_ids"].([]string)
	if len(refs1) != 1 || refs1[0] != "@3" {
		t.Errorf("tab 1 widget_ids: expected [@3], got %v", refs1)
	}

	// Tab IDs should be stringified UUIDs
	if result[0]["id"] != "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" {
		t.Errorf("tab 0 id: expected aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee, got %v", result[0]["id"])
	}
}

func TestBuildTerraformTabs_SingleTabAllWidgets(t *testing.T) {
	widgets := []datadogV1.Widget{
		{Id: Ptr[int64](1111111111111111)},
		{Id: Ptr[int64](2222222222222222)},
		{Id: Ptr[int64](3333333333333333)},
	}

	tabs := []datadogV1.DashboardTab{
		{
			Id:        uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"),
			Name:      "All Widgets",
			WidgetIds: []int64{1111111111111111, 2222222222222222, 3333333333333333},
		},
	}

	result := buildTerraformTabs(tabs, &widgets)

	if len(result) != 1 {
		t.Fatalf("expected 1 tab, got %d", len(result))
	}
	refs := result[0]["widget_ids"].([]string)
	if len(refs) != 3 || refs[0] != "@1" || refs[1] != "@2" || refs[2] != "@3" {
		t.Errorf("expected [@1 @2 @3], got %v", refs)
	}
}

func TestBuildTerraformTabs_UnknownWidgetIdSkipped(t *testing.T) {
	widgets := []datadogV1.Widget{
		{Id: Ptr[int64](1111111111111111)},
	}

	tabs := []datadogV1.DashboardTab{
		{
			Id:   uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"),
			Name: "Mixed",
			// 9999 doesn't match any widget — should be silently skipped
			WidgetIds: []int64{1111111111111111, 9999999999999999},
		},
	}

	result := buildTerraformTabs(tabs, &widgets)

	refs := result[0]["widget_ids"].([]string)
	if len(refs) != 1 || refs[0] != "@1" {
		t.Errorf("expected [@1] (unknown ID skipped), got %v", refs)
	}
}

func TestBuildTerraformTabs_EmptyTabs(t *testing.T) {
	widgets := []datadogV1.Widget{
		{Id: Ptr[int64](1111111111111111)},
	}

	result := buildTerraformTabs([]datadogV1.DashboardTab{}, &widgets)
	if len(result) != 0 {
		t.Errorf("expected 0 tabs, got %d", len(result))
	}
}

