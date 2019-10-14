package datadog

import (
	"github.com/hashicorp/terraform/helper/schema"
	datadog "github.com/zorkian/go-datadog-api"
)

func dataSourceDatadogSyntheticsLocations() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDatadogSyntheticsLocationsRead,

		// Locations are a list of string
		Schema: map[string]*schema.Schema{
			"locations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceDatadogSyntheticsLocationsRead(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*datadog.Client)

	syntheticsLocations, err := client.GetSyntheticsLocations()

	if err != nil {
		return err
	}

	// Create a list of location names to be used in the data source
	var locationsList []string

	// Fill locationsList with location names
	for _, location := range syntheticsLocations {
		// access the pointer of the struct element DisplayName
		locationsList = append(locationsList, location.GetDisplayName())
	}

	if len(syntheticsLocations) > 0 {
		d.SetId("datadog-synthetics-location")
		d.Set("locations", locationsList)
	} else {
		d.Set("locations", []string{})
	}

	return nil
}
