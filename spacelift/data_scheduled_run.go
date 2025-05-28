package spacelift

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataScheduledRun() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_scheduled_run` represents a scheduling configuration " +
			"for a Stack. It will trigger a run on the given timestamp/schedule.",

		ReadContext: dataScheduledRunRead,

		Schema: map[string]*schema.Schema{
			"scheduled_run_id": {
				Type:        schema.TypeString,
				Description: "ID of the scheduled run (stack_id/schedule_id)",
				Required:    true,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "Stack ID of the scheduled run",
				Computed:    true,
			},
			"schedule_id": {
				Type:        schema.TypeString,
				Description: "ID of the schedule",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the scheduled run",
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
				Description: "List of cron schedule expressions based on which the scheduled run should be triggered.",
				Computed:    true,
			},
			"timezone": {
				Type:        schema.TypeString,
				Description: "Timezone in which the schedule is expressed. Defaults to `UTC`.",
				Computed:    true,
			},
			"next_schedule": {
				Type:        schema.TypeInt,
				Description: "Timestamp (unix timestamp) of when the next run will be scheduled.",
				Computed:    true,
			},
			"runtime_config": {
				Type:        schema.TypeString,
				Description: "Customer provided runtime configuration for this scheduled run.",
				Computed:    true,
			},
		},
	}
}

func dataScheduledRunRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scheduledRunID := d.Get("scheduled_run_id").(string)

	idParts := strings.SplitN(scheduledRunID, "/", 2)
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
			ScheduledRun *structs.ScheduledRun `graphql:"scheduledRun(id: $id)"`
		} `graphql:"stack(id: $stack)"`
	}

	variables := map[string]interface{}{"stack": toID(stackID), "id": toID(scheduleID)}

	if err := meta.(*internal.Client).Query(ctx, "StackScheduledRunRead", &query, variables); err != nil {
		return diag.Errorf("could not query for scheduled_run: %v", internal.FromSpaceliftError(err))
	}

	if query.Stack == nil || query.Stack.ScheduledRun == nil {
		return diag.Errorf("could not find scheduled run: %s", scheduledRunID)
	}

	if err := structs.PopulateRunSchedule(d, query.Stack.ScheduledRun); err != nil {
		return diag.Errorf("could not populate scheduled run config: %v", err)
	}

	d.SetId(scheduledRunID)

	return nil
}
