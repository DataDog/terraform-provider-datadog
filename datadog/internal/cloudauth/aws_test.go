package cloudauth

import (
	"net/http"
	"testing"

	awsauth "github.com/DataDog/datadog-api-client-go/v2/auth/aws"
)

func TestNewAWSProvider(t *testing.T) {
	provider, err := NewAWSProvider(AWSConfig{Region: "us-east-1", HTTPClient: http.DefaultClient})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := provider.(*awsauth.Provider); !ok {
		t.Fatalf("provider type = %T, want *awsauth.Provider", provider)
	}
}

func TestValidateAWSCredentials(t *testing.T) {
	tests := []struct {
		name            string
		accessKeyID     string
		secretAccessKey string
		sessionToken    string
		wantErr         bool
	}{
		{
			name: "default credential chain",
		},
		{
			name:            "explicit long-lived credentials",
			accessKeyID:     "access-key",
			secretAccessKey: "secret-key",
		},
		{
			name:            "explicit temporary credentials",
			accessKeyID:     "access-key",
			secretAccessKey: "secret-key",
			sessionToken:    "session-token",
		},
		{
			name:        "missing secret key",
			accessKeyID: "access-key",
			wantErr:     true,
		},
		{
			name:            "missing access key",
			secretAccessKey: "secret-key",
			wantErr:         true,
		},
		{
			name:         "session token without keys",
			sessionToken: "session-token",
			wantErr:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateAWSCredentials(test.accessKeyID, test.secretAccessKey, test.sessionToken)
			if (err != nil) != test.wantErr {
				t.Fatalf("ValidateAWSCredentials() error = %v, wantErr %t", err, test.wantErr)
			}
		})
	}
}
