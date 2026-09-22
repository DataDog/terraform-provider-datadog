package datadog

import (
	"testing"
)

// TestExactMatchPriority tests that exact match takes priority over partial match
// This is a unit test for the logic introduced to fix issue #3370
func TestExactMatchPriority(t *testing.T) {
	testCases := []struct {
		name         string
		searchedName string
		patterns     []string
		expectExact  bool
		expectedName string
		expectError  bool
	}{
		{
			name:         "exact_match_found",
			searchedName: "US Tax Identification Number Scanner",
			patterns: []string{
				"Cyprus Tax Identification Number Scanner",
				"US Tax Identification Number Scanner",
				"Australia Tax File Number Scanner",
			},
			expectExact:  true,
			expectedName: "US Tax Identification Number Scanner",
			expectError:  false,
		},
		{
			name:         "exact_match_priority_over_partial",
			searchedName: "AWS Access Key ID Scanner",
			patterns: []string{
				"AWS Access Key ID Scanner",
				"AWS Secret Access Key Scanner",
			},
			expectExact:  true,
			expectedName: "AWS Access Key ID Scanner",
			expectError:  false,
		},
		{
			name:         "partial_match_single",
			searchedName: "Cyprus",
			patterns: []string{
				"Cyprus Tax Identification Number Scanner",
				"US Tax Identification Number Scanner",
			},
			expectExact:  false,
			expectedName: "Cyprus Tax Identification Number Scanner",
			expectError:  false,
		},
		{
			name:         "partial_match_multiple_error",
			searchedName: "Tax",
			patterns: []string{
				"Cyprus Tax Identification Number Scanner",
				"US Tax Identification Number Scanner",
			},
			expectExact: false,
			expectError: true,
		},
		{
			name:         "no_match",
			searchedName: "NonExistent Pattern",
			patterns: []string{
				"Cyprus Tax Identification Number Scanner",
				"US Tax Identification Number Scanner",
			},
			expectExact: false,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exactMatch, partialMatches := simulateFilterLogic(tc.searchedName, tc.patterns)

			if tc.expectExact {
				if exactMatch == "" {
					t.Errorf("Expected exact match for '%s', but got none", tc.searchedName)
				}
				if exactMatch != tc.expectedName {
					t.Errorf("Expected exact match '%s', but got '%s'", tc.expectedName, exactMatch)
				}
			} else if tc.expectError {
				if exactMatch != "" {
					t.Errorf("Expected no exact match, but got '%s'", exactMatch)
				}
				if len(partialMatches) == 1 {
					t.Errorf("Expected error (0 or >1 matches), but got exactly 1 match")
				}
			} else {
				if len(partialMatches) != 1 {
					t.Errorf("Expected 1 partial match, but got %d", len(partialMatches))
				}
				if partialMatches[0] != tc.expectedName {
					t.Errorf("Expected partial match '%s', but got '%s'", tc.expectedName, partialMatches[0])
				}
			}
		})
	}
}

// simulateFilterLogic simulates the exact match priority logic
func simulateFilterLogic(searchedName string, patterns []string) (exactMatch string, partialMatches []string) {
	import_strings := func(s, substr string) bool {
		// Case-insensitive contains
		sLower := toLower(s)
		substrLower := toLower(substr)
		return contains(sLower, substrLower)
	}

	for _, name := range patterns {
		// Exact match takes priority
		if name == searchedName {
			return name, nil
		}

		// Collect partial matches
		if import_strings(name, searchedName) {
			partialMatches = append(partialMatches, name)
		}
	}

	return "", partialMatches
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

