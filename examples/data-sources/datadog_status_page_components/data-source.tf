data "datadog_status_page" "example" {
  id = "ba32b61d-ba85-45b3-9976-fcc39e808351"
}

data "datadog_status_page_components" "all" {
  status_page_id = data.datadog_status_page.example.id
}
