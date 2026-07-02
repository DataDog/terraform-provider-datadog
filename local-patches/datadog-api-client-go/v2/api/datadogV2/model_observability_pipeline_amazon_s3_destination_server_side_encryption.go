// Unless explicitly stated otherwise all files in this repository are licensed under the Apache-2.0 License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2019-Present Datadog, Inc.

// NOTE: This file is a local hand-patch. It will be superseded once the datadog-api-spec
// change is merged and the datadog-api-client-go is regenerated.

package datadogV2

import (
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
)

// ObservabilityPipelineAmazonS3DestinationServerSideEncryption Server-side encryption algorithm used when storing objects in S3.
type ObservabilityPipelineAmazonS3DestinationServerSideEncryption string

// List of ObservabilityPipelineAmazonS3DestinationServerSideEncryption.
const (
	OBSERVABILITYPIPELINEAMAZONS3DESTINATIONSERVERSIDEENCRYPTION_AWS_KMS ObservabilityPipelineAmazonS3DestinationServerSideEncryption = "aws:kms"
	OBSERVABILITYPIPELINEAMAZONS3DESTINATIONSERVERSIDEENCRYPTION_AES256  ObservabilityPipelineAmazonS3DestinationServerSideEncryption = "AES256"
)

var allowedObservabilityPipelineAmazonS3DestinationServerSideEncryptionEnumValues = []ObservabilityPipelineAmazonS3DestinationServerSideEncryption{
	OBSERVABILITYPIPELINEAMAZONS3DESTINATIONSERVERSIDEENCRYPTION_AWS_KMS,
	OBSERVABILITYPIPELINEAMAZONS3DESTINATIONSERVERSIDEENCRYPTION_AES256,
}

// GetAllowedValues returns the list of possible values.
func (v *ObservabilityPipelineAmazonS3DestinationServerSideEncryption) GetAllowedValues() []ObservabilityPipelineAmazonS3DestinationServerSideEncryption {
	return allowedObservabilityPipelineAmazonS3DestinationServerSideEncryptionEnumValues
}

// UnmarshalJSON deserializes the given payload.
func (v *ObservabilityPipelineAmazonS3DestinationServerSideEncryption) UnmarshalJSON(src []byte) error {
	var value string
	err := datadog.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	*v = ObservabilityPipelineAmazonS3DestinationServerSideEncryption(value)
	return nil
}

// NewObservabilityPipelineAmazonS3DestinationServerSideEncryptionFromValue returns a pointer to a valid ObservabilityPipelineAmazonS3DestinationServerSideEncryption
// for the value passed as argument, or an error if the value passed is not allowed by the enum.
func NewObservabilityPipelineAmazonS3DestinationServerSideEncryptionFromValue(v string) (*ObservabilityPipelineAmazonS3DestinationServerSideEncryption, error) {
	ev := ObservabilityPipelineAmazonS3DestinationServerSideEncryption(v)
	if ev.IsValid() {
		return &ev, nil
	}
	return nil, fmt.Errorf("invalid value '%v' for ObservabilityPipelineAmazonS3DestinationServerSideEncryption: valid values are %v", v, allowedObservabilityPipelineAmazonS3DestinationServerSideEncryptionEnumValues)
}

// IsValid return true if the value is valid for the enum, false otherwise.
func (v ObservabilityPipelineAmazonS3DestinationServerSideEncryption) IsValid() bool {
	for _, existing := range allowedObservabilityPipelineAmazonS3DestinationServerSideEncryptionEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to ObservabilityPipelineAmazonS3DestinationServerSideEncryption value.
func (v ObservabilityPipelineAmazonS3DestinationServerSideEncryption) Ptr() *ObservabilityPipelineAmazonS3DestinationServerSideEncryption {
	return &v
}
