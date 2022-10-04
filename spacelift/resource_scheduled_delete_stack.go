package spacelift

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceScheduledDeleteStack() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_scheduled_delete_stack` represents a scheduling configuration " +
			"for a Stack. It will trigger a stack deletion task at the given timestamp.",

		CreateContext: resourceScheduledDeleteStackCreate,
		ReadContext:   resourceScheduledDeleteStackRead,
		UpdateContext: resourceScheduledDeleteStackUpdate,
		DeleteContext: resourceScheduledDeleteStackDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack for which to set up scheduling",
				Required:    true,
				ForceNew:    true,
			},
			"schedule_id": {
				Type:        schema.TypeString,
				Description: "ID of the schedule",
				Optional:    true,
				Computed:    true,
			},
			"at": {
				Type:             schema.TypeInt,
				Description:      "Timestamp (unix timestamp) at which time the scheduling should happen.",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"delete_resources": {
				Type:        schema.TypeBool,
				Description: "Indicates whether the resources of the stack should be deleted.",
				Default:     true,
				Optional:    true,
			},
		},
	}
}

func resourceScheduledDeleteStackCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	at := d.Get("at").(int)
	deleteResources := d.Get("delete_resources").(bool)

	var mutation struct {
		CreateDeleteSchedule structs.ScheduledStackDelete `graphql:"stackScheduledDeleteCreate(stack: $stack, input: $input)"`
	}

	variables := map[string]interface{}{
		"stack": toID(d.Get("stack_id").(string)),
		"input": structs.ScheduledDeleteInput{
			ShouldDeleteResources: graphql.Boolean(deleteResources),
			TimestampSchedule:     graphql.NewInt(graphql.Int(at)),
		},
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ScheduledDeleteCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create scheduled delete: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(fmt.Sprintf("%s/%s", d.Get("stack_id"), mutation.CreateDeleteSchedule.ID))

	return resourceScheduledDeleteStackRead(ctx, d, meta)
}

func resourceScheduledDeleteStackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error

	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 {
		return diag.Errorf("unexpected resource ID: %s", d.Id())
	}

	stackID, scheduleID := idParts[0], idParts[1]

	if err = d.Set("stack_id", stackID); err != nil {
		return diag.Errorf("could not set \"stack_id\"")
	}

	if err = d.Set("schedule_id", scheduleID); err != nil {
		return diag.Errorf("could not set \"schedule_id\"")
	}

	var query struct {
		Stack *struct {
			ScheduledDelete *structs.ScheduledStackDelete `graphql:"scheduledDelete(id: $id)"`
		} `graphql:"stack(id: $stack)"`
	}

	if err := meta.(*internal.Client).Query(ctx, "StackSchedulingRead", &query, map[string]interface{}{"stack": toID(stackID), "id": toID(scheduleID)}); err != nil {
		return diag.Errorf("could not query for scheduled stack_delete: %v", internal.FromSpaceliftError(err))
	}

	if query.Stack == nil || query.Stack.ScheduledDelete == nil {
		return nil
	}

	if err := structs.PopulateDeleteStackSchedule(d, query.Stack.ScheduledDelete); err != nil {
		return diag.Errorf("could not populate scheduled stack_delete: %v", err)
	}

	return nil
}

func resourceScheduledDeleteStackUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error

	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 {
		return diag.Errorf("unexpected resource ID: %s", d.Id())
	}

	stackID, scheduleID := idParts[0], idParts[1]

	if err := d.Set("stack_id", stackID); err != nil {
		return diag.Errorf("could not set \"stack_id\"")
	}
	if err = d.Set("schedule_id", scheduleID); err != nil {
		return diag.Errorf("could not set \"schedule_id\"")
	}
	at := d.Get("at").(int)
	deleteResources := d.Get("delete_resources").(bool)

	var mutation struct {
		UpdateDeleteSchedule structs.ScheduledStackDelete `graphql:"stackScheduledDeleteUpdate(stack: $stack, scheduledDelete: $scheduledDelete, input: $input)"`
	}

	variables := map[string]interface{}{
		"stack":           toID(stackID),
		"scheduledDelete": toID(scheduleID),
		"input": structs.ScheduledDeleteInput{
			ShouldDeleteResources: graphql.Boolean(deleteResources),
			TimestampSchedule:     graphql.NewInt(graphql.Int(at)),
		},
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ScheduledDeleteUpdate", &mutation, variables); err != nil {
		return diag.Errorf("could not create scheduled delete: %v", internal.FromSpaceliftError(err))
	}

	return resourceScheduledDeleteStackRead(ctx, d, meta)
}

func resourceScheduledDeleteStackDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 {
		return diag.Errorf("unexpected resource ID: %s", d.Id())
	}

	stackID, scheduleID := idParts[0], idParts[1]

	var mutation struct {
		DeleteDeleteStackSchedule structs.ScheduledStackDelete `graphql:"stackScheduledDeleteDelete(stack: $stack, scheduledDelete: $scheduledDelete)"`
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ScheduledStackDeleteDelete", &mutation, map[string]interface{}{
		"stack":           toID(stackID),
		"scheduledDelete": toID(scheduleID),
	}); err != nil {
		return diag.Errorf("could not delete scheduled stack_delete config: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}
