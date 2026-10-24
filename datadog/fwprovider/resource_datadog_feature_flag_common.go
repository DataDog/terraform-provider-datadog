package fwprovider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	errFeatureFlagNotFound            = errors.New("feature flag not found")
	errFeatureFlagEnvironmentNotFound = errors.New("feature flag environment not found")
)

func featureFlagResponseError(err error, httpResponse *http.Response, operation string) error {
	if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
		return fmt.Errorf("%w: %v", errFeatureFlagNotFound, err)
	}
	return utils.TranslateClientError(err, httpResponse, operation)
}

func isFeatureFlagHTTPNotFound(err error) bool {
	return errors.Is(err, errFeatureFlagNotFound)
}

func parseFeatureFlagUUID(value, fieldName string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s must be a valid UUID: %w", fieldName, err)
	}
	return id, nil
}

func featureFlagCompositeID(featureFlagID, environmentID uuid.UUID) string {
	return featureFlagID.String() + ":" + environmentID.String()
}

func parseFeatureFlagCompositeID(value string) (uuid.UUID, uuid.UUID, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return uuid.Nil, uuid.Nil, fmt.Errorf("expected <feature_flag_id>:<environment_id>, got %q", value)
	}
	featureFlagID, err := parseFeatureFlagUUID(parts[0], "feature_flag_id")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	environmentID, err := parseFeatureFlagUUID(parts[1], "environment_id")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return featureFlagID, environmentID, nil
}

func importFeatureFlagCompositeState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	featureFlagID, environmentID, err := parseFeatureFlagCompositeID(request.ID)
	if err != nil {
		response.Diagnostics.AddError("Invalid feature flag import ID", err.Error())
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, frameworkPath.Root("id"), request.ID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, frameworkPath.Root("feature_flag_id"), featureFlagID.String())...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, frameworkPath.Root("environment_id"), environmentID.String())...)
}

// checkFeatureFlagResponse ignores the known generated-client mismatch for
// feature_flag_environments[].allocations while still rejecting every other
// unparsed part of the response.
func checkFeatureFlagResponse(response *datadogV2.FeatureFlagResponse) error {
	if response == nil {
		return nil
	}
	copyForCheck := *response
	copyForCheck.Data.Attributes.FeatureFlagEnvironments = nil
	return utils.CheckForUnparsed(copyForCheck)
}

// featureFlagEnvironmentToRaw returns the original JSON object when the
// generated client could not decode allocations (declared as a map by the
// client but returned as an array by the API). Typed entries are round-tripped
// so both shapes have one read path.
func featureFlagEnvironmentToRaw(environment datadogV2.FeatureFlagEnvironment) (map[string]interface{}, error) {
	if environment.UnparsedObject != nil {
		return environment.UnparsedObject, nil
	}
	serialized, err := json.Marshal(environment)
	if err != nil {
		return nil, fmt.Errorf("marshal feature flag environment: %w", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(serialized, &raw); err != nil {
		return nil, fmt.Errorf("decode feature flag environment: %w", err)
	}
	return raw, nil
}

func findFeatureFlagEnvironment(response *datadogV2.FeatureFlagResponse, environmentID uuid.UUID) (map[string]interface{}, error) {
	if response == nil {
		return nil, errFeatureFlagEnvironmentNotFound
	}
	for _, environment := range response.Data.Attributes.FeatureFlagEnvironments {
		raw, err := featureFlagEnvironmentToRaw(environment)
		if err != nil {
			return nil, err
		}
		rawID, _ := raw["environment_id"].(string)
		if rawID == environmentID.String() {
			return raw, nil
		}
	}
	return nil, errFeatureFlagEnvironmentNotFound
}

func featureFlagVariantKeyToID(response *datadogV2.FeatureFlagResponse) map[string]uuid.UUID {
	variants := make(map[string]uuid.UUID)
	if response == nil {
		return variants
	}
	for _, variant := range response.Data.Attributes.Variants {
		variants[variant.Key] = variant.Id
	}
	return variants
}

func featureFlagVariantIDToKey(response *datadogV2.FeatureFlagResponse) map[uuid.UUID]string {
	variants := make(map[uuid.UUID]string)
	if response == nil {
		return variants
	}
	for _, variant := range response.Data.Attributes.Variants {
		variants[variant.Id] = variant.Key
	}
	return variants
}

// defaultFeatureFlagVariantKey derives the create-only default when the API
// exposes it consistently through per-environment default variant IDs. An
// explicitly configured value is otherwise preserved in Terraform state.
func defaultFeatureFlagVariantKey(response *datadogV2.FeatureFlagResponse) types.String {
	variantKeys := featureFlagVariantIDToKey(response)
	var defaultKey string
	for _, environment := range response.Data.Attributes.FeatureFlagEnvironments {
		raw, err := featureFlagEnvironmentToRaw(environment)
		if err != nil {
			continue
		}
		defaultVariantID, _ := raw["default_variant_id"].(string)
		if defaultVariantID == "" {
			continue
		}
		id, err := uuid.Parse(defaultVariantID)
		if err != nil {
			continue
		}
		key := variantKeys[id]
		if key == "" || (defaultKey != "" && defaultKey != key) {
			return types.StringNull()
		}
		defaultKey = key
	}
	if defaultKey == "" {
		return types.StringNull()
	}
	return types.StringValue(defaultKey)
}

func featureFlagTimeValue(value *time.Time) types.String {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(value.Format(time.RFC3339Nano))
}
