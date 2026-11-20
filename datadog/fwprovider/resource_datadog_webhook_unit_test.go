package fwprovider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebhookUpdateStateClearsRemovedFields guards against a regression of
// https://github.com/DataDog/terraform-provider-datadog/issues/4479.
//
// Read passes the prior state to updateState. When the API reports
// `custom_headers` or `payload` as null (or leaves it out), refresh must clear
// the attribute so the next plan shows the drift instead of keeping the stale
// value from state.
func TestWebhookUpdateStateClearsRemovedFields(t *testing.T) {
	cases := map[string]string{
		"explicit null": `{"name":"example","url":"https://example.com/hook","encode_as":"json","custom_headers":null,"payload":null}`,
		"omitted":       `{"name":"example","url":"https://example.com/hook","encode_as":"json"}`,
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			var resp datadogV1.WebhooksIntegration
			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			state := webhookModel{
				ID:            types.StringValue("example"),
				Name:          types.StringValue("example"),
				URL:           types.StringValue("https://example.com/hook"),
				Payload:       types.StringValue(`{"title":"$EVENT_TITLE"}`),
				CustomHeaders: types.StringValue(`{"Authorization":"Bearer example"}`),
				EncodeAs:      types.StringValue("json"),
			}

			(&webhookResource{}).updateState(context.Background(), &state, &resp)

			assert.True(t, state.CustomHeaders.IsNull(), "custom_headers should be null, got %s", state.CustomHeaders)
			assert.True(t, state.Payload.IsNull(), "payload should be null, got %s", state.Payload)
			assert.Equal(t, "json", state.EncodeAs.ValueString())
		})
	}
}

func TestWebhookUpdateStateKeepsReturnedFields(t *testing.T) {
	body := `{"name":"example","url":"https://example.com/hook","encode_as":"form","custom_headers":"{\"Authorization\":\"Bearer new\"}","payload":"{\"title\":\"$EVENT_TITLE\"}"}`
	var resp datadogV1.WebhooksIntegration
	require.NoError(t, json.Unmarshal([]byte(body), &resp))

	state := webhookModel{
		Payload:       types.StringUnknown(),
		CustomHeaders: types.StringValue(`{"Authorization":"Bearer old"}`),
		EncodeAs:      types.StringUnknown(),
	}

	(&webhookResource{}).updateState(context.Background(), &state, &resp)

	assert.Equal(t, "example", state.ID.ValueString())
	assert.Equal(t, "example", state.Name.ValueString())
	assert.Equal(t, "https://example.com/hook", state.URL.ValueString())
	assert.Equal(t, `{"Authorization":"Bearer new"}`, state.CustomHeaders.ValueString())
	assert.Equal(t, `{"title":"$EVENT_TITLE"}`, state.Payload.ValueString())
	assert.Equal(t, "form", state.EncodeAs.ValueString())
}
