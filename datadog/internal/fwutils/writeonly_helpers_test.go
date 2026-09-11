package fwutils

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func testWriteOnlySchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"api_key_wo": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				WriteOnly: true,
			},
			"api_key_wo_version": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func ptr(s string) *string { return &s }

func makeConfigValue(apiKey, apiKeyWo *string) tftypes.Value {
	apiKeyVal := tftypes.NewValue(tftypes.String, nil)
	if apiKey != nil {
		apiKeyVal = tftypes.NewValue(tftypes.String, *apiKey)
	}
	apiKeyWoVal := tftypes.NewValue(tftypes.String, nil)
	if apiKeyWo != nil {
		apiKeyWoVal = tftypes.NewValue(tftypes.String, *apiKeyWo)
	}
	return tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"api_key":            tftypes.String,
			"api_key_wo":         tftypes.String,
			"api_key_wo_version": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"api_key":            apiKeyVal,
		"api_key_wo":         apiKeyWoVal,
		"api_key_wo_version": tftypes.NewValue(tftypes.String, nil),
	})
}

func makeVersionValue(version *string) tftypes.Value {
	versionVal := tftypes.NewValue(tftypes.String, nil)
	if version != nil {
		versionVal = tftypes.NewValue(tftypes.String, *version)
	}
	return tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"api_key_wo_version": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"api_key_wo_version": versionVal,
	})
}

func TestCreateWriteOnlySecretAttributes(t *testing.T) {
	config := WriteOnlySecretConfig{
		OriginalAttr:         "api_key",
		WriteOnlyAttr:        "api_key_wo",
		TriggerAttr:          "api_key_wo_version",
		OriginalDescription:  "The API key for the account.",
		WriteOnlyDescription: "Write-only API key for the account.",
		TriggerDescription:   "Version for `api_key_wo` rotation.",
	}

	attrs := CreateWriteOnlySecretAttributes(config)

	if len(attrs) != 3 {
		t.Errorf("expected 3 attributes, got %d", len(attrs))
	}

	// Verify api_key properties
	apiKey := attrs["api_key"].(schema.StringAttribute)
	if !apiKey.Optional || !apiKey.Sensitive || apiKey.WriteOnly {
		t.Error("api_key: expected Optional=true, Sensitive=true, WriteOnly=false")
	}
	if len(apiKey.Validators) != 2 {
		t.Errorf("api_key should have 2 validators, got %d", len(apiKey.Validators))
	}

	// Verify api_key_wo properties
	apiKeyWo := attrs["api_key_wo"].(schema.StringAttribute)
	if !apiKeyWo.Optional || !apiKeyWo.Sensitive || !apiKeyWo.WriteOnly {
		t.Error("api_key_wo: expected Optional=true, Sensitive=true, WriteOnly=true")
	}
	if len(apiKeyWo.Validators) != 2 {
		t.Errorf("api_key_wo should have 2 validators, got %d", len(apiKeyWo.Validators))
	}

	// Verify api_key_wo_version properties
	version := attrs["api_key_wo_version"].(schema.StringAttribute)
	if !version.Optional || version.Sensitive {
		t.Error("api_key_wo_version: expected Optional=true, Sensitive=false")
	}
	if len(version.Validators) != 2 {
		t.Errorf("api_key_wo_version should have 2 validators, got %d", len(version.Validators))
	}
}

