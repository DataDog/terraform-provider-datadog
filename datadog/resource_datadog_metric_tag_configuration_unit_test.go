package datadog

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTagConfigurationRetryError(t *testing.T) {
	apiErr := errors.New("boom")

	tests := []struct {
		name          string
		err           error
		httpResponse  *http.Response
		wantNil       bool // classifier returned nil (success)
		wantRetryable bool // only meaningful when wantNil == false
	}{
		{
			name:    "success (nil error) is terminal",
			err:     nil,
			wantNil: true,
		},
		{
			name:          "409 conflict is retryable",
			err:           apiErr,
			httpResponse:  &http.Response{StatusCode: http.StatusConflict},
			wantRetryable: true,
		},
		{
			name:          "400 bad request is not retryable",
			err:           apiErr,
			httpResponse:  &http.Response{StatusCode: http.StatusBadRequest},
			wantRetryable: false,
		},
		{
			name:          "500 is not retryable here",
			err:           apiErr,
			httpResponse:  &http.Response{StatusCode: http.StatusInternalServerError},
			wantRetryable: false,
		},
		{
			name:          "nil http response (network error) is not retryable",
			err:           apiErr,
			httpResponse:  nil,
			wantRetryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createTagConfigurationRetryError("test.metric.name", tt.err, tt.httpResponse)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}
			assert.NotNil(t, got)
			assert.Equal(t, tt.wantRetryable, got.Retryable)
		})
	}
}
