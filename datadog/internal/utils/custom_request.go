package utils

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
)

// SendRequest send custom request
func SendRequest(ctx context.Context, client *datadogV1.APIClient, method, path string, body interface{}) ([]byte, *http.Response, error) {
	req, err := buildRequest(ctx, client, method, path, body)
	if err != nil {
		return nil, nil, err
	}

	httpRes, err := client.CallAPI(req)
	if err != nil {
		return nil, nil, err
	}

	var bodyResByte []byte
	bodyResByte, err = ioutil.ReadAll(httpRes.Body)
	defer httpRes.Body.Close()
	if err != nil {
		return nil, httpRes, err
	}

	if httpRes.StatusCode >= 300 {
		newErr := CustomRequestAPIError{
			body:  bodyResByte,
			error: httpRes.Status,
		}
		return nil, httpRes, newErr
	}

	return bodyResByte, httpRes, nil
}

func buildRequest(ctx context.Context, client *datadogV1.APIClient, method, path string, body interface{}) (*http.Request, error) {
	var (
		localVarPostBody        interface{}
		localVarPath            string
		localVarQueryParams     url.Values
		localVarFormQueryParams url.Values
		localVarFormFile        *datadogV1.FormFile
	)

	localBasePath, err := client.GetConfig().ServerURLWithContext(ctx, "")
	if err != nil {
		return nil, err
	}
	localVarPath = localBasePath + path

	localVarHeaderParams := make(map[string]string)
	localVarHeaderParams["Content-Type"] = "application/json"

	localVarHTTPHeaderAccepts := make(map[string]string)
	localVarHTTPHeaderAccepts["Accept"] = "application/json"

	if body != nil {
		localVarPostBody = body
	}

	if ctx != nil {
		if auth, ok := ctx.Value(datadogV1.ContextAPIKeys).(map[string]datadogV1.APIKey); ok {
			if apiKey, ok := auth["apiKeyAuth"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["DD-API-KEY"] = key
			}
		}
	}
	if ctx != nil {
		if auth, ok := ctx.Value(datadogV1.ContextAPIKeys).(map[string]datadogV1.APIKey); ok {
			if apiKey, ok := auth["appKeyAuth"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["DD-APPLICATION-KEY"] = key
			}
		}
	}

	req, err := client.PrepareRequest(ctx, localVarPath, method, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormQueryParams, localVarFormFile)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// CustomRequestAPIError Provides access to the body, and error on returned errors.
type CustomRequestAPIError struct {
	body  []byte
	error string
}

// Error returns non-empty string if there was an error.
func (e CustomRequestAPIError) Error() string {
	return e.error
}

// Body returns the raw bytes of the response
func (e CustomRequestAPIError) Body() []byte {
	return e.body
}
