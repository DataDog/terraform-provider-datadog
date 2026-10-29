package fwprovider

import (
	"errors"
	"net/http"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRumRetentionFilterWithRetry(t *testing.T) {
	attempts := 0

	_, err := createRumRetentionFilterWithRetry(t.Context(), func() (datadogV2.RumRetentionFilterResponse, *http.Response, error) {
		attempts++
		if attempts == 1 {
			return datadogV2.RumRetentionFilterResponse{}, &http.Response{StatusCode: http.StatusConflict}, errors.New("ConcurrentModification: retry")
		}
		return datadogV2.RumRetentionFilterResponse{}, nil, nil
	})

	require.NoError(t, err)
	assert.Equal(t, 2, attempts)
}

func TestIsRumRetentionFilterConcurrentModification(t *testing.T) {
	tests := map[string]struct {
		err       error
		httpResp  *http.Response
		wantRetry bool
	}{
		"concurrent modification conflict": {
			err:       errors.New("ConcurrentModification: retry"),
			httpResp:  &http.Response{StatusCode: http.StatusConflict},
			wantRetry: true,
		},
		"other conflict": {
			err:       errors.New("duplicate retention filter"),
			httpResp:  &http.Response{StatusCode: http.StatusConflict},
			wantRetry: false,
		},
		"server error": {
			err:       errors.New("ConcurrentModification: retry"),
			httpResp:  &http.Response{StatusCode: http.StatusInternalServerError},
			wantRetry: false,
		},
		"missing response": {
			err:       errors.New("ConcurrentModification: retry"),
			wantRetry: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.wantRetry, isRumRetentionFilterConcurrentModification(test.err, test.httpResp))
		})
	}
}
