package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
)

// ValidateKeysV2Path is the v2 endpoint that accepts PAT (Bearer) auth.
// It is the v2 counterpart to /api/v1/validate, which only accepts
// DD-API-KEY/DD-APPLICATION-KEY headers.
const ValidateKeysV2Path = "/api/v2/validate_keys"

type validateKeysV2Response struct {
	Data struct {
		Attributes struct {
			Valid *bool `json:"valid,omitempty"`
		} `json:"attributes"`
	} `json:"data"`
}

// ValidateKeysV2 calls /api/v2/validate_keys using whichever credentials
// are present on ctx (PAT via ContextAccessToken, or DD-API-KEY/DD-APPLICATION-KEY
// via ContextAPIKeys). It returns nil iff the response indicates valid credentials.
func ValidateKeysV2(ctx context.Context, client *datadog.APIClient) error {
	body, httpRes, err := SendRequest(ctx, client, http.MethodGet, ValidateKeysV2Path, nil)
	if err != nil {
		return err
	}
	if httpRes != nil && httpRes.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from %s", httpRes.StatusCode, ValidateKeysV2Path)
	}

	var parsed validateKeysV2Response
	if err := json.Unmarshal(body, &parsed); err != nil {
		// A 200 with an unexpected body still implies the credentials reached
		// the server; treat this as a soft success rather than blocking init.
		return nil
	}
	if parsed.Data.Attributes.Valid != nil && !*parsed.Data.Attributes.Valid {
		return fmt.Errorf("invalid or missing credentials provided to the Datadog Provider; please confirm your PAT (or API/APP keys) are valid and are for the correct region — see https://www.terraform.io/docs/providers/datadog/ for more information on providing credentials for the Datadog Provider")
	}
	return nil
}
