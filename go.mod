module github.com/terraform-providers/terraform-provider-datadog

require (
	github.com/DataDog/datadog-api-client-go v1.0.0-beta.16
	github.com/DataDog/datadog-go v3.6.0+incompatible // indirect
	github.com/cenkalti/backoff v2.1.1+incompatible // indirect
	github.com/dnaeon/go-vcr v1.0.1
	github.com/fatih/color v1.9.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/terraform-plugin-docs v0.4.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.4.3
	github.com/jonboulle/clockwork v0.1.0
	github.com/kr/pretty v0.2.0
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/zorkian/go-datadog-api v2.30.0+incompatible
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	gopkg.in/DataDog/dd-trace-go.v1 v1.29.0-rc.1.0.20210226170446-a8dc39ec3484
)

go 1.15

// Use branch of dd-trace-go for additional APM features
replace gopkg.in/DataDog/dd-trace-go.v1 => github.com/DataDog/dd-trace-go v1.29.0-alpha.1.0.20210128154316-c84d7933b726
