module github.com/terraform-providers/terraform-provider-datadog

require (
	github.com/DataDog/datadog-api-client-go v1.0.0-beta.17.0.20210316193424-3d75b3867ae8
	github.com/DataDog/datadog-go v3.6.0+incompatible // indirect
	github.com/cenkalti/backoff v2.1.1+incompatible // indirect
	github.com/dnaeon/go-vcr v1.0.1
	github.com/fatih/color v1.10.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1
	github.com/hashicorp/go-hclog v0.12.0 // indirect
	github.com/hashicorp/go-plugin v1.3.0 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/terraform-plugin-docs v0.4.0
	github.com/hashicorp/terraform-plugin-sdk v1.15.0
	github.com/jonboulle/clockwork v0.2.2
	github.com/kr/pretty v0.2.0
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/zorkian/go-datadog-api v2.30.0+incompatible
	golang.org/x/crypto v0.0.0-20210317152858-513c2a44f670 // indirect
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210317091845-390168757d9c // indirect
	golang.org/x/term v0.0.0-20210317153231-de623e64d2a6 // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	golang.org/x/tools v0.1.0 // indirect
	google.golang.org/grpc v1.30.0 // indirect
	gopkg.in/DataDog/dd-trace-go.v1 v1.29.0-rc.1.0.20210226170446-a8dc39ec3484
	gotest.tools/gotestsum v1.6.2 // indirect
)

go 1.16

// Use custom fork with performance fix in DecoderSpec
replace github.com/hashicorp/terraform-plugin-sdk v1.15.0 => github.com/therve/terraform-plugin-sdk v1.16.1-0.20210202202613-4d59f03d3b5f

// Use branch of dd-trace-go for additional APM features
replace gopkg.in/DataDog/dd-trace-go.v1 => github.com/DataDog/dd-trace-go v1.29.0-alpha.1.0.20210128154316-c84d7933b726
