package spacelift

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

type ScheduledTask struct {
	At       *int
	Command  string
	Every    []string
	Timezone string
}

func resourceScheduledTask() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_scheduled_task` represents a scheduling configuration " +
			"for a Stack. It will trigger task on the given schedule or timestamp",

		CreateContext: resourceScheduledTaskCreate,
		ReadContext:   resourceScheduledTaskRead,
		UpdateContext: resourceScheduledTaskUpdate,
		DeleteContext: resourceScheduledTaskDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack for which to set up the scheduled task",
				Required:    true,
				ForceNew:    true,
			},
			"schedule_id": {
				Type:        schema.TypeString,
				Description: "ID of the schedule",
				Optional:    true,
				Computed:    true,
			},
			"command": {
				Type:             schema.TypeString,
				Description:      "Command that will be run.",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"at": {
				Type:             schema.TypeInt,
				Description:      "Timestamp (unix timestamp) at which time the scheduled task should happen.",
				Optional:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
				ConflictsWith:    []string{"every", "timezone"},
			},
			"every": {
				Type:          schema.TypeList,
				Elem:          &schema.Schema{Type: schema.TypeString},
				MinItems:      1,
				Description:   "List of cron schedule expressions based on which the scheduled task should be triggered.",
				Optional:      true,
				ConflictsWith: []string{"at"},
			},
			"timezone": {
				Type:          schema.TypeString,
				Description:   "Timezone in which the schedule is expressed. Defaults to `UTC`.",
				Optional:      true,
				Default:       "UTC",
				ConflictsWith: []string{"at"},
			},
		},
	}
}

func resourceScheduledTaskCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scheduledTask, err := getScheduledTask(d)
	if err != nil {
		return diag.Errorf("could not extract scheduled task: %s", err)
	}

	var mutation struct {
		CreateTaskSchedule structs.ScheduledTask `graphql:"stackScheduledTaskCreate(stack: $stack, input: $input)"`
	}

	input := structs.ScheduledTaskInput{
		Command: graphql.String(scheduledTask.Command),
	}

	if scheduledTask.At != nil {
		input.TimestampSchedule = graphql.NewInt(graphql.Int(*scheduledTask.At))
	} else {
		var scheduleExpressions []graphql.String
		for _, expr := range scheduledTask.Every {
			scheduleExpressions = append(scheduleExpressions, graphql.String(expr))
		}

		input.CronSchedule = scheduleExpressions
		input.Timezone = graphql.NewString(graphql.String(scheduledTask.Timezone))
	}

	variables := map[string]interface{}{
		"stack": toID(d.Get("stack_id").(string)),
		"input": input,
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ScheduledDeleteCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create scheduled task: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(fmt.Sprintf("%s/%s", d.Get("stack_id"), mutation.CreateTaskSchedule.ID))

	return resourceScheduledTaskRead(ctx, d, meta)
}

func resourceScheduledTaskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
			ScheduledTask *structs.ScheduledTask `graphql:"scheduledTask(id: $id)"`
		} `graphql:"stack(id: $stack)"`
	}

	variables := map[string]interface{}{"stack": toID(stackID), "id": toID(scheduleID)}

	if err := meta.(*internal.Client).Query(ctx, "StackScheduledTaskRead", &query, variables); err != nil {
		return diag.Errorf("could not query for scheduled `task` config: %v", internal.FromSpaceliftError(err))
	}

	if query.Stack == nil || query.Stack.ScheduledTask == nil {
		d.SetId("")
		return nil
	}

	if err := structs.PopulateTaskSchedule(d, query.Stack.ScheduledTask); err != nil {
		return diag.Errorf("could not populate scheduled `task` config: %v", err)
	}

	return nil
}

func resourceScheduledTaskUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	scheduledTask, err := getScheduledTask(d)
	if err != nil {
		return diag.Errorf("could not extract scheduled task: %s", err)
	}

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

	var mutation struct {
		UpdateTaskSchedule structs.ScheduledTask `graphql:"stackScheduledTaskUpdate(stack: $stack, scheduledTask: $scheduledTask, input: $input)"`
	}

	input := structs.ScheduledTaskInput{
		Command: graphql.String(scheduledTask.Command),
	}

	if scheduledTask.At != nil {
		input.TimestampSchedule = graphql.NewInt(graphql.Int(*scheduledTask.At))
	} else {
		var scheduleExpressions []graphql.String
		for _, expr := range scheduledTask.Every {
			scheduleExpressions = append(scheduleExpressions, graphql.String(expr))
		}

		input.CronSchedule = scheduleExpressions
		input.Timezone = graphql.NewString(graphql.String(scheduledTask.Timezone))
	}

	variables := map[string]interface{}{
		"stack":         toID(stackID),
		"scheduledTask": toID(scheduleID),
		"input":         input,
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ScheduledTaskUpdate", &mutation, variables); err != nil {
		return diag.Errorf("could not update scheduled task: %v", internal.FromSpaceliftError(err))
	}

	return resourceScheduledTaskRead(ctx, d, meta)
}

func resourceScheduledTaskDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error

	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 {
		return diag.Errorf("unexpected resource ID: %s", d.Id())
	}

	stackID, scheduleID := idParts[0], idParts[1]

	if err = d.Set("stack_id", stackID); err != nil {
		return diag.Errorf("could not set stack id")
	}

	var mutation struct {
		DeleteTaskSchedule structs.ScheduledTask `graphql:"stackScheduledTaskDelete(stack: $stack, scheduledTask: $scheduledTask)"`
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ScheduledTaskDelete", &mutation, map[string]interface{}{
		"stack":         toID(stackID),
		"scheduledTask": toID(scheduleID),
	}); err != nil {
		return diag.Errorf("could not delete scheduled `task` config: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func getScheduledTask(d *schema.ResourceData) (*ScheduledTask, error) {
	cfg := &ScheduledTask{}

	command, ok := d.GetOk("command")
	if ok && command != nil {
		cfg.Command = command.(string)
	}

	every, everyOk := d.GetOk("every")
	at, atOk := d.GetOk("at")

	if everyOk && every != nil {
		v := every.([]interface{})
		if len(v) > 0 {
			var scheduleExpressions []string
			for _, expr := range v {
				scheduleExpressions = append(scheduleExpressions, expr.(string))
			}
			cfg.Every = scheduleExpressions
		}

		timezone, ok := d.GetOk("timezone")
		if ok && timezone != nil {
			cfg.Timezone = timezone.(string)
		}

	} else if atOk && at != nil {
		a := at.(int)
		cfg.At = &a
	} else {
		return nil, errors.New("Either `at` or `every` attribute is required")
	}

	return cfg, nil
}
