## 1.0.2 (Unreleased)
## 1.0.1 (December 06, 2017)

BUG FIXES:

* Fix issue reading resources that have been updated outside of Terraform ([#34](https://github.com/terraform-providers/terraform-provider-datadog/issues/34))

## 1.0.0 (October 20, 2017)

BUG FIXES:

* Improved detection of "drift" when graphs are reconfigured outside of Terraform. ([#27](https://github.com/terraform-providers/terraform-provider-datadog/issues/27))
* Fixed API response decoding error on graphs. ([#27](https://github.com/terraform-providers/terraform-provider-datadog/issues/27))

## 0.1.1 (September 26, 2017)

FEATURES: 

* **New Resource:** `datadog_metric_metadata` ([#17](https://github.com/terraform-providers/terraform-provider-datadog/issues/17))


## 0.1.0 (June 20, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
