package datadog

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceDatadogDowntime() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogDowntimeCreate,
		Read:   resourceDatadogDowntimeRead,
		Update: resourceDatadogDowntimeUpdate,
		Delete: resourceDatadogDowntimeDelete,
		Exists: resourceDatadogDowntimeExists,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogDowntimeImport,
		},

		Schema: map[string]*schema.Schema{
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				DiffSuppressFunc: func(k, oldVal, newVal string, d *schema.ResourceData) bool {
					_, recurrencePresent := d.GetOk("recurrence")
					return recurrencePresent
				},
				Description: "When true indicates this downtime is being actively applied",
			},
			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "When true indicates this downtime is not being applied",
			},
			"start": {
				Type:     schema.TypeInt,
				Optional: true,
				DiffSuppressFunc: func(k, oldVal, newVal string, d *schema.ResourceData) bool {
					_, startDatePresent := d.GetOk("start_date")
					return startDatePresent
				},
				Description: "Specify when this downtime should start",
			},
			"start_date": {
				Type:          schema.TypeString,
				ValidateFunc:  validation.IsRFC3339Time,
				ConflictsWith: []string{"start"},
				Optional:      true,
			},
			"end": {
				Type:     schema.TypeInt,
				Optional: true,
				DiffSuppressFunc: func(k, oldVal, newVal string, d *schema.ResourceData) bool {
					_, endDatePresent := d.GetOk("end_date")
					return endDatePresent
				},
				Description: "Optionally specify an end date when this downtime should expire",
			},
			"end_date": {
				Type:          schema.TypeString,
				ValidateFunc:  validation.IsRFC3339Time,
				ConflictsWith: []string{"end"},
				Optional:      true,
			},
			"timezone": {
				Type:         schema.TypeString,
				Default:      "UTC",
				Optional:     true,
				Description:  "The timezone for the downtime, default UTC",
				ValidateFunc: validateDatadogDowntimeTimezone,
			},
			"message": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An optional message to provide when creating the downtime, can include notification handles",
				StateFunc: func(val interface{}) string {
					return strings.TrimSpace(val.(string))
				},
			},
			"recurrence": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Optional recurring schedule for this downtime",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"period": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateDatadogDowntimeRecurrenceType,
						},
						"until_date": {
							Type:          schema.TypeInt,
							Optional:      true,
							ConflictsWith: []string{"recurrence.until_occurrences"},
						},
						"until_occurrences": {
							Type:          schema.TypeInt,
							Optional:      true,
							ConflictsWith: []string{"recurrence.until_date"},
						},
						"week_days": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateDatadogDowntimeRecurrenceWeekDays,
							},
						},
					},
				},
			},
			"scope": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "specify the group scope to which this downtime applies. For everything use '*'",
			},
			"monitor_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"monitor_tags"},
				Description:   "When specified, this downtime will only apply to this monitor",
			},
			"monitor_tags": {
				// we use TypeSet to represent tags, paradoxically to be able to maintain them ordered;
				// we order them explicitly in the read/create/update methods of this resource and using
				// TypeSet makes Terraform ignore differences in order when creating a plan
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A list of monitor tags (up to 25), i.e. tags that are applied directly to monitors to which the downtime applies",
				// MonitorTags conflicts with MonitorId and it also has a default of `["*"]`, which brings some problems:
				// * We can't use DefaultFunc to default to ["*"], since that's incompatible with
				//   ConflictsWith
				// * Since this is a TypeList, DiffSuppressFunc can't really be written well for it
				//   (it is called and expected to give result for each element, not for the whole
				//    list, so there's no way to tell in each iteration whether the new config value
				//    is an empty list).
				// Therefore we handle the "default" manually in resourceDatadogDowntimeRead function
				ConflictsWith: []string{"monitor_id"},
				Elem:          &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

// getDowntimeBoundaryTimestamp returns an int timestamp for start/end and the name of the
// attribute which it was extracted from (e.g. "end" or "end_date"). Arguments:
// * `d` - current `*schema.ResourceData`
// * `dateAttr` - name of the attribute in `d` which carries and RFC3339 date string, e.g. `end_date`
// * `tsAttr` - name of the attribute in `d` which carries integer timestamp, e.g. `end`
func getDowntimeBoundaryTimestamp(d *schema.ResourceData, dateAttr, tsAttr string) (ts int64, tsFrom string) {
	if attr, ok := d.GetOk(dateAttr); ok {
		if t, err := time.Parse(time.RFC3339, attr.(string)); err == nil {
			tsFrom = dateAttr
			ts = t.Unix()
		}
	} else if attr, ok := d.GetOk(tsAttr); ok {
		tsFrom = tsAttr
		ts = int64(attr.(int))
	}
	return ts, tsFrom
}

// downtimeBoundaryNeedsApply returns a boolean value signifying whether or not the boundary (start/end)
// should be included in the API POST/PUT request. Arguments:
// * `d` - current `*schema.ResourceData`
// * `tsFrom` - name of the attribute in `d` from which `configTs` was extracted
// * `apiTs` - current value (returned by API) of the boundary
// * `configTs` - desired value (from TF configuration) of the boundary
// * `updating` - `true` if this call is from Update method of the downtime resource, `false` if from Create
func downtimeBoundaryNeedsApply(d *schema.ResourceData, tsFrom string, apiTs, configTs int64, updating bool) (apply bool) {
	if tsFrom == "" {
		// if the boundary was not specified in the config, don't apply it
		return apply
	}

	if updating {
		// when updating, we apply when
		// * API-returned value is different than configured value
		// * the config value has changed
		if apiTs != configTs || d.HasChange(tsFrom) {
			apply = true
		}
	} else {
		// when creating, we always apply
		apply = true
	}

	return apply
}

func buildDowntimeStruct(authV1 context.Context, d *schema.ResourceData, client *datadogV1.APIClient, updating bool) (*datadogV1.Downtime, error) {
	// NOTE: for each of start/start_date/end/end_date, we only send the value when
	// it has changed or if the configured value is different than current value
	// (IOW there's a resource drift). This allows users to change other attributes
	// (e.g. scopes/message/...) without having to update the timestamps/dates to be
	// in the future (this works thanks to the downtime API allowing not to send these
	// values when they shouldn't be touched).
	var dt datadogV1.Downtime
	var currentStart = *datadogV1.PtrInt64(0)
	var currentEnd = *datadogV1.PtrInt64(0)
	if updating {
		id, err := strconv.ParseInt(d.Id(), 10, 64)
		if err != nil {
			return nil, err
		}

		var currdt datadogV1.Downtime
		currdt, _, err = client.DowntimesApi.GetDowntime(authV1, id).Execute()
		if err != nil {
			return nil, translateClientError(err, "error getting downtime")
		}
		currentStart = currdt.GetStart()
		currentEnd = currdt.GetEnd()
	}

	if attr, ok := d.GetOk("active"); ok {
		dt.SetActive(attr.(bool))
	}
	if attr, ok := d.GetOk("disabled"); ok {
		dt.SetDisabled(attr.(bool))
	}
	endValue, endAttrName := getDowntimeBoundaryTimestamp(d, "end_date", "end")
	if downtimeBoundaryNeedsApply(d, endAttrName, currentEnd, endValue, updating) {
		dt.SetEnd(endValue)
	}

	if attr, ok := d.GetOk("message"); ok {
		dt.SetMessage(strings.TrimSpace(attr.(string)))
	}
	if attr, ok := d.GetOk("monitor_id"); ok {
		dt.SetMonitorId(int64(attr.(int)))
	}
	if _, ok := d.GetOk("recurrence"); ok {
		var recurrence datadogV1.DowntimeRecurrence

		if attr, ok := d.GetOk("recurrence.0.period"); ok {
			recurrence.SetPeriod(int32(attr.(int)))
		}
		if attr, ok := d.GetOk("recurrence.0.type"); ok {
			recurrence.SetType(attr.(string))
		}
		if attr, ok := d.GetOk("recurrence.0.until_date"); ok {
			recurrence.SetUntilDate(int64(attr.(int)))
		}
		if attr, ok := d.GetOk("recurrence.0.until_occurrences"); ok {
			recurrence.SetUntilOccurrences(int32(attr.(int)))
		}
		if attr, ok := d.GetOk("recurrence.0.week_days"); ok {
			weekDays := make([]string, 0, len(attr.([]interface{})))
			for _, weekDay := range attr.([]interface{}) {
				weekDays = append(weekDays, weekDay.(string))
			}
			recurrence.SetWeekDays(weekDays)
		}

		dt.SetRecurrence(recurrence)
	}
	var scope []string
	for _, s := range d.Get("scope").([]interface{}) {
		scope = append(scope, s.(string))
	}
	dt.SetScope(scope)
	var tags []string
	for _, mt := range d.Get("monitor_tags").(*schema.Set).List() {
		tags = append(tags, mt.(string))
	}
	sort.Strings(tags)
	dt.SetMonitorTags(tags)

	startValue, startAttrName := getDowntimeBoundaryTimestamp(d, "start_date", "start")
	if downtimeBoundaryNeedsApply(d, startAttrName, currentStart, startValue, updating) {
		dt.SetStart(startValue)
	}

	if attr, ok := d.GetOk("timezone"); ok {
		dt.SetTimezone(attr.(string))
	}

	return &dt, nil
}

func resourceDatadogDowntimeExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return false, err
	}

	downtime, _, err := datadogClientV1.DowntimesApi.GetDowntime(authV1, id).Execute()
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, translateClientError(err, "error checking downtime exists")
	}

	if t, ok := downtime.GetCanceledOk(); ok && t != nil {
		// when the Downtime is deleted via UI, it is in fact still returned through API, it's just "canceled"
		// in this case, we need to consider it deleted, as canceled downtimes can't be used again
		return false, nil
	}

	return true, nil
}

func resourceDatadogDowntimeCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	dts, err := buildDowntimeStruct(authV1, d, datadogClientV1, false)
	if err != nil {
		return fmt.Errorf("failed to parse resource configuration: %s", err.Error())
	}
	dt, _, err := datadogClientV1.DowntimesApi.CreateDowntime(authV1).Body(*dts).Execute()
	if err != nil {
		return translateClientError(err, "error creating downtime")
	}

	d.SetId(strconv.Itoa(int(dt.GetId())))

	return resourceDatadogDowntimeRead(d, meta)
}

func resourceDatadogDowntimeRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	dt, _, err := datadogClientV1.DowntimesApi.GetDowntime(authV1, id).Execute()
	if err != nil {
		return translateClientError(err, "error getting downtime")
	}

	log.Printf("[DEBUG] downtime: %v", dt)

	if err := d.Set("active", dt.GetActive()); err != nil {
		return err
	}
	if err := d.Set("disabled", dt.GetDisabled()); err != nil {
		return err
	}
	if err := d.Set("end", dt.GetEnd()); err != nil {
		return err
	}
	if err := d.Set("message", dt.GetMessage()); err != nil {
		return err
	}
	if err := d.Set("monitor_id", dt.GetMonitorId()); err != nil {
		return err
	}

	if err := d.Set("timezone", dt.GetTimezone()); err != nil {
		return err
	}

	if r, ok := dt.GetRecurrenceOk(); ok && r != nil {
		recurrence := make(map[string]interface{})
		recurrenceList := make([]map[string]interface{}, 0, 1)

		if attr, ok := r.GetPeriodOk(); ok {
			recurrence["period"] = attr
		}
		if attr, ok := r.GetTypeOk(); ok {
			recurrence["type"] = attr
		}
		if attr, ok := r.GetUntilDateOk(); ok {
			recurrence["until_date"] = attr
		}
		if attr, ok := r.GetUntilOccurrencesOk(); ok {
			recurrence["until_occurrences"] = attr
		}
		if r.GetWeekDays() != nil {
			weekDays := make([]string, 0, len(r.GetWeekDays()))
			for _, weekDay := range *r.WeekDays {
				weekDays = append(weekDays, weekDay)
			}
			recurrence["week_days"] = weekDays
		}
		recurrenceList = append(recurrenceList, recurrence)
		d.Set("recurrence", recurrenceList)
	}
	d.Set("scope", dt.Scope)
	// See the comment for monitor_tags in the schema definition above
	if !reflect.DeepEqual(dt.GetMonitorTags(), []string{"*"}) {
		tags := dt.GetMonitorTags()
		sort.Strings(tags)
		d.Set("monitor_tags", tags)
	}
	d.Set("start", dt.GetStart())

	return nil
}

func resourceDatadogDowntimeUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	dt, err := buildDowntimeStruct(authV1, d, datadogClientV1, true)
	if err != nil {
		return fmt.Errorf("failed to parse resource configuration: %s", err.Error())
	}
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}
	// above downtimeStruct returns nil if downtime is not set. Hence, if we are handling the cases where downtime
	// is replaced, the ID of the downtime will be set to 0.
	dt.SetId(id)

	if _, _, err = datadogClientV1.DowntimesApi.UpdateDowntime(authV1, id).Body(*dt).Execute(); err != nil {
		return translateClientError(err, "error updating downtime")
	}
	// handle the case when a downtime is replaced
	d.SetId(strconv.FormatInt(dt.GetId(), 10))

	return resourceDatadogDowntimeRead(d, meta)
}

func resourceDatadogDowntimeDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	if _, err = datadogClientV1.DowntimesApi.CancelDowntime(authV1, id).Execute(); err != nil {
		return translateClientError(err, "error deleting downtime")
	}

	return nil
}

func resourceDatadogDowntimeImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogDowntimeRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func validateDatadogDowntimeRecurrenceType(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	switch value {
	case "days", "months", "weeks", "years":
		break
	default:
		errors = append(errors, fmt.Errorf(
			"%q contains an invalid recurrence type parameter %q. Valid parameters are days, months, weeks, or years", k, value))
	}
	return
}

func validateDatadogDowntimeRecurrenceWeekDays(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	switch value {
	case "Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun":
		break
	default:
		errors = append(errors, fmt.Errorf(
			"%q contains an invalid recurrence week day parameter %q. Valid parameters are Mon, Tue, Wed, Thu, Fri, Sat, or Sun", k, value))
	}
	return
}

func validateDatadogDowntimeTimezone(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	switch strings.ToLower(value) {
	case "utc", "":
		break
	case "local", "localtime":
		// get current zone from machine
		zone, _ := time.Now().Local().Zone()
		return validateDatadogDowntimeRecurrenceType(zone, k)
	default:
		_, err := time.LoadLocation(value)
		if err != nil {
			errors = append(errors, fmt.Errorf(
				"%q contains an invalid timezone parameter: %q, Valid parameters are IANA Time Zone names",
				k, value))
		}
	}
	return
}
