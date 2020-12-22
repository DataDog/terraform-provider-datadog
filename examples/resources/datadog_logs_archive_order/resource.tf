resource "datadog_logs_archive_order" "sample_archive_order" {
    archive_ids = [
        "${datadog_logs_archive.sample_archive_1.id}",
        "${datadog_logs_archive.sample_archive_2.id}"
    ]
}
