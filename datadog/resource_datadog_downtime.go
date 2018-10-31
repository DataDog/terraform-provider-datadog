package datadog

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	datadog "github.com/zorkian/go-datadog-api"
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
			},
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"start": {
				Type:     schema.TypeInt,
				Optional: true,
				DiffSuppressFunc: func(k, oldVal, newVal string, d *schema.ResourceData) bool {
					_, endDatePresent := d.GetOk("start_date")
					return endDatePresent
				},
			},
			"start_date": {
				Type:          schema.TypeString,
				ValidateFunc:  validation.ValidateRFC3339TimeString,
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
			},
			"end_date": {
				Type:          schema.TypeString,
				ValidateFunc:  validation.ValidateRFC3339TimeString,
				ConflictsWith: []string{"end"},
				Optional:      true,
			},
			"message": {
				Type:     schema.TypeString,
				Optional: true,
				StateFunc: func(val interface{}) string {
					return strings.TrimSpace(val.(string))
				},
			},
			"recurrence": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
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
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"monitor_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func buildDowntimeStruct(d *schema.ResourceData) *datadog.Downtime {
	var dt datadog.Downtime

	if attr, ok := d.GetOk("active"); ok {
		dt.SetActive(attr.(bool))
	}
	if attr, ok := d.GetOk("disabled"); ok {
		dt.SetDisabled(attr.(bool))
	}
	if attr, ok := d.GetOk("end"); ok {
		dt.SetEnd(attr.(int))
	} else if attr, ok := d.GetOk("end_date"); ok {
		if t, err := time.Parse(time.RFC3339, attr.(string)); err == nil {
			dt.SetEnd(int(t.Unix()))
		}
	}
	if attr, ok := d.GetOk("message"); ok {
		dt.SetMessage(strings.TrimSpace(attr.(string)))
	}
	if attr, ok := d.GetOk("monitor_id"); ok {
		dt.SetMonitorId(attr.(int))
	}
	if _, ok := d.GetOk("recurrence"); ok {
		var recurrence datadog.Recurrence

		if attr, ok := d.GetOk("recurrence.0.period"); ok {
			recurrence.SetPeriod(attr.(int))
		}
		if attr, ok := d.GetOk("recurrence.0.type"); ok {
			recurrence.SetType(attr.(string))
		}
		if attr, ok := d.GetOk("recurrence.0.until_date"); ok {
			recurrence.SetUntilDate(attr.(int))
		}
		if attr, ok := d.GetOk("recurrence.0.until_occurrences"); ok {
			recurrence.SetUntilOccurrences(attr.(int))
		}
		if attr, ok := d.GetOk("recurrence.0.week_days"); ok {
			weekDays := make([]string, 0, len(attr.([]interface{})))
			for _, weekDay := range attr.([]interface{}) {
				weekDays = append(weekDays, weekDay.(string))
			}
			recurrence.WeekDays = weekDays
		}

		dt.SetRecurrence(recurrence)
	}
	scope := []string{}
	for _, s := range d.Get("scope").([]interface{}) {
		scope = append(scope, s.(string))
	}
	dt.Scope = scope
	if attr, ok := d.GetOk("start"); ok {
		dt.SetStart(attr.(int))
	} else if attr, ok := d.GetOk("start_date"); ok {
		if t, err := time.Parse(time.RFC3339, attr.(string)); err == nil {
			dt.SetStart(int(t.Unix()))
		}
	}

	return &dt
}

func resourceDatadogDowntimeExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	client := meta.(*datadog.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, err
	}

	if _, err = client.GetDowntime(id); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func resourceDatadogDowntimeCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	dts := buildDowntimeStruct(d)
	dt, err := client.CreateDowntime(dts)
	if err != nil {
		return fmt.Errorf("error updating downtime: %s", err.Error())
	}

	d.SetId(strconv.Itoa(dt.GetId()))

	return nil
}

func resourceDatadogDowntimeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	dt, err := client.GetDowntime(id)
	if err != nil {
		return err
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

	if r, ok := dt.GetRecurrenceOk(); ok {
		recurrence := make(map[string]interface{})
		recurrenceList := make([]map[string]interface{}, 0, 1)

		if attr, ok := r.GetPeriodOk(); ok {
			recurrence["period"] = strconv.Itoa(attr)
		}
		if attr, ok := r.GetTypeOk(); ok {
			recurrence["type"] = attr
		}
		if attr, ok := r.GetUntilDateOk(); ok {
			recurrence["until_date"] = strconv.Itoa(attr)
		}
		if attr, ok := r.GetUntilOccurrencesOk(); ok {
			recurrence["until_occurrences"] = strconv.Itoa(attr)
		}
		if r.WeekDays != nil {
			weekDays := make([]string, 0, len(r.WeekDays))
			for _, weekDay := range r.WeekDays {
				weekDays = append(weekDays, weekDay)
			}
			recurrence["week_days"] = weekDays
		}
		recurrenceList = append(recurrenceList, recurrence)
		d.Set("recurrence", recurrenceList)
	}
	d.Set("scope", dt.Scope)
	d.Set("start", dt.GetStart())

	return nil
}

func resourceDatadogDowntimeUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	dt := buildDowntimeStruct(d)
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}
	dt.SetId(id)

	if err = client.UpdateDowntime(dt); err != nil {
		return fmt.Errorf("error updating downtime: %s", err.Error())
	}

	return resourceDatadogDowntimeRead(d, meta)
}

func resourceDatadogDowntimeDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	if err = client.DeleteDowntime(id); err != nil {
		return err
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
