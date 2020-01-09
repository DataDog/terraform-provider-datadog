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
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

type LocationMap map[string]string

func dataSourceDatadogSyntheticsLocationsRead(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*datadog.Client)

	syntheticsLocations, err := client.GetSyntheticsLocations()

	if err != nil {
		return err
	}

	// Create a list of location names to be used in the data source
	var locationsList []string

	// Declare the list of maps
	var LocationMapSlice []LocationMap

	// Fill locationsList with the above map containing region, display name and name
	for _, location := range syntheticsLocations {
		// access the pointer of each struct element
		lm := LocationMap{"region": location.GetRegion(), "display_name": location.GetDisplayName(), "name": location.GetName()}
		LocationMapSlice = append(LocationMapSlice, lm)
	}

	for _, location := range LocationMapSlice {
		// In order to create a Synthetics test, we only need a list of location "name"
		locationsList = append(locationsList, location["name"])
	}

	if len(syntheticsLocations) > 0 {
		d.SetId("datadog-synthetics-location")
		d.Set("locations", locationsList)
	} else {
		d.Set("locations", []string{})
	}

	return nil
}
