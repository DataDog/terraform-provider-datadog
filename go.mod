module github.com/terraform-providers/terraform-provider-datadog

require (
	github.com/DataDog/datadog-api-client-go v1.2.1-0.20210720115543-0d7c91d1e2e9
	github.com/DataDog/datadog-go v3.6.0+incompatible // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/dnaeon/go-vcr v1.0.1
	github.com/fatih/color v1.9.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/terraform-plugin-docs v0.4.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.7.0
	github.com/jonboulle/clockwork v0.1.0
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/zorkian/go-datadog-api v2.30.0+incompatible
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007 // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	gopkg.in/DataDog/dd-trace-go.v1 v1.30.0-rc.1.0.20210420124628-f63633f38e8f
	gopkg.in/warnings.v0 v0.1.2
)

go 1.16

// Use custom fork of tfplugindocs to fix a bug in doc generation https://github.com/DataDog/terraform-provider-datadog/issues/1024
replace github.com/hashicorp/terraform-plugin-docs v0.4.0 => github.com/zippolyte/terraform-plugin-docs v0.4.1-0.20210422155525-d4f2c7590b53

// Use branch of dd-trace-go for additional APM features
replace gopkg.in/DataDog/dd-trace-go.v1 => github.com/DataDog/dd-trace-go v1.29.0-alpha.1.0.20210128154316-c84d7933b726
