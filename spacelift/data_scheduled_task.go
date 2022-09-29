package spacelift

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataScheduledTask() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_scheduled_task` represents a scheduling configuration " +
			"for a Stack. It will trigger a task on the given timestamp/schedule.",

		ReadContext: dataScheduledTaskRead,

		Schema: map[string]*schema.Schema{
			"scheduled_task_id": {
				Type:        schema.TypeString,
				Description: "ID of the scheduled task (stack_id/schedule_id)",
				Required:    true,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "Stack ID of the scheduled task",
				Computed:    true,
			},
			"schedule_id": {
				Type:        schema.TypeString,
				Description: "ID of the schedule",
				Computed:    true,
			},
			"at": {
				Type:        schema.TypeInt,
				Description: "Timestamp (unix timestamp) at which time the scheduling should happen.",
				Computed:    true,
			},
			"every": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of cron schedule expressions based on which the scheduled task should be triggered.",
				Computed:    true,
			},
			"command": {
				Type:        schema.TypeString,
				Description: "Command that will be run.",
				Computed:    true,
			},
			"timezone": {
				Type:        schema.TypeString,
				Description: "Timezone in which the schedule is expressed. Defaults to `UTC`.",
				Computed:    true,
			},
		},
	}
}

func dataScheduledTaskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scheduledTaskID := d.Get("scheduled_task_id").(string)

	idParts := strings.SplitN(scheduledTaskID, "/", 2)
	if len(idParts) != 2 {
		return diag.Errorf("unexpected resource ID: %s", d.Id())
	}

	stackID, scheduleID := idParts[0], idParts[1]

	var err error

	if err = d.Set("stack_id", stackID); err != nil {
		return diag.Errorf("could not set stack id")
	}

	if err = d.Set("schedule_id", scheduleID); err != nil {
		return diag.Errorf("could not set schedule id")
	}

	var query struct {
		Stack *struct {
			ScheduledTask *structs.ScheduledTask `graphql:"scheduledTask(id: $id)"`
		} `graphql:"stack(id: $stack)"`
	}

	variables := map[string]interface{}{"stack": toID(stackID), "id": toID(scheduleID)}

	if err := meta.(*internal.Client).Query(ctx, "StackScheduledTaskRead", &query, variables); err != nil {
		return diag.Errorf("could not query for scheduled_task: %v", internal.FromSpaceliftError(err))
	}

	if query.Stack == nil || query.Stack.ScheduledTask == nil {
		return diag.Errorf("could not find scheduled task: %s", scheduledTaskID)
	}

	if err := structs.PopulateTaskSchedule(d, query.Stack.ScheduledTask); err != nil {
		return diag.Errorf("could not populate scheduled `task` config: %v", err)
	}

	d.SetId(scheduledTaskID)

	return nil
}
