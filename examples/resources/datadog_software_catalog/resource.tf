// v3 service entity 
resource "datadog_software_catalog" "service_v3" {
  entity = <<EOF
EOF
}

// v3 datastore entity 
resource "datadog_software_catalog" "datastore_v3" {
  entity = <<EOF
EOF
}

// v3 queue entity 
resource "datadog_software_catalog" "queue_v3" {
  entity = <<EOF
EOF
}

// v3 system entity 
resource "datadog_software_catalog" "system_v3" {
  entity = <<EOF
EOF
}