func TestGetSecretForCreate(t *testing.T) {
	tests := []struct {
		name          string
		apiKey        *string
		apiKeyWo      *string
		wantShouldSet bool
		wantValue     string
	}{
		{
			name:          "write-only mode",
			apiKeyWo:      ptr("secret123"),
			wantShouldSet: true,
			wantValue:     "secret123",
		},
		{
			name:          "plaintext mode",
			apiKey:        ptr("plaintext_secret"),
			wantShouldSet: true,
			wantValue:     "plaintext_secret",
		},
		{
			name:          "no secret",
			wantShouldSet: false,
		},
		{
			name:          "empty string is valid",
			apiKey:        ptr(""),
			wantShouldSet: true,
			wantValue:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			handler := WriteOnlySecretHandler{
				Config: WriteOnlySecretConfig{
					OriginalAttr:  "api_key",
					WriteOnlyAttr: "api_key_wo",
					TriggerAttr:   "api_key_wo_version",
				},
			}

			tfConfig := tfsdk.Config{
				Raw:    makeConfigValue(tt.apiKey, tt.apiKeyWo),
				Schema: testWriteOnlySchema(),
			}

			result := handler.GetSecretForCreate(ctx, &tfConfig)

			if result.Diagnostics.HasError() {
				t.Errorf("unexpected error: %v", result.Diagnostics)
			}
			if result.ShouldSetValue != tt.wantShouldSet {
				t.Errorf("ShouldSetValue = %v, want %v", result.ShouldSetValue, tt.wantShouldSet)
			}
			if tt.wantShouldSet && result.Value != tt.wantValue {
				t.Errorf("Value = %q, want %q", result.Value, tt.wantValue)
			}
		})
	}
}

func TestGetSecretForUpdate(t *testing.T) {
	tests := []struct {
		name                   string
		stateVersion           *string
		planVersion            *string
		apiKey                 *string
		apiKeyWo               *string
		apiKeyWoUnknown        bool
		secretRequiredOnUpdate bool
		wantShouldSet          bool
		wantValue              string
	}{
		{
			name:          "version changed triggers update",
			stateVersion:  ptr("1"),
			planVersion:   ptr("2"),
			apiKeyWo:      ptr("newsecret"),
			wantShouldSet: true,
			wantValue:     "newsecret",
		},
		{
			name:          "version unchanged skips update (partial update)",
			stateVersion:  ptr("1"),
			planVersion:   ptr("1"),
			apiKeyWo:      ptr("secret123"),
			wantShouldSet: false,
		},
		{
			name:          "null to value transition triggers update",
			stateVersion:  nil,
			planVersion:   ptr("1"),
			apiKeyWo:      ptr("newsecret"),
			wantShouldSet: true,
			wantValue:     "newsecret",
		},
		{
			name:          "plaintext fallback",
			stateVersion:  nil,
			planVersion:   nil,
			apiKey:        ptr("plaintext_value"),
			wantShouldSet: true,
			wantValue:     "plaintext_value",
		},
		{
			name:            "unknown write-only with no plaintext",
			stateVersion:    ptr("1"),
			planVersion:     ptr("2"),
			apiKeyWoUnknown: true,
			wantShouldSet:   false,
		},
		{
			name:                   "secret required on update ignores version",
			stateVersion:           ptr("1"),
			planVersion:            ptr("1"),
			apiKeyWo:               ptr("secret123"),
			secretRequiredOnUpdate: true,
			wantShouldSet:          true,
			wantValue:              "secret123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testSchema := testWriteOnlySchema()
			handler := WriteOnlySecretHandler{
				Config: WriteOnlySecretConfig{
					OriginalAttr:  "api_key",
					WriteOnlyAttr: "api_key_wo",
					TriggerAttr:   "api_key_wo_version",
				},
				SecretRequiredOnUpdate: tt.secretRequiredOnUpdate,
			}

			priorState := tfsdk.State{
				Raw:    makeVersionValue(tt.stateVersion),
				Schema: testSchema,
			}
			plan := tfsdk.Plan{
				Raw:    makeVersionValue(tt.planVersion),
				Schema: testSchema,
			}

			// Build config value, handling unknown case
			var configRaw tftypes.Value
			if tt.apiKeyWoUnknown {
				apiKeyVal := tftypes.NewValue(tftypes.String, nil)
				configRaw = tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"api_key":            tftypes.String,
						"api_key_wo":         tftypes.String,
						"api_key_wo_version": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"api_key":            apiKeyVal,
					"api_key_wo":         tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					"api_key_wo_version": tftypes.NewValue(tftypes.String, nil),
				})
			} else {
				configRaw = makeConfigValue(tt.apiKey, tt.apiKeyWo)
			}

			tfConfig := tfsdk.Config{
				Raw:    configRaw,
				Schema: testSchema,
			}

			req := resource.UpdateRequest{
				State: priorState,
				Plan:  plan,
			}

			result := handler.GetSecretForUpdate(ctx, &tfConfig, &req)

			if result.Diagnostics.HasError() {
				t.Errorf("unexpected error: %v", result.Diagnostics)
			}
			if result.ShouldSetValue != tt.wantShouldSet {
				t.Errorf("ShouldSetValue = %v, want %v", result.ShouldSetValue, tt.wantShouldSet)
			}
			if tt.wantShouldSet && result.Value != tt.wantValue {
				t.Errorf("Value = %q, want %q", result.Value, tt.wantValue)
			}
		})
	}
}

