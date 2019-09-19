resource "datadog_logs_pipeline" "my" {
	name = "updated pipeline"
	is_enabled = false
	filter {
		query = "source:kafka"
	}
	processor {
		date_remapper {
			name = "test date remapper"
			is_enabled = true
			sources = ["verbose"]
		}
	}
	processor {
    		date_remapper {
    			name = "other date remapper"
    			is_enabled = true
    			sources = ["verbose"]
    		}
    	}
	processor {
		status_remapper {
			is_enabled = true
			sources = ["redis.severity"]
		}
	}
	processor {
		attribute_remapper {
			name = "Simple attribute remapper"
			is_enabled = true
			sources = ["db.instance"]
			source_type = "tag"
		  	target = "db"
			target_type = "tag"
			preserve_source = true
			override_on_conflict = false
		}
	}
}

resource "datadog_logs_pipelineorder" "pipelines" {
	name = "pipelines"
	pipelines = [
		"TOYNsNfjTD6zTXVg8_ej1g",
        "VxXfWxegScyjG8mMJwnFIA",
        "GGVTp-5PT_O9Xhmsxnsu_w",
        "VgZXJneKR2qh2WcfAQi6fA",
        "hHNF6-ykTFiW2KrlO4z6kw"
	]
}
