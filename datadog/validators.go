package datadog

import (
	"bytes"
	"encoding/json"
	"fmt"
)

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

func getMetadataFromJSON(jsonBytes []byte, unmarshalled interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(jsonBytes))
	// make sure we return errors on attributes that we don't expect in metadata
	decoder.DisallowUnknownFields()
	err := decoder.Decode(unmarshalled)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata_json: %s", err)
	}
	return nil
}
