package cloudauth

import (
	"fmt"
	"net/http"

	awsauth "github.com/DataDog/datadog-api-client-go/auth/aws"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
)

// AWSConfig contains the Terraform provider settings that influence AWS
// delegated authentication.
type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	HTTPClient      *http.Client
}

// NewAWSProvider creates the optional AWS delegated-authentication adapter.
// Explicit Terraform credentials take precedence; when none are supplied the
// AWS SDK default configuration and credential chain is used.
func NewAWSProvider(config AWSConfig) (datadog.DelegatedTokenProvider, error) {
	options := []awsauth.Option{awsauth.WithRegion(config.Region)}
	if config.HTTPClient != nil {
		options = append(options, awsauth.WithHTTPClient(config.HTTPClient))
	}

	anyStaticCredential := config.AccessKeyID != "" || config.SecretAccessKey != "" || config.SessionToken != ""
	if anyStaticCredential {
		if config.AccessKeyID == "" || config.SecretAccessKey == "" {
			return nil, fmt.Errorf("aws_access_key_id and aws_secret_access_key must both be set when either is configured")
		}
		options = append(options, awsauth.WithStaticCredentials(config.AccessKeyID, config.SecretAccessKey, config.SessionToken))
	}

	provider, err := awsauth.New(options...)
	if err != nil {
		return nil, fmt.Errorf("configuring AWS delegated authentication: %w", err)
	}
	return provider, nil
}
