package validators

import "testing"

func TestResourceDatadogDowntimeRecurrenceTypeValidation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "daily",
			ErrCount: 1,
		},
		{
			Value:    "days",
			ErrCount: 0,
		},
		{
			Value:    "days,weeks",
			ErrCount: 1,
		},
		{
			Value:    "months",
			ErrCount: 0,
		},
		{
			Value:    "years",
			ErrCount: 0,
		},
		{
			Value:    "weeks",
			ErrCount: 0,
		},
	}

	for _, tc := range cases {
		_, errors := ValidateDatadogDowntimeRecurrenceType(tc.Value, "datadog_downtime_recurrence_type")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected Datadog Downtime Recurrence Type validation to trigger %d error(s) for value %q - instead saw %d",
				tc.ErrCount, tc.Value, len(errors))
		}
	}
}

func TestResourceDatadogDowntimeRecurrenceWeekDaysValidation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "Mon",
			ErrCount: 0,
		},
		{
			Value:    "Mon,",
			ErrCount: 1,
		},
		{
			Value:    "Monday",
			ErrCount: 1,
		},
		{
			Value:    "mon",
			ErrCount: 1,
		},
		{
			Value:    "mon,",
			ErrCount: 1,
		},
		{
			Value:    "monday",
			ErrCount: 1,
		},
		{
			Value:    "mon,Tue",
			ErrCount: 1,
		},
		{
			Value:    "Mon,tue",
			ErrCount: 1,
		},
		{
			Value:    "Mon,Tue",
			ErrCount: 1,
		},
		{
			Value:    "Mon, Tue",
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := ValidateDatadogDowntimeRecurrenceWeekDays(tc.Value, "datadog_downtime_recurrence_week_days")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected Datadog Downtime Recurrence Week Days validation to trigger %d error(s) for value %q - instead saw %d",
				tc.ErrCount, tc.Value, len(errors))
		}
	}
}
