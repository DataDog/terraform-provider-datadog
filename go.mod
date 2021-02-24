module github.com/terraform-providers/terraform-provider-datadog

require (
	github.com/DataDog/datadog-api-client-go v1.0.0-beta.15.0.20210223212241-85384a53160c
	github.com/cenkalti/backoff v2.1.1+incompatible // indirect
	github.com/dnaeon/go-vcr v1.0.1
	github.com/fatih/color v1.9.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1
	github.com/hashicorp/go-hclog v0.12.0 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/terraform-plugin-docs v0.4.0
	github.com/hashicorp/terraform-plugin-sdk v1.15.0
	github.com/jonboulle/clockwork v0.1.0
	github.com/kr/pretty v0.2.0
	github.com/stretchr/objx v0.1.1 // indirect
	github.com/zorkian/go-datadog-api v2.30.0+incompatible
	gopkg.in/DataDog/dd-trace-go.v1 v1.29.0-alpha.1.0.20210212162759-6903fc5e42cd
)

go 1.15

// Use custom fork with performance fix in DecoderSpec
replace github.com/hashicorp/terraform-plugin-sdk v1.15.0 => github.com/therve/terraform-plugin-sdk v1.16.1-0.20210202202613-4d59f03d3b5f

// Use branch of dd-trace-go for additional APM features
replace gopkg.in/DataDog/dd-trace-go.v1 => github.com/DataDog/dd-trace-go v1.29.0-alpha.1.0.20210128154316-c84d7933b726
