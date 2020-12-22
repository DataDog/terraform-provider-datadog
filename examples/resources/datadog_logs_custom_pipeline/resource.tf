resource "datadog_logs_custom_pipeline" "sample_pipeline" {
    filter {
        query = "source:foo"
    }
    name = "sample pipeline"
    is_enabled = true
    processor {
        arithmetic_processor {
            expression = "(time1 - time2)*1000"
            target = "my_arithmetic"
            is_replace_missing = true
            name = "sample arithmetic processor"
            is_enabled = true
        }
    }
    processor {
        attribute_remapper {
            sources = ["db.instance"]
            source_type = "tag"
            target = "db"
            target_type = "attribute"
            target_format = "string"
            preserve_source = true
            override_on_conflict = false
            name = "sample attribute processor"
            is_enabled = true
        }
    }
    processor {
        category_processor {
            target = "foo.severity"
            category {
                name = "debug"
                filter {
                    query = "@severity: \".\""
                }
            }
            category {
                name = "verbose"
                filter {
                    query = "@severity: \"-\""
                }
            }
            name = "sample category processor"
            is_enabled = true
        }
    }
    processor {
        date_remapper {
            sources = ["_timestamp", "published_date"]
            name = "sample date remapper"
            is_enabled = true
        }
    }
    processor {
        geo_ip_parser {
            sources = ["network.client.ip"]
            target = "network.client.geoip"
            name = "sample geo ip parser"
            is_enabled = true
        }
    }
    processor {
        grok_parser {
            samples = ["sample log 1"]
            source = "message"
            grok {
                support_rules = ""
                match_rules = "Rule %%{word:my_word2} %%{number:my_float2}"
            }
            name = "sample grok parser"
            is_enabled = true
        }
    }
    processor {
        lookup_processor {
            source = "service_id"
            target = "service_name"
            lookup_table = ["1,my service"]
            default_lookup = "unknown service"
            name = "sample lookup processor"
            is_enabled = true
        }
    }
    processor {
        message_remapper {
            sources = ["msg"]
            name = "sample message remapper"
            is_enabled = true
        }
    }
    processor {
        pipeline {
            filter {
                query = "source:foo"
            }
            processor {
                url_parser {
                    name = "sample url parser"
                    sources = ["url", "extra"]
                    target = "http_url"
                    normalize_ending_slashes = true
                }
            }
            name = "nested pipeline"
            is_enabled = true
        }
    }
    processor {
        service_remapper {
            sources = ["service"]
            name = "sample service remapper"
            is_enabled = true
        }
    }
    processor {
        status_remapper {
            sources = ["info", "trace"]
            name = "sample status remapper"
            is_enabled = true
        }
    }
    processor {
        string_builder_processor {
            target = "user_activity"
            template = "%%{user.name} logged in at %%{timestamp}"
            name = "sample string builder processor"
            is_enabled = true
            is_replace_missing = false
        }
    }
    processor {
        trace_id_remapper {
            sources = ["dd.trace_id"]
            name = "sample trace id remapper"
            is_enabled = true
        }
    }
    processor {
        user_agent_parser {
            sources = ["user", "agent"]
            target = "http_agent"
            is_encoded = false
            name = "sample user agent parser"
            is_enabled = true
        }
    }
}
