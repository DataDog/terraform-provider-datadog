package utils

import (
	"bytes"
	"encoding/json"
	"github.com/hashicorp/go-cleanhttp"
	"net/http"
	"os"
)

type HttpClient struct {
	apiKey, appKey, baseUrl string
	Client                  *http.Client
	extraHeaders            map[string]string
}

func NewHttpClient(apiKey, appKey string) *HttpClient {
	client := cleanhttp.DefaultClient()
	baseUrl := os.Getenv("DATADOG_HOST")
	if baseUrl == "" {
		baseUrl = "https://api.datadoghq.com"
	}

	return &HttpClient{
		apiKey:       apiKey,
		appKey:       appKey,
		baseUrl:      baseUrl,
		extraHeaders: map[string]string{},
		Client:       client,
	}
}

func (c *HttpClient) SetUrl(url string) {
	c.baseUrl = url
}

func (c *HttpClient) SetExtraHeaders(headers map[string]string) {
	c.extraHeaders = headers
}

func (c *HttpClient) SendRequest(method, path string, body map[string]interface{}) (map[string]interface{}, error) {
	// Build request body
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}

	// Build request
	req, _ := http.NewRequest(method, c.baseUrl+path, &buf)

	// Set request headers
	for k, v := range c.extraHeaders {
		req.Header.Set(k, v)
	}
	req.Header.Set("DD-API-KEY", c.apiKey)
	req.Header.Set("DD-APPLICATION-KEY", c.appKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.Client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&result)

	return result, nil
}
