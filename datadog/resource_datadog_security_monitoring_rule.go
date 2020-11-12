package datadog

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceDatadogSecurityMonitoringRuleResponse() *schema.Resource {
	return &schema.Resource{
		// TODO: resource vs data source
		Exists: resourceDatadogSecurityMonitoringRuleResponseExists,
		Create: resourceDatadogSecurityMonitoringRuleResponseCreate,
		Read:   resourceDatadogSecurityMonitoringRuleResponseRead,
		Update: resourceDatadogSecurityMonitoringRuleResponseUpdate,
		Delete: resourceDatadogSecurityMonitoringRuleResponseDelete,
		// TODO: importer?
		Schema: map[string]*schema.Schema{
			"cases": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Cases for generating signals.",
			},

			"createdAt": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "When the rule was created, timestamp in milliseconds.",
			},

			"creationAuthorId": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "User ID of the user who created the rule.",
			},

			"id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the rule.",
			},

			"isDefault": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether the rule is included by default.",
			},

			"isDeleted": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether the rule has been deleted.",
			},

			"isEnabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether the rule is enabled.",
			},

			"message": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Message for generated signals.",
			},

			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the rule.",
			},

			"options": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "",
				MaxItems:    1,

				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"evaluationWindow": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Users resource type.",
						},

						"keepAlive": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Users resource type.",
						},

						"maxSignalDuration": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Users resource type.",
						},
					},
				},
			},

			"queries": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Queries for selecting logs which are part of the rule.",
			},

			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Tags for generated signals.",
			},

			"version": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The version of the rule.",
			},
		},
	}
}

func resourceDatadogSecurityMonitoringRuleResponseExists(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDatadogSecurityMonitoringRuleResponseCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDatadogSecurityMonitoringRuleResponseRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDatadogSecurityMonitoringRuleResponseUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDatadogSecurityMonitoringRuleResponseDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
