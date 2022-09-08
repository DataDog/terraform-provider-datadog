resource "datadog_service_definition" "service_definition" {
  service_definition = <<EOF
{
  "schema-version": "v2",
  "dd-service": "testservice",
  "team": "Team A",
  "contacts": [],
  "repos": [],
  "tags": [],
  "integrations": {},
  "dd-team": "team-a",
  "docs": [],
  "extensions": {},
  "links": []
}
EOF
}
