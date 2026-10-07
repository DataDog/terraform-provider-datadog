package fwprovider

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

type fakeMonitorAutomationAPI struct {
	get func(context.Context, int64) (datadogV2.MonitorAutomationResponse, *http.Response, error)
	put func(context.Context, int64, datadogV2.MonitorAutomationRequest) (datadogV2.MonitorAutomationResponse, *http.Response, error)
}

func (f fakeMonitorAutomationAPI) GetMonitorAutomation(ctx context.Context, id int64) (datadogV2.MonitorAutomationResponse, *http.Response, error) {
	return f.get(ctx, id)
}
func (f fakeMonitorAutomationAPI) UpdateMonitorAutomation(ctx context.Context, id int64, body datadogV2.MonitorAutomationRequest) (datadogV2.MonitorAutomationResponse, *http.Response, error) {
	return f.put(ctx, id, body)
}

func automationResponse(enabled bool) datadogV2.MonitorAutomationResponse {
	return datadogV2.MonitorAutomationResponse{Data: datadogV2.MonitorAutomationData{
		Id: "123", Type: datadogV2.MONITORAUTOMATIONTYPE_MONITOR_AUTOMATION,
		Attributes: datadogV2.MonitorAutomationAttributes{Enabled: enabled},
	}}
}

func TestBitsAIMonitorAutomationApply(t *testing.T) {
	for _, desired := range []bool{false, true} {
		t.Run(types.BoolValue(desired).String(), func(t *testing.T) {
			writes, reads := 0, 0
			r := &bitsAIMonitorAutomationResource{auth: context.Background(), api: fakeMonitorAutomationAPI{
				put: func(_ context.Context, id int64, body datadogV2.MonitorAutomationRequest) (datadogV2.MonitorAutomationResponse, *http.Response, error) {
					writes++
					require.Equal(t, int64(123), id)
					require.Equal(t, desired, body.Data.Attributes.Enabled)
					return automationResponse(desired), &http.Response{StatusCode: 200}, nil
				},
				get: func(_ context.Context, id int64) (datadogV2.MonitorAutomationResponse, *http.Response, error) {
					reads++
					require.Equal(t, int64(123), id)
					return automationResponse(desired), &http.Response{StatusCode: 200}, nil
				},
			}}
			plan := bitsAIMonitorAutomationModel{MonitorID: types.StringValue("123"), Enabled: types.BoolValue(desired)}
			applied, diags := r.apply(context.Background(), &plan)
			require.True(t, applied)
			require.False(t, diags.HasError(), diags)
			require.Equal(t, "123", plan.ID.ValueString())
			require.Equal(t, 1, writes)
			require.Equal(t, 1, reads)
		})
	}
}

func TestBitsAIMonitorAutomationPreservesIDAfterWrite(t *testing.T) {
	r := &bitsAIMonitorAutomationResource{auth: context.Background(), api: fakeMonitorAutomationAPI{
		put: func(context.Context, int64, datadogV2.MonitorAutomationRequest) (datadogV2.MonitorAutomationResponse, *http.Response, error) {
			return automationResponse(true), nil, nil
		},
		get: func(context.Context, int64) (datadogV2.MonitorAutomationResponse, *http.Response, error) {
			return datadogV2.MonitorAutomationResponse{}, nil, errors.New("read failed")
		},
	}}
	plan := bitsAIMonitorAutomationModel{MonitorID: types.StringValue("123"), Enabled: types.BoolValue(true)}
	applied, diags := r.apply(context.Background(), &plan)
	require.True(t, applied)
	require.True(t, diags.HasError())
	require.Equal(t, "123", plan.ID.ValueString())
}

func TestBitsAIMonitorAutomationWaitsForIndex(t *testing.T) {
	writes, reads := 0, 0
	r := &bitsAIMonitorAutomationResource{auth: context.Background(), api: fakeMonitorAutomationAPI{
		put: func(context.Context, int64, datadogV2.MonitorAutomationRequest) (datadogV2.MonitorAutomationResponse, *http.Response, error) {
			writes++
			if writes == 1 {
				return datadogV2.MonitorAutomationResponse{}, &http.Response{StatusCode: 404}, errors.New("new monitor not indexed yet")
			}
			return automationResponse(true), nil, nil
		},
		get: func(context.Context, int64) (datadogV2.MonitorAutomationResponse, *http.Response, error) {
			reads++
			return automationResponse(reads > 1), nil, nil
		},
	}}
	plan := bitsAIMonitorAutomationModel{MonitorID: types.StringValue("123"), Enabled: types.BoolValue(true)}
	applied, diags := r.apply(context.Background(), &plan)
	require.True(t, applied)
	require.False(t, diags.HasError(), diags)
	require.Equal(t, 2, writes)
	require.Equal(t, 2, reads)
}

func TestBitsAIMonitorAutomationDoesNotRetryPermissionErrors(t *testing.T) {
	calls := 0
	r := &bitsAIMonitorAutomationResource{auth: context.Background(), api: fakeMonitorAutomationAPI{
		put: func(context.Context, int64, datadogV2.MonitorAutomationRequest) (datadogV2.MonitorAutomationResponse, *http.Response, error) {
			calls++
			return datadogV2.MonitorAutomationResponse{}, &http.Response{StatusCode: 403}, errors.New("forbidden")
		},
	}}
	plan := bitsAIMonitorAutomationModel{MonitorID: types.StringValue("123"), Enabled: types.BoolValue(true)}
	applied, diags := r.apply(context.Background(), &plan)
	require.False(t, applied)
	require.True(t, diags.HasError())
	require.Equal(t, 1, calls)
}

func TestBitsAIMonitorAutomationWaitCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, waitMonitorAutomation(ctx, func() (bool, error) { return false, nil }), context.Canceled)
}

func TestBitsAIMonitorAutomationID(t *testing.T) {
	for _, value := range []string{"", "0", "-1", "abc", "9223372036854775808"} {
		_, err := monitorAutomationID(value)
		require.Error(t, err)
	}
	id, err := monitorAutomationID("123")
	require.NoError(t, err)
	require.Equal(t, int64(123), id)
}