func TestWriteOnlySecretConfigPaths(t *testing.T) {
	root := WriteOnlySecretConfig{
		OriginalAttr:  "password",
		WriteOnlyAttr: "password_wo",
		TriggerAttr:   "password_wo_version",
	}
	if got, want := root.attrPath("password_wo").String(), "password_wo"; got != want {
		t.Errorf("root attrPath = %q, want %q", got, want)
	}
	if got, want := root.attrExpression("password_wo").String(), "password_wo"; got != want {
		t.Errorf("root attrExpression = %q, want %q", got, want)
	}

	nested := root
	nested.ParentBlocks = []string{"authentication", "basic"}
	if got, want := nested.attrPath("password_wo").String(), "authentication.basic.password_wo"; got != want {
		t.Errorf("nested attrPath = %q, want %q", got, want)
	}
	if got, want := nested.attrExpression("password_wo").String(), "authentication.basic.password_wo"; got != want {
		t.Errorf("nested attrExpression = %q, want %q", got, want)
	}
}

// validatorDescriptions joins the descriptions of an attribute's validators.
// The validators render their path expressions into those descriptions, which is
// the only way to observe the paths from outside the validators package.
func validatorDescriptions(t *testing.T, attr schema.Attribute) string {
	t.Helper()
	stringAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatalf("expected schema.StringAttribute, got %T", attr)
	}
	var descriptions []string
	for _, attrValidator := range stringAttr.Validators {
		descriptions = append(descriptions, attrValidator.Description(context.Background()))
	}
	return strings.Join(descriptions, "\n")
}

func TestCreateWriteOnlySecretAttributesPaths(t *testing.T) {
	root := WriteOnlySecretConfig{
		OriginalAttr:  "password",
		WriteOnlyAttr: "password_wo",
		TriggerAttr:   "password_wo_version",
	}
	nested := root
	nested.ParentBlocks = []string{"authentication", "basic"}

	cases := []struct {
		name   string
		config WriteOnlySecretConfig
		prefix string
	}{
		{name: "root keeps unprefixed paths", config: root},
		{name: "nested block prefixes paths", config: nested, prefix: "authentication.basic."},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			attrs := CreateWriteOnlySecretAttributes(tc.config)
			original := validatorDescriptions(t, attrs["password"])
			writeOnly := validatorDescriptions(t, attrs["password_wo"])
			trigger := validatorDescriptions(t, attrs["password_wo_version"])

			// Validators render a path collection as "[first,second]". Keeping the
			// brackets in the expectation makes each token unambiguous: "[password_wo]"
			// cannot match inside "[password_wo_version]", and an unprefixed
			// expectation cannot match a prefixed path.
			pathList := func(attrs ...string) string {
				prefixed := make([]string, 0, len(attrs))
				for _, attr := range attrs {
					prefixed = append(prefixed, tc.prefix+attr)
				}
				return fmt.Sprintf("%q", "["+strings.Join(prefixed, ",")+"]")
			}
			wantPair := pathList("password", "password_wo")

			// Both members of the ExactlyOneOf pair reference each other.
			for attr, got := range map[string]string{"password": original, "password_wo": writeOnly} {
				if !strings.Contains(got, wantPair) {
					t.Errorf("%s validators should reference %s, got:\n%s", attr, wantPair, got)
				}
			}
			// The write-only attribute requires its version trigger, and vice versa.
			if want := pathList("password_wo_version"); !strings.Contains(writeOnly, want) {
				t.Errorf("password_wo validators should reference %s, got:\n%s", want, writeOnly)
			}
			if want := pathList("password_wo"); !strings.Contains(trigger, want) {
				t.Errorf("password_wo_version validators should reference %s, got:\n%s", want, trigger)
			}
			// PreferWriteOnlyAttribute renders its path unquoted.
			if want := "attribute " + tc.prefix + "password_wo should be preferred"; !strings.Contains(original, want) {
				t.Errorf("password validators should contain %q, got:\n%s", want, original)
			}
		})
	}
}

