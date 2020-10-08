package datadog

import "fmt"

func archiveOrderConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_logs_archive" "archive_1" {
	name = "%s-first"
	is_enabled = true
	filter {
		query = "source:redis"
	}
}
resource "datadog_logs_archive" "archive_2" {
	name = "%s-second"
	is_enabled = true
	filter {
		query = "source:agent"
	}
}

resource "datadog_logs_archive_order" "archives" {
	depends_on = [
		"datadog_logs_archive.archive_1",
		"datadog_logs_archive.archive_2"
	]
	name = "%s"
	archives = [
		"${datadog_logs_archive.archive_1.id}",
		"${datadog_logs_archive.archive_2.id}"
	]
}`, uniq, uniq, uniq)
}

func ArchiveOrderUpdateConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_logs_archive" "archive_1" {
	name = "%s-first"
	is_enabled = true
	filter {
		query = "source:redis"
	}
}
resource "datadog_logs_archive" "archive_2" {
	name = "%s-second"
	is_enabled = true
	filter {
		query = "source:agent"
	}
}

resource "datadog_logs_archive_order" "archives" {
	depends_on = [
		"datadog_logs_archive.archive_1",
		"datadog_logs_archive.archive_2"
	]
	name = "%s"
	archives = [
		"${datadog_logs_archive.archive_2.id}",
		"${datadog_logs_archive.archive_1.id}"
	]
}`, uniq, uniq, uniq)
}
