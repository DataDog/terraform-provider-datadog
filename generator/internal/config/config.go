package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the top-level generator configuration.
type Config struct {
	Specs       map[string]SpecConfig       `yaml:"specs"`
	DataSources map[string]DataSourceConfig `yaml:"datasources"`
}

// SpecConfig references an OpenAPI spec file.
type SpecConfig struct {
	Path string `yaml:"path"`
}

// DataSourceConfig describes a data source to generate.
type DataSourceConfig struct {
	Read OperationRef  `yaml:"read"`
	List *OperationRef `yaml:"list,omitempty"` // optional, enables filter-fallback lookup
	Spec string        `yaml:"spec"`           // optional, defaults to first key in specs map
}

// OperationRef identifies an API operation by path and method.
type OperationRef struct {
	Path   string `yaml:"path"`
	Method string `yaml:"method"`
}

// LoadConfig reads and validates a YAML configuration file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validating config %s: %w", path, err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.Specs) == 0 {
		return fmt.Errorf("specs map must not be empty")
	}

	for name, ds := range c.DataSources {
		if ds.Read.Path == "" {
			return fmt.Errorf("datasource %q: read.path is required", name)
		}
		method := strings.ToLower(ds.Read.Method)
		if method == "" {
			return fmt.Errorf("datasource %q: read.method is required", name)
		}
		validMethods := map[string]bool{"get": true, "post": true, "put": true, "patch": true, "delete": true}
		if !validMethods[method] {
			return fmt.Errorf("datasource %q: read.method %q is not a valid HTTP method", name, ds.Read.Method)
		}

		// Validate list operation if present
		if ds.List != nil {
			if ds.List.Path == "" {
				return fmt.Errorf("datasource %q: list.path is required", name)
			}
			listMethod := strings.ToLower(ds.List.Method)
			if listMethod == "" {
				return fmt.Errorf("datasource %q: list.method is required", name)
			}
			if !validMethods[listMethod] {
				return fmt.Errorf("datasource %q: list.method %q is not a valid HTTP method", name, ds.List.Method)
			}
		}

		specRef := ds.Spec
		if specRef == "" {
			specRef = c.defaultSpec()
		}
		if _, ok := c.Specs[specRef]; !ok {
			return fmt.Errorf("datasource %q: spec %q not found in specs map", name, specRef)
		}
	}

	return nil
}

// defaultSpec returns the first spec key in alphabetical order.
func (c *Config) defaultSpec() string {
	keys := make([]string, 0, len(c.Specs))
	for k := range c.Specs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys[0]
}

// ResolveSpec returns the spec name for a data source, applying the default if needed.
func (c *Config) ResolveSpec(ds DataSourceConfig) string {
	if ds.Spec != "" {
		return ds.Spec
	}
	return c.defaultSpec()
}
