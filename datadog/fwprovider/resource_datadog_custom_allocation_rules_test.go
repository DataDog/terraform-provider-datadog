package fwprovider

import (
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

// Helper function to create a rule with normal fields
func createCustomAllocationRule(id string, name string, orderId int64) datadogV2.ArbitraryRuleResponseData {
	rule := datadogV2.ArbitraryRuleResponseData{
		Id: &id,
	}
	// Note: We'd set attributes here but the API client struct is complex
	// For these tests, we'll focus on the UnparsedObject case which is what actually matters
	return rule
}

// Helper function to create a rule with UnparsedObject (simulating API client deserialization issue)
func createCustomAllocationRuleWithUnparsed(id string, name string, orderId int64) datadogV2.ArbitraryRuleResponseData {
	// Simulate the case where the API client failed to deserialize and put data in UnparsedObject
	rule := datadogV2.ArbitraryRuleResponseData{
		UnparsedObject: map[string]interface{}{
			"id":   id,
			"type": "arbitrary_rule",
			"attributes": map[string]interface{}{
				"rule_name": name,
				"order_id":  orderId,
			},
		},
	}
	return rule
}

// Test helper functions for extracting rule data

func TestExtractRuleID_NormalCase(t *testing.T) {
	t.Run("Successfully extract ID from normal rule", func(t *testing.T) {
		id := "test-rule-id"
		rule := createCustomAllocationRule(id, "Test Rule", 1)

		extractedID, ok := extractRuleID(rule)

		if !ok {
			t.Error("Should successfully extract ID")
		}
		if extractedID != id {
			t.Errorf("Expected ID %s, got %s", id, extractedID)
		}
	})
}

func TestExtractRuleID_UnparsedObject(t *testing.T) {
	t.Run("Extract ID from UnparsedObject", func(t *testing.T) {
		id := "test-rule-id"
		rule := createCustomAllocationRuleWithUnparsed(id, "Test Rule", 1)

		extractedID, ok := extractRuleID(rule)

		if !ok {
			t.Error("Should successfully extract ID from UnparsedObject")
		}
		if extractedID != id {
			t.Errorf("Expected ID %s, got %s", id, extractedID)
		}
	})
}

func TestExtractRuleID_NoID(t *testing.T) {
	t.Run("Fail to extract ID when none exists", func(t *testing.T) {
		rule := datadogV2.ArbitraryRuleResponseData{}

		extractedID, ok := extractRuleID(rule)

		if ok {
			t.Error("Should fail to extract ID when none exists")
		}
		if extractedID != "" {
			t.Error("Extracted ID should be empty")
		}
	})
}

func TestExtractRuleID_EmptyString(t *testing.T) {
	t.Run("Fail to extract empty string ID", func(t *testing.T) {
		emptyID := ""
		rule := datadogV2.ArbitraryRuleResponseData{
			Id: &emptyID,
		}

		extractedID, ok := extractRuleID(rule)

		if ok {
			t.Error("Should fail to extract empty ID")
		}
		if extractedID != "" {
			t.Error("Extracted ID should be empty")
		}
	})
}

func TestExtractRuleFields_UnparsedObject(t *testing.T) {
	t.Run("Extract rule_name from UnparsedObject", func(t *testing.T) {
		id := "test-id"
		name := "Test Rule Name"
		orderId := int64(5)
		rule := createCustomAllocationRuleWithUnparsed(id, name, orderId)

		extractedID, extractedName, extractedOrderId, ok := extractRuleFields(rule)

		if !ok {
			t.Error("Should successfully extract fields")
		}
		if extractedID != id {
			t.Errorf("Expected ID %s, got %s", id, extractedID)
		}
		if extractedName != name {
			t.Errorf("Expected name %s from UnparsedObject, got %s", name, extractedName)
		}
		if extractedOrderId != orderId {
			t.Errorf("Expected order_id %d, got %d", orderId, extractedOrderId)
		}
	})
}

func TestExtractRuleFields_NoFields(t *testing.T) {
	t.Run("Return empty values when no fields exist", func(t *testing.T) {
		rule := datadogV2.ArbitraryRuleResponseData{}

		extractedID, extractedName, extractedOrderId, ok := extractRuleFields(rule)

		if ok {
			t.Error("Should fail to extract when no fields exist")
		}
		if extractedID != "" {
			t.Errorf("Expected empty ID, got %s", extractedID)
		}
		if extractedName != "" {
			t.Errorf("Expected empty name, got %s", extractedName)
		}
		if extractedOrderId != 0 {
			t.Errorf("Expected order_id 0, got %d", extractedOrderId)
		}
	})
}

func TestExtractRuleFields_PartialFields(t *testing.T) {
	t.Run("Extract fields when some are missing", func(t *testing.T) {
		id := "partial-test-id"
		rule := datadogV2.ArbitraryRuleResponseData{
			UnparsedObject: map[string]interface{}{
				"id":   id,
				"type": "arbitrary_rule",
				"attributes": map[string]interface{}{
					// rule_name is missing, only order_id is present
					"order_id": int64(3),
				},
			},
		}

		extractedID, extractedName, extractedOrderId, ok := extractRuleFields(rule)

		if !ok {
			t.Error("Should successfully extract ID even if other fields are missing")
		}
		if extractedID != id {
			t.Errorf("Expected ID %s, got %s", id, extractedID)
		}
		if extractedName != "" {
			t.Errorf("Expected empty name when missing, got %s", extractedName)
		}
		if extractedOrderId != 3 {
			t.Errorf("Expected order_id 3, got %d", extractedOrderId)
		}
	})
}

func TestExtractRuleFields_DifferentNumericTypes(t *testing.T) {
	t.Run("Extract order_id from different numeric types", func(t *testing.T) {
		tests := []struct {
			name     string
			orderVal interface{}
			expected int64
		}{
			{"float64", float64(42), 42},
			{"int", int(42), 42},
			{"int64", int64(42), 42},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				rule := datadogV2.ArbitraryRuleResponseData{
					UnparsedObject: map[string]interface{}{
						"id":   "test-id",
						"type": "arbitrary_rule",
						"attributes": map[string]interface{}{
							"rule_name": "Test Rule",
							"order_id":  test.orderVal,
						},
					},
				}

				_, _, extractedOrderId, ok := extractRuleFields(rule)

				if !ok {
					t.Error("Should successfully extract fields")
				}
				if extractedOrderId != test.expected {
					t.Errorf("Expected order_id %d from %s, got %d", test.expected, test.name, extractedOrderId)
				}
			})
		}
	})
}

