package web_integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// AccountAttributes holds the mutable fields of an integration account.
// NOTE: Used for both create and update requests.
type AccountAttributes struct {
	Name     string                 `json:"name,omitempty"`
	Settings map[string]interface{} `json:"settings,omitempty"`
	Secrets  map[string]interface{} `json:"secrets,omitempty"`
}

// accountRequest is the JSONAPI write envelope sent on POST and PATCH.
type accountRequest struct {
	Data struct {
		Type       string            `json:"type"`
		Attributes AccountAttributes `json:"attributes"`
	} `json:"data"`
}

// AccountResponseAttributes holds the fields returned by the API.
// IMPORTANT: Secrets are never included in any response.
type AccountResponseAttributes struct {
	Name     string                 `json:"name"`
	Settings map[string]interface{} `json:"settings"`
}

// AccountData is a single account resource in JSONAPI format.
type AccountData struct {
	ID         string                    `json:"id"`
	Type       string                    `json:"type"`
	Attributes AccountResponseAttributes `json:"attributes"`
}

// AccountResponse is the single-resource JSONAPI response (GET, POST, PATCH).
type AccountResponse struct {
	Data AccountData `json:"data"`
}

// Client calls the AMS Web Integrations public REST API.
type Client struct {
	client *datadog.APIClient
	auth   context.Context
}

func New(auth context.Context, client *datadog.APIClient) *Client {
	return &Client{client: client, auth: auth}
}

func (c *Client) accountsPath(integration string) string {
	return fmt.Sprintf("/api/v2/web-integrations/%s/accounts", integration)
}

func (c *Client) GetAccount(ctx context.Context, integration, accountID string) (*AccountResponse, *http.Response, error) {
	// Adding /{id} at the end of the endpoint path to get the account by ID
	path := fmt.Sprintf("%s/%s", c.accountsPath(integration), accountID)
	body, httpResp, err := utils.SendRequest(c.auth, c.client, "GET", path, nil)
	if err != nil {
		return nil, httpResp, err
	}
	var result AccountResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, httpResp, fmt.Errorf("parsing GET response: %w", err)
	}
	return &result, httpResp, nil
}

func (c *Client) CreateAccount(ctx context.Context, integration string, attrs AccountAttributes) (*AccountResponse, *http.Response, error) {
	var req accountRequest
	req.Data.Type = "Account"
	req.Data.Attributes = attrs

	body, httpResp, err := utils.SendRequest(c.auth, c.client, "POST", c.accountsPath(integration), req)
	if err != nil {
		return nil, httpResp, err
	}
	var result AccountResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, httpResp, fmt.Errorf("parsing POST response: %w", err)
	}
	return &result, httpResp, nil
}

func (c *Client) UpdateAccount(ctx context.Context, integration, accountID string, attrs AccountAttributes) (*AccountResponse, *http.Response, error) {
	var req accountRequest
	req.Data.Type = "Account"
	req.Data.Attributes = attrs

	path := fmt.Sprintf("%s/%s", c.accountsPath(integration), accountID)
	body, httpResp, err := utils.SendRequest(c.auth, c.client, "PATCH", path, req)
	if err != nil {
		return nil, httpResp, err
	}
	var result AccountResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, httpResp, fmt.Errorf("parsing PATCH response: %w", err)
	}
	return &result, httpResp, nil
}

func (c *Client) DeleteAccount(ctx context.Context, integration, accountID string) (*http.Response, error) {
	path := fmt.Sprintf("%s/%s", c.accountsPath(integration), accountID)
	_, httpResp, err := utils.SendRequest(c.auth, c.client, "DELETE", path, nil)
	return httpResp, err
}
