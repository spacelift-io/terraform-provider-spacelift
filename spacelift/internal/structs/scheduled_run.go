package structs

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"
)

// ScheduledRun represents a scheduled run.
type ScheduledRun struct {
	ID                  string           `json:"id"`
	Name                graphql.String   `json:"name"`
	CronSchedule        []graphql.String `json:"cronSchedule"`
	TimestampSchedule   *graphql.Float   `json:"timestampSchedule"`
	NextSchedule        *graphql.Int     `json:"nextSchedule"`
	Timezone            *graphql.String  `json:"timezone"`
	CustomRuntimeConfig *RuntimeConfig   `json:"customRuntimeConfig"`
}

// ScheduledRunInput represents the input for creating/updating a scheduled run.
type ScheduledRunInput struct {
	Name              graphql.String      `json:"name"`
	CronSchedule      []graphql.String    `json:"cronSchedule"`
	TimestampSchedule *graphql.Int        `json:"timestampSchedule"`
	Timezone          *graphql.String     `json:"timezone"`
	RuntimeConfig     *RuntimeConfigInput `json:"runtimeConfig"`
}

// RuntimeConfig represents a runtime configuration.
type RuntimeConfig struct {
	YamlConfig graphql.String `json:"yamlConfig"`
}

// RuntimeConfigInput represents the input for runtime configuration.
type RuntimeConfigInput struct {
	YamlConfig *graphql.String `json:"yamlConfig"`
}

func PopulateRunSchedule(d *schema.ResourceData, r *ScheduledRun) error {
	if err := d.Set("name", r.Name); err != nil {
		return errors.Wrap(err, "could not set \"name\" attribute")
	}

	if len(r.CronSchedule) != 0 {
		if err := d.Set("every", r.CronSchedule); err != nil {
			return errors.Wrap(err, "could not set \"every\" attribute")
		}

		if err := d.Set("timezone", r.Timezone); err != nil {
			return errors.Wrap(err, "could not set \"timezone\" attribute")
		}
	} else if r.TimestampSchedule != nil {
		if err := d.Set("at", int(*r.TimestampSchedule)); err != nil {
			return errors.Wrap(err, "could not set \"at\" attribute")
		}
	}

	if r.NextSchedule != nil {
		if err := d.Set("next_schedule", int(*r.NextSchedule)); err != nil {
			return errors.Wrap(err, "could not set \"next_schedule\" attribute")
		}
	}

	if r.CustomRuntimeConfig != nil {
		if err := d.Set("runtime_config", r.CustomRuntimeConfig.YamlConfig); err != nil {
			return errors.Wrap(err, "could not set \"runtime_config\" attribute")
		}
	}

	return nil
}