func TestGetRulesWithPositions_EmptyList(t *testing.T) {
	t.Run("Handle empty rules list", func(t *testing.T) {
		rules := []datadogV2.ArbitraryRuleResponseData{}
		managedIDs := make(map[string]bool)
		
		result := getRulesWithPositions(rules, managedIDs, false)

		if len(result) != 0 {
			t.Errorf("Expected empty result, got %d rules", len(result))
		}
	})
}

func TestGetRulesWithPositions_SingleRule(t *testing.T) {
	t.Run("Handle single rule", func(t *testing.T) {
		rule := createCustomAllocationRuleWithUnparsed("rule-1", "Rule One", 1)
		rules := []datadogV2.ArbitraryRuleResponseData{rule}
		managedIDs := map[string]bool{"rule-1": true}
		
		result := getRulesWithPositions(rules, managedIDs, false)

		if len(result) != 1 {
			t.Errorf("Expected 1 rule, got %d", len(result))
		}
		if result[0].ID != "rule-1" {
			t.Errorf("Expected ID rule-1, got %s", result[0].ID)
		}
		if result[0].Name != "Rule One" {
			t.Errorf("Expected name 'Rule One', got %s", result[0].Name)
		}
		if result[0].Position != 1 {
			t.Errorf("Expected position 1, got %d", result[0].Position)
		}
	})
}

func TestGetRulesWithPositions_ManagedOnly(t *testing.T) {
	t.Run("Filter to managed rules only", func(t *testing.T) {
		rule1 := createCustomAllocationRuleWithUnparsed("managed-1", "Managed Rule", 1)
		rule2 := createCustomAllocationRuleWithUnparsed("unmanaged-1", "Unmanaged Rule", 2)
		rules := []datadogV2.ArbitraryRuleResponseData{rule1, rule2}
		managedIDs := map[string]bool{"managed-1": true}
		
		result := getRulesWithPositions(rules, managedIDs, true) // managedOnly = true

		if len(result) != 1 {
			t.Errorf("Expected 1 managed rule, got %d", len(result))
		}
		if result[0].ID != "managed-1" {
			t.Errorf("Expected managed rule ID, got %s", result[0].ID)
		}
	})
}

func TestGetRulesWithPositions_AllRules(t *testing.T) {
	t.Run("Include all rules regardless of managed status", func(t *testing.T) {
		rule1 := createCustomAllocationRuleWithUnparsed("managed-1", "Managed Rule", 2)
		rule2 := createCustomAllocationRuleWithUnparsed("unmanaged-1", "Unmanaged Rule", 1)
		rules := []datadogV2.ArbitraryRuleResponseData{rule1, rule2}
		managedIDs := map[string]bool{"managed-1": true}
		
		result := getRulesWithPositions(rules, managedIDs, false) // managedOnly = false

		if len(result) != 2 {
			t.Errorf("Expected 2 rules, got %d", len(result))
		}
		// Should be sorted by position: unmanaged-1 (pos 1), managed-1 (pos 2)
		if result[0].ID != "unmanaged-1" {
			t.Errorf("Expected first rule to be unmanaged-1, got %s", result[0].ID)
		}
		if result[1].ID != "managed-1" {
			t.Errorf("Expected second rule to be managed-1, got %s", result[1].ID)
		}
	})
}