func testNestedWriteOnlySchema() schema.Schema {
	return schema.Schema{
		Blocks: map[string]schema.Block{
			"authentication": schema.SingleNestedBlock{
				Blocks: map[string]schema.Block{
					"basic": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"username":            schema.StringAttribute{Optional: true},
							"password":            schema.StringAttribute{Optional: true, Sensitive: true},
							"password_wo":         schema.StringAttribute{Optional: true, Sensitive: true, WriteOnly: true},
							"password_wo_version": schema.StringAttribute{Optional: true},
						},
					},
				},
			},
		},
	}
}

func makeNestedConfigValue(password, passwordWo *string) tftypes.Value {
	basicType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"username":            tftypes.String,
		"password":            tftypes.String,
		"password_wo":         tftypes.String,
		"password_wo_version": tftypes.String,
	}}
	authType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"basic": basicType}}
	rootType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"authentication": authType}}

	strVal := func(s *string) tftypes.Value {
		if s == nil {
			return tftypes.NewValue(tftypes.String, nil)
		}
		return tftypes.NewValue(tftypes.String, *s)
	}

	basic := tftypes.NewValue(basicType, map[string]tftypes.Value{
		"username":            tftypes.NewValue(tftypes.String, "datadog"),
		"password":            strVal(password),
		"password_wo":         strVal(passwordWo),
		"password_wo_version": tftypes.NewValue(tftypes.String, nil),
	})
	auth := tftypes.NewValue(authType, map[string]tftypes.Value{"basic": basic})
	return tftypes.NewValue(rootType, map[string]tftypes.Value{"authentication": auth})
}

func TestGetSecretForCreateFromNestedBlock(t *testing.T) {
	handler := &WriteOnlySecretHandler{
		Config: WriteOnlySecretConfig{
			OriginalAttr:  "password",
			WriteOnlyAttr: "password_wo",
			TriggerAttr:   "password_wo_version",
			ParentBlocks:  []string{"authentication", "basic"},
		},
	}

	cases := []struct {
		name       string
		password   *string
		passwordWo *string
		wantSet    bool
		wantValue  string
	}{
		{name: "write-only wins", passwordWo: ptr("wo-secret"), wantSet: true, wantValue: "wo-secret"},
		{name: "plaintext fallback", password: ptr("plain-secret"), wantSet: true, wantValue: "plain-secret"},
		{name: "neither set", wantSet: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config := &tfsdk.Config{
				Schema: testNestedWriteOnlySchema(),
				Raw:    makeNestedConfigValue(tc.password, tc.passwordWo),
			}
			result := handler.GetSecretForCreate(context.Background(), config)
			if result.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics: %v", result.Diagnostics)
			}
			if result.ShouldSetValue != tc.wantSet {
				t.Errorf("ShouldSetValue = %v, want %v", result.ShouldSetValue, tc.wantSet)
			}
			if result.Value != tc.wantValue {
				t.Errorf("Value = %q, want %q", result.Value, tc.wantValue)
			}
		})
	}
}

