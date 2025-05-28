package structs

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"
)

// ScheduledRun represents a scheduled run.
type ScheduledRun struct {
	ID                string         `graphql:"id"`
	Name              string         `graphql:"name"`
	CronSchedule      []string       `graphql:"cronSchedule"`
	TimestampSchedule int            `graphql:"timestampSchedule"`
	NextSchedule      *int           `graphql:"nextSchedule"`
	Timezone          *string        `graphql:"timezone"`
	RuntimeConfig     *RuntimeConfig `graphql:"customRuntimeConfig"`
}

// ScheduledRunInput represents the input for creating/updating a scheduled run.
type ScheduledRunInput struct {
	Name              graphql.String      `json:"name"`
	CronSchedule      []graphql.String    `json:"cronSchedule"`
	TimestampSchedule *graphql.Int        `json:"timestampSchedule"`
	Timezone          *graphql.String     `json:"timezone"`
	RuntimeConfig     *RuntimeConfigInput `json:"runtimeConfig"`
}

func PopulateRunSchedule(d *schema.ResourceData, r *ScheduledRun) diag.Diagnostics {
	if err := d.Set("name", r.Name); err != nil {
		return diag.Errorf("could not set \"name\" attribute")
	}

	if len(r.CronSchedule) != 0 {
		if err := d.Set("every", r.CronSchedule); err != nil {
			return diag.Errorf("could not set \"every\" attribute")
		}

		if err := d.Set("timezone", r.Timezone); err != nil {
			return diag.Errorf("could not set \"timezone\" attribute")
		}
	} else {
		if err := d.Set("at", r.TimestampSchedule); err != nil {
			return diag.Errorf("could not set \"at\" attribute")
		}
	}

	if r.NextSchedule != nil {
		if err := d.Set("next_schedule", r.NextSchedule); err != nil {
			return diag.Errorf("could not set \"next_schedule\" attribute")
		}
	}

	if r.RuntimeConfig != nil {
		runtimeConfig, err := ExportRuntimeConfigToMap(r.RuntimeConfig)
		if err != nil {
			return err
		}
		if err := d.Set("runtime_config", []interface{}{runtimeConfig}); err != nil {
			return diag.Errorf("could not set \"runtime_config\" attribute")
		}
	}

	return nil
}
