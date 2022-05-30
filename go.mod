module github.com/terraform-providers/terraform-provider-datadog

require (
	github.com/DataDog/datadog-api-client-go v1.14.1-0.20220519094618-d6d952dd0074
	github.com/DataDog/datadog-go v4.8.3+incompatible // indirect
	github.com/DataDog/dd-sdk-go-testing v0.0.0-20211116174033-1cd082e322ad
	github.com/DataDog/sketches-go v1.2.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/dnaeon/go-vcr v1.0.1
	github.com/fatih/color v1.10.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/terraform-plugin-docs v0.8.1
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.16.0
	github.com/jonboulle/clockwork v0.2.2
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/tinylib/msgp v1.1.6 // indirect
	github.com/zorkian/go-datadog-api v2.30.0+incompatible
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	golang.org/x/sys v0.0.0-20211123173158-ef496fb156ab // indirect
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11 // indirect
	gopkg.in/DataDog/dd-trace-go.v1 v1.34.0
	gopkg.in/warnings.v0 v0.1.2
)

go 1.16

// Use custom fork of tfplugindocs to fix a bug in doc generation https://github.com/DataDog/terraform-provider-datadog/issues/1024
replace github.com/hashicorp/terraform-plugin-docs v0.8.1 => github.com/skarimo/terraform-plugin-docs v0.8.2-0.20220511175755-afbb94458dc8
