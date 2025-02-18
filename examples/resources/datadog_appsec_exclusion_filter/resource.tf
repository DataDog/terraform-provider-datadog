# Create new appsec_exclusion_filter resource

resource "datadog_appsec_exclusion_filter" "foo" {
    description = "Exclude false positives on a path"
    enabled = True
    event_query = "UPDATE ME"
    ip_list = "UPDATE ME"
    on_match = "UPDATE ME"
    parameters = "UPDATE ME"
    path_glob = "/accounts/*"
    rules_target {
    rule_id = "dog-913-009"
    tags {
    category = "attack_attempt"
    type = "lfi"
    }
    }
    scope {
    env = "www"
    service = "prod"
    }
}