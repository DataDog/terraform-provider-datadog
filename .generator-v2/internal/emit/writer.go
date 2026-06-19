package emit

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// WriteFile writes content to path, skipping the write when on-disk content
// matches. In check mode it reports what would happen without touching disk.
// Parent directories are created as needed.
func WriteFile(path string, content []byte, check bool) (model.ArtifactStatus, error) {
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return model.ArtifactStatusFailed, err
	}

	if errors.Is(err, os.ErrNotExist) {
		if !check {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return model.ArtifactStatusFailed, err
			}
			if err := os.WriteFile(path, content, 0o644); err != nil {
				return model.ArtifactStatusFailed, err
			}
		}
		return model.ArtifactStatusCreated, nil
	}

	if bytes.Equal(existing, content) {
		return model.ArtifactStatusUnchanged, nil
	}

	if !check {
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return model.ArtifactStatusFailed, err
		}
	}
	return model.ArtifactStatusUpdated, nil
}
