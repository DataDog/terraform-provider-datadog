package observability_pipeline

import (
	"encoding/json"
	"testing"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandMtlsServerTlsWithClientIdentityAllowlist(t *testing.T) {
	tls := ExpandMtlsServerTlsWithClientIdentity([]MtlsServerTlsWithClientIdentityModel{{
		CrtFile:           types.StringValue("/etc/certs/server.crt"),
		CaFile:            types.StringValue("/etc/certs/ca.crt"),
		KeyFile:           types.StringValue("/etc/certs/server.key"),
		VerifyCertificate: types.BoolValue(true),
		ClientIdentity: []ClientIdentityModel{{
			Mode: types.StringValue("audit"),
			Allowlist: []ClientIdentityAllowlistModel{{
				Rules: []ClientIdentityRuleModel{
					{
						Name:        types.StringValue("collectors"),
						CommonNames: []types.String{types.StringValue("collector-*")},
						DnsSans:     []types.String{types.StringValue("*.collectors.example.com")},
					},
					{
						Name:    types.StringValue("spiffe"),
						UriSans: []types.String{types.StringValue("spiffe://example.com/ns/prod/**")},
					},
				},
			}},
		}},
	}})

	payload, err := json.Marshal(tls)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"crt_file": "/etc/certs/server.crt",
		"ca_file": "/etc/certs/ca.crt",
		"key_file": "/etc/certs/server.key",
		"verify_certificate": true,
		"client_identity": {
			"mode": "audit",
			"allowlist": {
				"rules": [
					{
						"name": "collectors",
						"common_names": ["collector-*"],
						"dns_sans": ["*.collectors.example.com"]
					},
					{
						"name": "spiffe",
						"uri_sans": ["spiffe://example.com/ns/prod/**"]
					}
				]
			}
		}
	}`, string(payload))
}

func TestExpandMtlsServerTlsWithClientIdentityAllowlistFile(t *testing.T) {
	tls := ExpandMtlsServerTlsWithClientIdentity([]MtlsServerTlsWithClientIdentityModel{{
		CrtFile:           types.StringValue("/etc/certs/server.crt"),
		CaFile:            types.StringValue("/etc/certs/ca.crt"),
		VerifyCertificate: types.BoolValue(true),
		ClientIdentity: []ClientIdentityModel{{
			Mode:          types.StringValue("enforce"),
			AllowlistFile: types.StringValue("client-allowlist.yaml"),
		}},
	}})

	payload, err := json.Marshal(tls)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"crt_file": "/etc/certs/server.crt",
		"ca_file": "/etc/certs/ca.crt",
		"verify_certificate": true,
		"client_identity": {
			"mode": "enforce",
			"allowlist_file": "client-allowlist.yaml"
		}
	}`, string(payload))
}

func TestFlattenMtlsServerTlsWithClientIdentity(t *testing.T) {
	var tls datadogV2.ObservabilityPipelineMtlsServerTlsWithClientIdentity
	err := json.Unmarshal([]byte(`{
		"crt_file": "/etc/certs/server.crt",
		"ca_file": "/etc/certs/ca.crt",
		"verify_certificate": true,
		"client_identity": {
			"allowlist": {
				"rules": [{"name": "collectors", "dns_sans": ["*.collectors.example.com"]}]
			}
		}
	}`), &tls)
	require.NoError(t, err)

	out := FlattenMtlsServerTlsWithClientIdentity(&tls)
	require.Len(t, out, 1)
	assert.Equal(t, types.BoolValue(true), out[0].VerifyCertificate)
	assert.True(t, out[0].KeyFile.IsNull())
	require.Len(t, out[0].ClientIdentity, 1)

	identity := out[0].ClientIdentity[0]
	assert.Equal(t, types.StringValue("enforce"), identity.Mode, "omitted mode should read back as the schema default")
	assert.True(t, identity.AllowlistFile.IsNull())
	require.Len(t, identity.Allowlist, 1)
	assert.Equal(t, []ClientIdentityRuleModel{{
		Name:    types.StringValue("collectors"),
		DnsSans: []types.String{types.StringValue("*.collectors.example.com")},
	}}, identity.Allowlist[0].Rules)
}

func TestFlattenMtlsServerTlsWithClientIdentityWithoutClientIdentity(t *testing.T) {
	tls := datadogV2.NewObservabilityPipelineMtlsServerTlsWithClientIdentity("/etc/certs/server.crt")

	out := FlattenMtlsServerTlsWithClientIdentity(tls)
	require.Len(t, out, 1)
	assert.Empty(t, out[0].ClientIdentity)
}
