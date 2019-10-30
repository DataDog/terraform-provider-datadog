package datadog

import "fmt"

func validateAggregatorMethod(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	validMethods := map[string]struct{}{
		"avg":   {},
		"max":   {},
		"min":   {},
		"sum":   {},
		"last":  {},
		"count": {},
	}
	if _, ok := validMethods[value]; !ok {
		errors = append(errors, fmt.Errorf(
			`%q contains an invalid method %q. Valid methods are either "avg", "max", "min", "sum", "count", or "last"`, k, value))
	}
	return
}
