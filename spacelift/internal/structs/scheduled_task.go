package structs

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"
)

// ScheduledTask represents a scheduled task.
type ScheduledTask struct {
	ID                string           `json:"id"`
	Command           graphql.String   `json:"command"`
	CronSchedule      []graphql.String `json:"cronSchedule"`
	TimestampSchedule graphql.Int      `json:"timestampSchedule"`
	Timezone          *graphql.String  `json:"timezone"`
}

// ScheduledTaskInput represents the input for creating/updating a scheduled task.
type ScheduledTaskInput struct {
	Command           graphql.String   `json:"command"`
	CronSchedule      []graphql.String `json:"cronSchedule"`
	TimestampSchedule *graphql.Int     `json:"timestampSchedule"`
	Timezone          *graphql.String  `json:"timezone"`
}

func PopulateTaskSchedule(d *schema.ResourceData, t *ScheduledTask) error {
	if err := d.Set("command", t.Command); err != nil {
		return errors.Wrap(err, "could not set \"command\" attribute")
	}

	if len(t.CronSchedule) != 0 {
		if err := d.Set("every", t.CronSchedule); err != nil {
			return errors.Wrap(err, "could not set \"every\" attribute")
		}

		if err := d.Set("timezone", t.Timezone); err != nil {
			return errors.Wrap(err, "could not set \"timezone\" attribute")
		}
	} else {
		if err := d.Set("at", t.TimestampSchedule); err != nil {
			return errors.Wrap(err, "could not set \"at\" attribute")
		}
	}

	return nil
}
