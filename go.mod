module github.com/terraform-providers/terraform-provider-datadog

require (
	github.com/DataDog/datadog-api-client-go v1.5.0
	github.com/DataDog/datadog-go v4.8.2+incompatible // indirect
	github.com/DataDog/dd-sdk-go-testing v0.0.0-20210929140144-5d69f0a9bd49
	github.com/DataDog/sketches-go v1.2.1 // indirect
	github.com/Microsoft/go-winio v0.5.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/dnaeon/go-vcr v1.0.1
	github.com/fatih/color v1.10.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/terraform-plugin-docs v0.4.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.7.0
	github.com/jonboulle/clockwork v0.2.2
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/tinylib/msgp v1.1.6 // indirect
	github.com/zorkian/go-datadog-api v2.30.0+incompatible
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	golang.org/x/sys v0.0.0-20211015200801-69063c4bb744 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	golang.org/x/tools v0.1.0 // indirect
	gopkg.in/DataDog/dd-trace-go.v1 v1.33.0
	gopkg.in/warnings.v0 v0.1.2
)

go 1.16

// Use custom fork of tfplugindocs to fix a bug in doc generation https://github.com/DataDog/terraform-provider-datadog/issues/1024
replace github.com/hashicorp/terraform-plugin-docs v0.4.0 => github.com/zippolyte/terraform-plugin-docs v0.4.1-0.20210422155525-d4f2c7590b53
