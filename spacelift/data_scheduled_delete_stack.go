package spacelift

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataScheduledDeleteStack() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_scheduled_delete_stack` represents a scheduling configuration " +
			"for a Stack. It will trigger a stack deletion task at the given timestamp.",

		ReadContext: dataScheduledDeleteStackRead,

		Schema: map[string]*schema.Schema{
			"scheduled_delete_stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the scheduled delete_stack (stack_id/schedule_id)",
				Required:    true,
			},
			"schedule_id": {
				Type:        schema.TypeString,
				Description: "ID of the schedule",
				Computed:    true,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "Stack ID of the scheduling config",
				Computed:    true,
			},
			"at": {
				Type:        schema.TypeInt,
				Description: "Timestamp (unix timestamp) at which time the scheduling should happen.",
				Computed:    true,
			},
			"delete_resources": {
				Type:        schema.TypeBool,
				Description: "Indicates whether the resources of the stack should be deleted.",
				Computed:    true,
			},
		},
	}
}

func dataScheduledDeleteStackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scheduledDeleteStackID := d.Get("scheduled_delete_stack_id").(string)

	idParts := strings.SplitN(scheduledDeleteStackID, "/", 2)
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
			ScheduledDelete *structs.ScheduledStackDelete `graphql:"scheduledDelete(id: $id)"`
		} `graphql:"stack(id: $stack)"`
	}

	variables := map[string]interface{}{"stack": toID(stackID), "id": toID(scheduleID)}

	if err := meta.(*internal.Client).Query(ctx, "StackScheduledDeleteStackRead", &query, variables); err != nil {
		return diag.Errorf("could not query for scheduled stack_delete: %v", internal.FromSpaceliftError(err))
	}

	if query.Stack == nil || query.Stack.ScheduledDelete == nil {
		return diag.Errorf("could not find scheduled delete_stack")
	}

	if err := structs.PopulateDeleteStackSchedule(d, query.Stack.ScheduledDelete); err != nil {
		return diag.Errorf("could not populate scheduled delete stack: %v", err)
	}

	d.SetId(scheduledDeleteStackID)

	return nil
}