func TestSortRulesByPosition(t *testing.T) {
	t.Run("Sort rules by position in ascending order", func(t *testing.T) {
		rules := []ruleWithPosition{
			{ID: "rule-3", Position: 3, Name: "Third"},
			{ID: "rule-1", Position: 1, Name: "First"},
			{ID: "rule-2", Position: 2, Name: "Second"},
		}
		
		sortRulesByPosition(rules)

		expectedOrder := []string{"rule-1", "rule-2", "rule-3"}
		for i, rule := range rules {
			if rule.ID != expectedOrder[i] {
				t.Errorf("Position %d: expected %s, got %s", i, expectedOrder[i], rule.ID)
			}
		}
	})
}

func TestFindUnmanagedRules_AllManaged(t *testing.T) {
	t.Run("No unmanaged rules when all are managed", func(t *testing.T) {
		rules := []ruleWithPosition{
			{ID: "managed-1", Position: 1, Name: "Rule 1"},
			{ID: "managed-2", Position: 2, Name: "Rule 2"},
		}
		managedIDs := map[string]bool{
			"managed-1": true,
			"managed-2": true,
		}
		
		result := findUnmanagedRules(rules, managedIDs)

		if len(result.Rules) != 0 {
			t.Errorf("Expected no unmanaged rules, got %d", len(result.Rules))
		}
		if result.AllAtEnd != false {
			t.Error("AllAtEnd should be false when no unmanaged rules exist")
		}
	})
}

func TestFindUnmanagedRules_AtEnd(t *testing.T) {
	t.Run("Unmanaged rules at the end", func(t *testing.T) {
		rules := []ruleWithPosition{
			{ID: "managed-1", Position: 1, Name: "Rule 1"},
			{ID: "managed-2", Position: 2, Name: "Rule 2"},
			{ID: "unmanaged-1", Position: 3, Name: "Rule 3"},
			{ID: "unmanaged-2", Position: 4, Name: "Rule 4"},
		}
		managedIDs := map[string]bool{
			"managed-1": true,
			"managed-2": true,
		}
		
		result := findUnmanagedRules(rules, managedIDs)

		if len(result.Rules) != 2 {
			t.Errorf("Expected 2 unmanaged rules, got %d", len(result.Rules))
		}
		if !result.AllAtEnd {
			t.Error("AllAtEnd should be true when unmanaged rules are at the end")
		}
		expectedPositions := []int{2, 3} // 0-based positions in the array
		for i, pos := range result.Positions {
			if pos != expectedPositions[i] {
				t.Errorf("Expected position %d, got %d", expectedPositions[i], pos)
			}
		}
	})
}

func TestFindUnmanagedRules_InMiddle(t *testing.T) {
	t.Run("Unmanaged rules in the middle", func(t *testing.T) {
		rules := []ruleWithPosition{
			{ID: "managed-1", Position: 1, Name: "Rule 1"},
			{ID: "unmanaged-1", Position: 2, Name: "Rule 2"},
			{ID: "managed-2", Position: 3, Name: "Rule 3"},
		}
		managedIDs := map[string]bool{
			"managed-1": true,
			"managed-2": true,
		}
		
		result := findUnmanagedRules(rules, managedIDs)

		if len(result.Rules) != 1 {
			t.Errorf("Expected 1 unmanaged rule, got %d", len(result.Rules))
		}
		if result.AllAtEnd {
			t.Error("AllAtEnd should be false when unmanaged rules are in the middle")
		}
		if result.Rules[0].ID != "unmanaged-1" {
			t.Errorf("Expected unmanaged rule ID unmanaged-1, got %s", result.Rules[0].ID)
		}
	})
}

func TestFormatUnmanagedRuleDetails_WithPositions(t *testing.T) {
	t.Run("Format unmanaged rule details with positions", func(t *testing.T) {
		rules := []ruleWithPosition{
			{ID: "rule-1", Position: 1, Name: "Test Rule 1"},
			{ID: "rule-2", Position: 2, Name: ""},
		}
		
		details := formatUnmanagedRuleDetails(rules, true)

		expected := []string{
			"'Test Rule 1' (ID: rule-1, Position: 1)",
			"ID: rule-2 (Position: 2)",
		}
		
		if len(details) != len(expected) {
			t.Errorf("Expected %d details, got %d", len(expected), len(details))
		}
		for i, detail := range details {
			if detail != expected[i] {
				t.Errorf("Detail %d: expected %s, got %s", i, expected[i], detail)
			}
		}
	})
}

func TestFormatUnmanagedRuleDetails_WithoutPositions(t *testing.T) {
	t.Run("Format unmanaged rule details without positions", func(t *testing.T) {
		rules := []ruleWithPosition{
			{ID: "rule-1", Position: 1, Name: "Test Rule 1"},
			{ID: "rule-2", Position: 2, Name: ""},
		}
		
		details := formatUnmanagedRuleDetails(rules, false)

		expected := []string{
			"'Test Rule 1' (rule-1)",
			"rule-2",
		}
		
		if len(details) != len(expected) {
			t.Errorf("Expected %d details, got %d", len(expected), len(details))
		}
		for i, detail := range details {
			if detail != expected[i] {
				t.Errorf("Detail %d: expected %s, got %s", i, expected[i], detail)
			}
		}
	})
}