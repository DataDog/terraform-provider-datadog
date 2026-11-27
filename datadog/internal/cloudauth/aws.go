package cloudauth

import (
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	awsauth "github.com/DataDog/datadog-api-client-go/v2/auth/aws"
)

// AWSConfig contains the Terraform provider settings that influence AWS
// delegated authentication.
type AWSConfig struct {
	Region     string
	HTTPClient *http.Client
}

// NewAWSProvider creates the optional AWS delegated-authentication adapter.
// Explicit credentials are supplied through datadog.ContextAWSVariables; when
// none are supplied the AWS SDK default configuration and credential chain is used.
func NewAWSProvider(config AWSConfig) (datadog.DelegatedTokenProvider, error) {
	options := []awsauth.Option{awsauth.WithRegion(config.Region)}
	if config.HTTPClient != nil {
		options = append(options, awsauth.WithHTTPClient(config.HTTPClient))
	}

	provider, err := awsauth.New(options...)
	if err != nil {
		return nil, fmt.Errorf("configuring AWS delegated authentication: %w", err)
	}
	return provider, nil
}

// ValidateAWSCredentials rejects incomplete explicit Terraform credentials even
// when authentication is deferred by validate=false.
func ValidateAWSCredentials(accessKeyID, secretAccessKey, sessionToken string) error {
	if accessKeyID != "" || secretAccessKey != "" || sessionToken != "" {
		if accessKeyID == "" || secretAccessKey == "" {
			return fmt.Errorf("aws_access_key_id and aws_secret_access_key must both be set when either is configured")
		}
	}
	return nil
}
