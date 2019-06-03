module github.com/terraform-providers/terraform-provider-datadog

require (
	github.com/cenkalti/backoff v0.0.0-20161020194410-b02f2bbce11d // indirect
	github.com/hashicorp/go-cleanhttp v0.5.0
	github.com/hashicorp/terraform v0.12.0
	github.com/kr/pretty v0.1.0
	github.com/zorkian/go-datadog-api v2.20.1-0.20190521074352-d479e1923790+incompatible
)

replace github.com/zorkian/go-datadog-api => github.com/DataDog/go-datadog-api v0.0.0-20190529090230-16add2864293
