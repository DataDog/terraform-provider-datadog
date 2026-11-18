package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const statusPageNotificationSettingsTestID = "00000000-0000-4000-8000-000000000201"

// statusPageNotificationSettingsTestData returns a status page response body. A nil
// setting is omitted from the response, as in responses recorded before the
// API exposed it.
func statusPageNotificationSettingsTestData(emailNotifications, msTeamsSubscriptions *bool) string {
	settings := ""
	if emailNotifications != nil {
		settings += fmt.Sprintf(`,"email_notifications_enabled":%t`, *emailNotifications)
	}
	if msTeamsSubscriptions != nil {
		settings += fmt.Sprintf(`,"ms_teams_subscriptions_enabled":%t`, *msTeamsSubscriptions)
	}
	return fmt.Sprintf(`{"data":{"id":%q,"type":"status_pages","attributes":{"name":"Example Status","type":"private","domain_prefix":"example-status","visualization_type":"bars_only"%s}}}`,
		statusPageNotificationSettingsTestID, settings)
}

func TestStatusPageNotificationSettingsCreateUpdateRead(t *testing.T) {
	t.Parallel()
	enabled, disabled := true, false
	for _, test := range []struct {
		name                 string
		emailNotifications   types.Bool
		msTeamsSubscriptions types.Bool
		// API response values; nil omits the setting from the response.
		responseEmail   *bool
		responseMsTeams *bool
	}{
		{"enabled", types.BoolValue(true), types.BoolValue(true), &enabled, &enabled},
		{"disabled", types.BoolValue(false), types.BoolValue(false), &disabled, &disabled},
		{"mixed", types.BoolValue(true), types.BoolValue(false), &enabled, &disabled},
		{"omitted", types.BoolNull(), types.BoolNull(), &enabled, &disabled},
		{"unknown", types.BoolUnknown(), types.BoolUnknown(), &enabled, &disabled},
		{"absent from response", types.BoolUnknown(), types.BoolUnknown(), nil, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				if request.Method == http.MethodPost {
					assert.Equal(t, "/api/v2/statuspages", request.URL.Path)
				} else {
					assert.Equal(t, "/api/v2/statuspages/"+statusPageNotificationSettingsTestID, request.URL.Path)
				}
				if request.Method == http.MethodPost || request.Method == http.MethodPatch {
					var body struct {
						Data struct {
							Attributes map[string]interface{} `json:"attributes"`
						} `json:"data"`
					}
					assert.NoError(t, json.NewDecoder(request.Body).Decode(&body))
					for attribute, value := range map[string]types.Bool{
						"email_notifications_enabled":    test.emailNotifications,
						"ms_teams_subscriptions_enabled": test.msTeamsSubscriptions,
					} {
						if value.IsNull() || value.IsUnknown() {
							assert.NotContains(t, body.Data.Attributes, attribute)
						} else {
							assert.Equal(t, value.ValueBool(), body.Data.Attributes[attribute], attribute)
						}
					}
				}
				w.Header().Set("Content-Type", "application/json")
				if request.Method == http.MethodPost {
					w.WriteHeader(http.StatusCreated)
				}
				fmt.Fprint(w, statusPageNotificationSettingsTestData(test.responseEmail, test.responseMsTeams))
			}))
			defer server.Close()

			wantEmail, wantMsTeams := types.BoolNull(), types.BoolNull()
			if test.responseEmail != nil {
				wantEmail = types.BoolValue(*test.responseEmail)
			}
			if test.responseMsTeams != nil {
				wantMsTeams = types.BoolValue(*test.responseMsTeams)
			}
			assertSettings := func(state tfsdk.State) {
				t.Helper()
				var got statusPageModel
				require.False(t, state.Get(ctx, &got).HasError())
				require.Equal(t, wantEmail, got.EmailNotificationsEnabled)
				require.Equal(t, wantMsTeams, got.MsTeamsSubscriptionsEnabled)
			}

			r := &statusPageResource{Api: newStatusPagesAPIForTest(server), Auth: ctx}
			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			model := statusPageModel{
				ID:                          types.StringUnknown(),
				Name:                        types.StringValue("Example Status"),
				Type:                        types.StringValue("private"),
				DomainPrefix:                types.StringValue("example-status"),
				VisualizationType:           types.StringValue("bars_only"),
				EmailNotificationsEnabled:   test.emailNotifications,
				MsTeamsSubscriptionsEnabled: test.msTeamsSubscriptions,
			}
			plan := tfsdk.Plan{Schema: schemaResponse.Schema}
			require.False(t, plan.Set(ctx, model).HasError())

			created := resource.CreateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
			r.Create(ctx, resource.CreateRequest{Plan: plan}, &created)
			require.False(t, created.Diagnostics.HasError(), "%v", created.Diagnostics)
			assertSettings(created.State)

			model.ID = types.StringValue(statusPageNotificationSettingsTestID)
			require.False(t, plan.Set(ctx, model).HasError())
			updated := resource.UpdateResponse{State: created.State}
			r.Update(ctx, resource.UpdateRequest{Plan: plan, State: created.State}, &updated)
			require.False(t, updated.Diagnostics.HasError(), "%v", updated.Diagnostics)
			assertSettings(updated.State)

			read := resource.ReadResponse{State: updated.State}
			r.Read(ctx, resource.ReadRequest{State: updated.State}, &read)
			require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
			assertSettings(read.State)
		})
	}
}

func TestStatusPageNotificationSettingsDataSource(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	enabled, disabled := true, false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "/api/v2/statuspages/"+statusPageNotificationSettingsTestID, request.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, statusPageNotificationSettingsTestData(&enabled, &disabled))
	}))
	defer server.Close()

	d := &statusPageDataSource{Api: newStatusPagesAPIForTest(server), Auth: ctx}
	var schemaResponse datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schemaResponse)
	objectType := schemaResponse.Schema.Type().TerraformType(ctx).(tftypes.Object)
	values := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
	for name, attributeType := range objectType.AttributeTypes {
		values[name] = tftypes.NewValue(attributeType, nil)
	}
	values["id"] = tftypes.NewValue(tftypes.String, statusPageNotificationSettingsTestID)
	config := tfsdk.Config{Raw: tftypes.NewValue(objectType, values), Schema: schemaResponse.Schema}

	response := datasource.ReadResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	d.Read(ctx, datasource.ReadRequest{Config: config}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)

	var got statusPageDataSourceModel
	require.False(t, response.State.Get(ctx, &got).HasError())
	require.Equal(t, types.StringValue("private"), got.Type)
	require.Equal(t, types.BoolValue(true), got.EmailNotificationsEnabled)
	require.Equal(t, types.BoolValue(false), got.MsTeamsSubscriptionsEnabled)
}
