package cloudauth

import (
	"net/http"
	"testing"

	awsauth "github.com/DataDog/datadog-api-client-go/auth/aws"
)

func TestNewAWSProvider(t *testing.T) {
	tests := []struct {
		name    string
		config  AWSConfig
		wantErr bool
	}{
		{
			name:   "default credential chain",
			config: AWSConfig{Region: "us-east-1", HTTPClient: http.DefaultClient},
		},
		{
			name: "explicit long-lived credentials",
			config: AWSConfig{
				AccessKeyID:     "access-key",
				SecretAccessKey: "secret-key",
			},
		},
		{
			name: "explicit temporary credentials",
			config: AWSConfig{
				AccessKeyID:     "access-key",
				SecretAccessKey: "secret-key",
				SessionToken:    "session-token",
			},
		},
		{
			name:    "missing secret key",
			config:  AWSConfig{AccessKeyID: "access-key"},
			wantErr: true,
		},
		{
			name:    "session token without keys",
			config:  AWSConfig{SessionToken: "session-token"},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider, err := NewAWSProvider(test.config)
			if (err != nil) != test.wantErr {
				t.Fatalf("NewAWSProvider() error = %v, wantErr %t", err, test.wantErr)
			}
			if test.wantErr {
				return
			}
			if _, ok := provider.(*awsauth.Provider); !ok {
				t.Fatalf("provider type = %T, want *awsauth.Provider", provider)
			}
		})
	}
}
