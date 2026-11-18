package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/emit"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ownersFileName is the tfgen-owned record of each generated data source's
// annotation-owner tag (the x-datadog-tf-generator tag that declared it). An
// orphan's annotation is gone from the spec, so a scoped --reconcile reads its
// owner from here. It sits beside datasources_generated.go and is written on
// every generate run, independently of --emit-tests.
const ownersFileName = "datasources_generated_owners.json"

func ownersPath(outputRoot string) string {
	return filepath.Join(outputRoot, ownersFileName)
}

// readOwners loads the artifact name → owner tag map; a missing file is empty.
func readOwners(outputRoot string) (map[string]string, error) {
	data, err := os.ReadFile(ownersPath(outputRoot))
	if errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	owners := map[string]string{}
	if err := json.Unmarshal(data, &owners); err != nil {
		return nil, fmt.Errorf("generate: %s: %w", ownersPath(outputRoot), err)
	}
	return owners, nil
}

func writeOwners(outputRoot string, owners map[string]string, check bool) (model.ArtifactStatus, error) {
	data, err := json.MarshalIndent(owners, "", "  ") // map keys are sorted
	if err != nil {
		return model.ArtifactStatusFailed, err
	}
	return emit.WriteFile(ownersPath(outputRoot), append(data, '\n'), check)
}

// syncOwners merges this run's owners into the file, leaving other entries
// intact. A run that generated no data source leaves the file untouched.
func syncOwners(outputRoot string, add map[string]string, check bool) (bool, error) {
	if len(add) == 0 {
		return false, nil
	}
	owners, err := readOwners(outputRoot)
	if err != nil {
		return false, err
	}
	for name, tag := range add {
		owners[name] = tag
	}
	status, err := writeOwners(outputRoot, owners, check)
	return wouldChange(status), err
}

// removeOwner drops name's entry; an absent entry or file changes nothing.
func removeOwner(outputRoot, name string, check bool) error {
	owners, err := readOwners(outputRoot)
	if err != nil {
		return err
	}
	if _, ok := owners[name]; !ok {
		return nil
	}
	delete(owners, name)
	_, err = writeOwners(outputRoot, owners, check)
	return err
}
