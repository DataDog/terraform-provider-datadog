# Configure flaky test management policies for a repository

resource "datadog_flaky_tests_management_policies" "my_repo" {
  repository_id = "github.com/my-org/my-repo"

  attempt_to_fix {
    retries = 5
  }

  quarantined {
    enabled = true

    auto_quarantine_rule {
      enabled        = true
      window_seconds = 1209600
    }

    branch_rule {
      enabled                = true
      branches               = ["main"]
      excluded_branches      = ["release/*"]
      excluded_test_services = ["my-legacy-service"]
    }

    failure_rate_rule {
      enabled   = true
      branches  = ["main"]
      min_runs  = 10
      threshold = 0.3
    }
  }

  disabled {
    enabled = true

    auto_disable_rule {
      enabled        = true
      status         = "quarantined"
      window_seconds = 2592000
    }

    branch_rule {
      enabled  = true
      branches = ["main"]
    }

    failure_rate_rule {
      enabled   = true
      branches  = ["main"]
      min_runs  = 10
      status    = "quarantined"
      threshold = 0.5
    }
  }
}
