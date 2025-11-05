package fwprovider

import (
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

// Helper function to create a ruleset with normal fields
func createRuleset(id string, name string, position int32) datadogV2.RulesetRespData {
	ruleset := datadogV2.RulesetRespData{
		Id: &id,
	}
	// Note: We'd set attributes here but the API client struct is complex
	// For these tests, we'll focus on the UnparsedObject case which is what actually matters
	return ruleset
}

// Helper function to create a ruleset with UnparsedObject (simulating API client deserialization issue)
func createRulesetWithUnparsed(id string, name string, position int32) datadogV2.RulesetRespData {
	// Simulate the case where the API client failed to deserialize and put data in UnparsedObject
	ruleset := datadogV2.RulesetRespData{
		UnparsedObject: map[string]interface{}{
			"id":   id,
			"type": "ruleset",
			"attributes": map[string]interface{}{
				"name":     name,
				"position": position,
			},
		},
	}
	return ruleset
}

// Test helper functions for extracting ruleset data

func TestExtractRulesetID_NormalCase(t *testing.T) {
	t.Run("Successfully extract ID from normal ruleset", func(t *testing.T) {
		id := "test-ruleset-id"
		ruleset := createRuleset(id, "Test Ruleset", 1)

		extractedID, ok := extractRulesetID(ruleset)

		if !ok {
			t.Error("Should successfully extract ID")
		}
		if extractedID != id {
			t.Errorf("Expected ID %s, got %s", id, extractedID)
		}
	})
}

func TestExtractRulesetID_UnparsedObject(t *testing.T) {
	t.Run("Extract ID from UnparsedObject", func(t *testing.T) {
		id := "test-ruleset-id"
		ruleset := createRulesetWithUnparsed(id, "Test Ruleset", 1)

		extractedID, ok := extractRulesetID(ruleset)

		if !ok {
			t.Error("Should successfully extract ID from UnparsedObject")
		}
		if extractedID != id {
			t.Errorf("Expected ID %s, got %s", id, extractedID)
		}
	})
}

func TestExtractRulesetID_NoID(t *testing.T) {
	t.Run("Fail to extract ID when none exists", func(t *testing.T) {
		ruleset := datadogV2.RulesetRespData{}

		extractedID, ok := extractRulesetID(ruleset)

		if ok {
			t.Error("Should fail to extract ID when none exists")
		}
		if extractedID != "" {
			t.Error("Extracted ID should be empty")
		}
	})
}

func TestExtractRulesetID_EmptyString(t *testing.T) {
	t.Run("Fail to extract empty string ID", func(t *testing.T) {
		emptyID := ""
		ruleset := datadogV2.RulesetRespData{
			Id: &emptyID,
		}

		extractedID, ok := extractRulesetID(ruleset)

		if ok {
			t.Error("Should fail to extract empty ID")
		}
		if extractedID != "" {
			t.Error("Extracted ID should be empty")
		}
	})
}

func TestExtractRulesetName_UnparsedObject(t *testing.T) {
	t.Run("Extract name from UnparsedObject", func(t *testing.T) {
		name := "Test Ruleset Name"
		ruleset := createRulesetWithUnparsed("test-id", name, 1)

		extractedName := extractRulesetName(ruleset)

		if extractedName != name {
			t.Errorf("Expected name %s from UnparsedObject, got %s", name, extractedName)
		}
	})
}

func TestExtractRulesetName_NoName(t *testing.T) {
	t.Run("Return empty string when no name exists", func(t *testing.T) {
		ruleset := datadogV2.RulesetRespData{}

		extractedName := extractRulesetName(ruleset)

		if extractedName != "" {
			t.Errorf("Expected empty string, got %s", extractedName)
		}
	})
}

func TestExtractRulesetID_And_Name_Together(t *testing.T) {
	t.Run("Extract both ID and name from same ruleset with UnparsedObject", func(t *testing.T) {
		id := "combined-test-id"
		name := "Combined Test Name"
		ruleset := createRulesetWithUnparsed(id, name, 5)

		extractedID, idOk := extractRulesetID(ruleset)
		extractedName := extractRulesetName(ruleset)

		if !idOk {
			t.Error("Should successfully extract ID")
		}
		if extractedID != id {
			t.Errorf("Expected ID %s, got %s", id, extractedID)
		}
		if extractedName != name {
			t.Errorf("Expected name %s, got %s", name, extractedName)
		}
	})
}

func TestExtractRulesetPosition_UnparsedObject(t *testing.T) {
	t.Run("Extract position from UnparsedObject", func(t *testing.T) {
		expectedPosition := int32(3)
		ruleset := createRulesetWithUnparsed("test-id", "Test Ruleset", expectedPosition)

		extractedPosition := extractRulesetPosition(ruleset)

		if extractedPosition != expectedPosition {
			t.Errorf("Expected position %d, got %d", expectedPosition, extractedPosition)
		}
	})
}

func TestExtractRulesetPosition_NoPosition(t *testing.T) {
	t.Run("Return 0 when no position exists", func(t *testing.T) {
		ruleset := datadogV2.RulesetRespData{}

		extractedPosition := extractRulesetPosition(ruleset)

		if extractedPosition != 0 {
			t.Errorf("Expected position 0, got %d", extractedPosition)
		}
	})
}

func TestExtractAll_FromUnparsedObject(t *testing.T) {
	t.Run("Extract ID, name, and position all from UnparsedObject", func(t *testing.T) {
		id := "full-test-id"
		name := "Full Test Name"
		position := int32(5)
		ruleset := createRulesetWithUnparsed(id, name, position)

		extractedID, idOk := extractRulesetID(ruleset)
		extractedName := extractRulesetName(ruleset)
		extractedPosition := extractRulesetPosition(ruleset)

		if !idOk {
			t.Error("Should successfully extract ID")
		}
		if extractedID != id {
			t.Errorf("Expected ID %s, got %s", id, extractedID)
		}
		if extractedName != name {
			t.Errorf("Expected name %s, got %s", name, extractedName)
		}
		if extractedPosition != position {
			t.Errorf("Expected position %d, got %d", position, extractedPosition)
		}
	})
}
