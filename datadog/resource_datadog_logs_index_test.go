package datadog

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceDatadogLogsIndexDelete_PreventDeletion(t *testing.T) {
	d := schema.TestResourceDataRaw(t, indexSchema, map[string]interface{}{
		"prevent_deletion": true,
	})
	d.SetId("main")

	diags := resourceDatadogLogsIndexDelete(context.Background(), d, nil)

	require.Len(t, diags, 1)
	assert.Contains(t, diags[0].Summary, "Deletion of logs index 'main' is prevented by 'prevent_deletion' flag")
}

func TestResourceDatadogLogsIndexDelete_AllowDeletion(t *testing.T) {
	// When prevent_deletion is false the guard is skipped and the function proceeds
	// to the API call. Since meta is nil it will panic before reaching the API —
	// we just confirm no prevent_deletion diagnostic is returned before the panic.
	d := schema.TestResourceDataRaw(t, indexSchema, map[string]interface{}{
		"prevent_deletion": false,
	})
	d.SetId("other-index")

	defer func() { recover() }() // absorb the nil-meta panic
	diags := resourceDatadogLogsIndexDelete(context.Background(), d, nil)

	for _, diag := range diags {
		assert.NotContains(t, diag.Summary, "prevent_deletion")
	}
}
