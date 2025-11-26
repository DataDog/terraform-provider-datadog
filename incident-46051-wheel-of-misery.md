# Wheel of Misery Exercise: Incident 46051
## Reference Tables GRPC SaveTable Endpoint Availability Failure

**Incident ID**: 46051  
**Date**: November 20, 2025  
**Severity**: SEV-5  
**Environment**: Staging  
**Service**: reference-tables-grpc  
**Team**: redapl-experiences

---

## Expected Alert

### Primary Alert
- **Monitor**: Synthetics test failure for Reference Tables GRPC uploads
  - Test ID: `ajx-zbt-khs`
  - Location: Staging environment
  - Alert Channel: `slack-reference-tables-ops`

### Secondary Alert
- **Monitor**: `[reference-tables-grpc] Low availability in {{datacenter.name}}`
  - Metric: Availability of `savetable` endpoint drops below 90%
  - Query: `avg(last_15m):((1 - (default_zero(count:grpc.dd.server.call.duration{(grpc.status:data_loss OR grpc.status:deadline_exceeded OR grpc.status:internal OR grpc.status:resource_exhausted OR grpc.status:unavailable OR grpc.status:unimplemented OR grpc.status:unknown) AND NOT datacenter:us4.prod.dog AND NOT grpc.method:grpc.health.v1.health/check AND NOT grpc.method:grpc.health.v1.health/watch AND datacenter:* AND service:reference-tables-grpc} by {datacenter,grpc.method}.as_rate()) / clamp_min(default_zero(count:grpc.dd.server.call.duration{NOT datacenter:us4.prod.dog AND NOT grpc.method:grpc.health.v1.health/check AND NOT grpc.method:grpc.health.v1.health/watch AND datacenter:* AND service:reference-tables-grpc} by {datacenter,grpc.method}.as_rate()), 0.1983742558288575))) * 100) < 90`
  - Alert Channels: `pagerduty-reference-tables-high-urgency`, `slack-reference-tables-ops`

### Error Indicators
- **API Error**: `InternalError: We encountered an internal error. Please try again.`
- **Logs**: Internal errors in the `savetable` endpoint
- **Code Location**: `domains/redapl/apps/apis/reference-tables-grpc/server.go` (around line 2xxx)

---

## Expected Response

### 1. Initial Investigation
1. **Check Synthetic Test Trace**
   - Navigate to: `https://ddstaging.datadoghq.com/synthetics/details/ajx-zbt-khs`
   - Review waterfall view to identify failure point
   - Check timing and error details

2. **Review Availability Dashboard**
   - Navigate to: `https://ddstaging.datadoghq.com/dashboard/p6u-au8-7v6/reference-tables-grpc-tf`
   - Analyze availability trends for `savetable` endpoint
   - Identify time window of degradation

3. **Examine Application Logs**
   - Query logs for `service:reference-tables-grpc` and `savetable` errors
   - Look for `InternalError` messages
   - Check specific host logs if available

### 2. Root Cause Analysis
1. **Identify Error Pattern**
   - Review error messages in logs
   - Check for S3-related failures
   - Examine stack traces or error details

2. **Check Recent Changes**
   - Review recent deployments or feature flag changes
   - Look for commits related to S3 upload functionality
   - Check feature flag status: `reference-tables-s3-chaos-complete-multipart-upload`

3. **Verify Root Cause**
   - Confirm that the feature flag `reference-tables-s3-chaos-complete-multipart-upload` is enabled
   - Review commit: `f0bb4f81765f6a44fe97926944c588ab4e95de5a`
   - Understand how this flag affects S3 multipart upload completion

### 3. Resolution Steps
1. **Disable Feature Flag**
   - Navigate to Mosaic: `https://mosaic.us1.ddbuild.io/feature-flags/reference-tables-s3-chaos-complete-multipart-upload`
   - Set targeting rule to `DisableFF`
   - Verify the change is live

2. **Verify Recovery**
   - Monitor synthetic test: `https://ddstaging.datadoghq.com/synthetics/details/ajx-zbt-khs`
   - Check that test results return to passing state
   - Verify availability metrics return to normal (>90%)

3. **Incident Closure**
   - Mark incident as `stable` once recovery confirmed
   - Mark incident as `resolved` after verification
   - Update incident summary with root cause

---

## How to Induce the Failure State

### Prerequisites
- Access to Mosaic feature flag management
- Access to staging environment
- Ability to monitor synthetic tests and dashboards

### Steps to Induce Failure

1. **Enable the Chaos Feature Flag**
   - Navigate to Mosaic feature flag management: `https://mosaic.us1.ddbuild.io/feature-flags/reference-tables-s3-chaos-complete-multipart-upload`
   - Enable the feature flag `reference-tables-s3-chaos-complete-multipart-upload`
   - Ensure the flag is targeting the staging environment
   - Verify the flag is live

2. **Trigger the Failure**
   - The feature flag causes S3 multipart upload completion to fail
   - This failure cascades to the `savetable` endpoint, causing `InternalError` responses
   - No additional steps required - the flag activation will cause failures

3. **Expected Failure Manifestation**
   - **Synthetic Test**: Will begin failing within minutes of flag activation
   - **API Errors**: `savetable` endpoint will return `InternalError: We encountered an internal error. Please try again.`
   - **Availability**: Will drop below 90% threshold
   - **Logs**: Will show internal errors related to S3 upload failures

### Verification of Failure State

1. **Check Synthetic Test Status**
   - Monitor: `https://ddstaging.datadoghq.com/synthetics/details/ajx-zbt-khs`
   - Should show failing tests within 5-15 minutes

2. **Monitor Availability Dashboard**
   - Check: `https://ddstaging.datadoghq.com/dashboard/p6u-au8-7v6/reference-tables-grpc-tf`
   - Availability should drop below 90%

3. **Review Logs**
   - Query: `service:reference-tables-grpc` AND `savetable` AND `error`
   - Should show `InternalError` messages
   - Look for S3-related error context

### Cleanup (To Restore Normal State)

1. **Disable the Feature Flag**
   - Navigate back to Mosaic
   - Set targeting rule to `DisableFF`
   - Verify flag is disabled

2. **Verify Recovery**
   - Synthetic test should pass within 5-15 minutes
   - Availability should return above 90%
   - Logs should show successful requests

---

## Additional Context

### Related Code
- **Commit**: `f0bb4f81765f6a44fe97926944c588ab4e95de5a`
- **File**: `domains/redapl/apps/apis/reference-tables-grpc/server.go`
- **Feature Flag**: `reference-tables-s3-chaos-complete-multipart-upload`

### Monitoring Resources
- **Synthetic Test**: `ajx-zbt-khs`
- **Dashboard**: `p6u-au8-7v6` (Reference Tables GRPC TF)
- **Slack Channel**: `#slack-reference-tables-ops`
- **PagerDuty**: `pagerduty-reference-tables-high-urgency`

### Notes
- This is a chaos engineering feature flag designed to test failure scenarios
- The failure is intentional and controlled via feature flag
- Recovery is immediate upon disabling the flag
- No data loss or permanent damage occurs - this is a transient failure scenario

