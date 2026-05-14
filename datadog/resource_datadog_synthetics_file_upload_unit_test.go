package datadog

import (
	"encoding/base64"
	"strings"
	"testing"

	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildDatadogBodyFiles_EncodingAutoSet validates that buildDatadogBodyFiles
// automatically sets encoding to "base64" when content is provided without an
// explicit encoding. Without this, the worker sends base64 text to S3 instead
// of the decoded binary.
func TestBuildDatadogBodyFiles_EncodingAutoSet(t *testing.T) {
	// Generate a small dummy binary payload and base64-encode it so the test
	// doesn't require any real file fixture in the repo.
	rawBytes := []byte("dummy binary content \x00\x01\x02\x03")
	b64Content := base64.StdEncoding.EncodeToString(rawBytes)

	tests := []struct {
		name             string
		encoding         string
		expectedEncoding string
	}{
		{
			name:             "no encoding set — should default to base64",
			encoding:         "",
			expectedEncoding: "base64",
		},
		{
			name:             "explicit encoding preserved",
			encoding:         "utf-8",
			expectedEncoding: "utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := []interface{}{
				map[string]interface{}{
					"name":               "file.bin",
					"original_file_name": "file.bin",
					"type":               "application/octet-stream",
					"size":               len(rawBytes),
					"content":            b64Content,
					"encoding":           tt.encoding,
					"bucket_key":         "",
				},
			}

			files := buildDatadogBodyFiles(attr)
			require.Len(t, files, 1)
			assert.Equal(t, tt.expectedEncoding, files[0].GetEncoding())
		})
	}
}

// TestBuildDatadogBodyFiles_NoContent validates that encoding is NOT set when no
// content is provided (e.g., when the file was already uploaded and the bucket
// key is being reused).
func TestBuildDatadogBodyFiles_NoContent(t *testing.T) {
	attr := []interface{}{
		map[string]interface{}{
			"name":               "file.bin",
			"original_file_name": "file.bin",
			"type":               "application/octet-stream",
			"size":               42,
			"content":            "",
			"encoding":           "",
			"bucket_key":         "some-bucket-key",
		},
	}

	files := buildDatadogBodyFiles(attr)
	require.Len(t, files, 1)
	// Encoding must not be set — the HasEncoding helper returns false when unset.
	assert.False(t, files[0].HasEncoding())
}

// TestSyntheticsRequestFileContentValidation verifies the content field's
// ValidateFunc boundary: the upper bound is ceil(3,145,728 * 4/3) = 4,194,304
// bytes — the maximum base64 string that decodes to ≤3 MB. The test generates
// strings at the boundary in-memory; no real file fixtures are committed.
func TestSyntheticsRequestFileContentValidation(t *testing.T) {
	fileSchema := syntheticsTestRequestFile()
	contentSchema := fileSchema.Elem.(*sdkschema.Resource).Schema["content"]
	validateFunc := contentSchema.ValidateFunc

	maxLen := 4194304

	tests := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name:      "empty string — below minimum",
			content:   "",
			wantError: true,
		},
		{
			name:      "single byte — at minimum",
			content:   "A",
			wantError: false,
		},
		{
			name:      "exactly at max length",
			content:   strings.Repeat("A", maxLen),
			wantError: false,
		},
		{
			name:      "one byte over max length",
			content:   strings.Repeat("A", maxLen+1),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateFunc(tt.content, "content")
			if tt.wantError {
				assert.NotEmpty(t, errs, "expected validation error but got none")
			} else {
				assert.Empty(t, errs, "expected no validation error but got: %v", errs)
			}
		})
	}
}