// modeOnlySecretConfig uses reflection so this red test remains executable
// before T156 adds the Mode and Required contract to WriteOnlySecretConfig.
// A missing field is reported as the intended contract failure rather than a
// package compilation error that would hide the rest of the provider suite.
func modeOnlySecretConfig(t *testing.T, required bool) WriteOnlySecretConfig {
	t.Helper()
	config := WriteOnlySecretConfig{
		OriginalAttr:         "api_key",
		WriteOnlyAttr:        "api_key_wo",
		TriggerAttr:          "api_key_wo_version",
		OriginalDescription:  "The API key for the account.",
		WriteOnlyDescription: "Write-only API key for the account.",
		TriggerDescription:   "Version for `api_key_wo` rotation.",
	}
	value := reflect.ValueOf(&config).Elem()
	mode := value.FieldByName("Mode")
	if !mode.IsValid() {
		t.Fatal("WriteOnlySecretConfig is missing Mode; T156 must add WriteOnlySecretModeOnly")
	}
	switch mode.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		mode.SetInt(1)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		mode.SetUint(1)
	default:
		t.Fatalf("WriteOnlySecretConfig.Mode must be an integer-backed enum, got %s", mode.Kind())
	}
	requiredField := value.FieldByName("Required")
	if !requiredField.IsValid() {
		t.Fatal("WriteOnlySecretConfig is missing Required")
	}
	if requiredField.Kind() != reflect.Bool {
		t.Fatalf("WriteOnlySecretConfig.Required must be bool, got %s", requiredField.Kind())
	}
	requiredField.SetBool(required)
	return config
}

func modeOnlyValue(secret, version *string) tftypes.Value {
	stringValue := func(value *string) tftypes.Value {
		if value == nil {
			return tftypes.NewValue(tftypes.String, nil)
		}
		return tftypes.NewValue(tftypes.String, *value)
	}
	objectType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"api_key_wo":         tftypes.String,
		"api_key_wo_version": tftypes.String,
	}}
	return tftypes.NewValue(objectType, map[string]tftypes.Value{
		"api_key_wo":         stringValue(secret),
		"api_key_wo_version": stringValue(version),
	})
}

func TestCreateWriteOnlySecretAttributesModeOnly(t *testing.T) {
	for _, required := range []bool{false, true} {
		t.Run(fmt.Sprintf("required=%t", required), func(t *testing.T) {
			config := modeOnlySecretConfig(t, required)
			attrs := CreateWriteOnlySecretAttributes(config)

			if len(attrs) != 2 {
				t.Fatalf("mode-only schema must expose exactly _wo and _wo_version, got keys %#v", reflect.ValueOf(attrs).MapKeys())
			}
			if _, exists := attrs[config.OriginalAttr]; exists {
				t.Fatalf("mode-only schema must not expose plaintext attribute %q", config.OriginalAttr)
			}
			writeOnly, ok := attrs[config.WriteOnlyAttr].(schema.StringAttribute)
			if !ok {
				t.Fatalf("%s must be schema.StringAttribute, got %T", config.WriteOnlyAttr, attrs[config.WriteOnlyAttr])
			}
			version, ok := attrs[config.TriggerAttr].(schema.StringAttribute)
			if !ok {
				t.Fatalf("%s must be schema.StringAttribute, got %T", config.TriggerAttr, attrs[config.TriggerAttr])
			}
			if !writeOnly.WriteOnly || writeOnly.Sensitive || writeOnly.Computed {
				t.Errorf("_wo flags = WriteOnly:%t Sensitive:%t Computed:%t; want true, false, false", writeOnly.WriteOnly, writeOnly.Sensitive, writeOnly.Computed)
			}
			if required {
				if !writeOnly.Required || writeOnly.Optional || !version.Required || version.Optional {
					t.Errorf("required mode-only pair must make both attributes Required: _wo=%#v version=%#v", writeOnly, version)
				}
				return
			}
			if !writeOnly.Optional || writeOnly.Required || !version.Optional || version.Required {
				t.Errorf("optional mode-only pair must make both attributes Optional: _wo=%#v version=%#v", writeOnly, version)
			}
			writeOnlyValidators := validatorDescriptions(t, writeOnly)
			versionValidators := validatorDescriptions(t, version)
			if !strings.Contains(writeOnlyValidators, `"[api_key_wo_version]"`) {
				t.Errorf("optional _wo must require its version, got:\n%s", writeOnlyValidators)
			}
			if !strings.Contains(versionValidators, `"[api_key_wo]"`) {
				t.Errorf("optional version must require _wo, got:\n%s", versionValidators)
			}
		})
	}
}

