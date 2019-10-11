module github.com/terraform-providers/terraform-provider-datadog

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

require (
	github.com/cenkalti/backoff v2.1.1+incompatible // indirect
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.0
	github.com/hashicorp/terraform v0.12.5
	github.com/kr/pretty v0.1.0
	github.com/zorkian/go-datadog-api v2.24.0+incompatible
)
