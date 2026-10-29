package fwprovider

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryRumRetentionFilterOperation(t *testing.T) {
	attempts := 0

	err := retryRumRetentionFilterOperation(t.Context(), func() (*http.Response, error) {
		attempts++
		if attempts == 1 {
			return &http.Response{StatusCode: http.StatusConflict}, errors.New("ConcurrentModification: retry")
		}
		return nil, nil
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
