package observability_pipeline

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestExpandAzureDataExplorerDestinationIncludesIngestionMode(t *testing.T) {
	ctx := context.Background()
	inputs, diags := types.ListValueFrom(ctx, types.StringType, []string{"source-1"})
	require.False(t, diags.HasError())

	destination := ExpandAzureDataExplorerDestination(ctx, "destination-1", inputs, &AzureDataExplorerDestinationModel{
		Database: types.StringValue("logs_db"),
		Table:    types.StringValue("application_logs"),
		Auth: []AzureDataExplorerAuthModel{{
			ManagedIdentity: []AzureDataExplorerAuthManagedIdentityModel{{}},
		}},
	})

	payload, err := json.Marshal(destination)
	require.NoError(t, err)
	var fields map[string]interface{}
	require.NoError(t, json.Unmarshal(payload, &fields))
	// The live API requires this field even though Terraform does not expose it.
	// Omitting it can pass older recordings while causing creates to fail with 400.
	require.Equal(t, "streaming", fields["ingestion_mode"])
}