func TestCreateWriteOnlySecretAttributesModeOnlyValidNestedSchema(t *testing.T) {
	config := modeOnlySecretConfig(t, false)
	config.ParentBlocks = []string{"authentication", "basic"}
	basicAttributes := CreateWriteOnlySecretAttributes(config)
	basicAttributes["username"] = schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Readable username returned by the API.",
	}

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"authentication": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Authentication configuration.",
				Attributes: map[string]schema.Attribute{
					"basic": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Basic authentication configuration.",
						Attributes:  basicAttributes,
					},
				},
			},
		},
	}

	if diagnostics := testSchema.ValidateImplementation(context.Background()); diagnostics.HasError() {
		t.Fatalf("mode-only nested resource schema must pass framework validation: %v", diagnostics)
	}
}

func TestWriteOnlySecretHandlerModeOnlyReadsConfigurationWithoutPlaintextFallback(t *testing.T) {
	config := modeOnlySecretConfig(t, false)
	testSchema := schema.Schema{Attributes: CreateWriteOnlySecretAttributes(config)}
	handler := &WriteOnlySecretHandler{Config: config}

	for _, tc := range []struct {
		name      string
		secret    *string
		wantSet   bool
		wantValue string
	}{
		{name: "configured write-only value", secret: ptr("configuration-only-secret"), wantSet: true, wantValue: "configuration-only-secret"},
		{name: "optional secret omitted"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tfConfig := &tfsdk.Config{Raw: modeOnlyValue(tc.secret, ptr("1")), Schema: testSchema}
			result := handler.GetSecretForCreate(context.Background(), tfConfig)
			if result.Diagnostics.HasError() {
				t.Fatalf("mode-only Create must not look up absent plaintext attribute: %v", result.Diagnostics)
			}
			if result.ShouldSetValue != tc.wantSet || result.Value != tc.wantValue {
				t.Errorf("result = (%t, %q), want (%t, %q)", result.ShouldSetValue, result.Value, tc.wantSet, tc.wantValue)
			}
		})
	}
}

func TestWriteOnlySecretHandlerModeOnlyUpdateRotation(t *testing.T) {
	config := modeOnlySecretConfig(t, false)
	testSchema := schema.Schema{Attributes: CreateWriteOnlySecretAttributes(config)}
	handler := &WriteOnlySecretHandler{Config: config}

	for _, tc := range []struct {
		name         string
		priorVersion string
		planVersion  string
		wantSet      bool
	}{
		{name: "unchanged version omits secret", priorVersion: "1", planVersion: "1", wantSet: false},
		{name: "changed version rotates once", priorVersion: "1", planVersion: "2", wantSet: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tfConfig := &tfsdk.Config{Raw: modeOnlyValue(ptr("rotated-secret"), ptr(tc.planVersion)), Schema: testSchema}
			request := &resource.UpdateRequest{
				State: tfsdk.State{Raw: modeOnlyValue(nil, ptr(tc.priorVersion)), Schema: testSchema},
				Plan:  tfsdk.Plan{Raw: modeOnlyValue(nil, ptr(tc.planVersion)), Schema: testSchema},
			}
			result := handler.GetSecretForUpdate(context.Background(), tfConfig, request)
			if result.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics: %v", result.Diagnostics)
			}
			if result.ShouldSetValue != tc.wantSet {
				t.Errorf("ShouldSetValue = %t, want %t", result.ShouldSetValue, tc.wantSet)
			}
			if tc.wantSet && result.Value != "rotated-secret" {
				t.Errorf("Value = %q, want rotated-secret", result.Value)
			}
		})
	}
}
