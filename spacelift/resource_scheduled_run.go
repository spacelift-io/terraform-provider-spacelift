package spacelift

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceScheduledRun() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_scheduled_run` represents a scheduling configuration " +
			"for a Stack. It will trigger a run on the given schedule or timestamp",

		CreateContext: resourceScheduledRunCreate,
		ReadContext:   resourceScheduledRunRead,
		UpdateContext: resourceScheduledRunUpdate,
		DeleteContext: resourceScheduledRunDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack for which to set up the scheduled run",
				Required:    true,
				ForceNew:    true,
			},
			"schedule_id": {
				Type:        schema.TypeString,
				Description: "ID of the schedule",
				Optional:    true,
				Computed:    true,
			},
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the scheduled run",
				Optional:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"at": {
				Type:          schema.TypeInt,
				Description:   "Timestamp (unix timestamp) at which time the scheduled run should happen.",
				Optional:      true,
				ConflictsWith: []string{"every", "timezone"},
			},
			"every": {
				Type:          schema.TypeList,
				Elem:          &schema.Schema{Type: schema.TypeString},
				MinItems:      1,
				Description:   "List of cron schedule expressions based on which the scheduled run should be triggered.",
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
			"next_schedule": {
				Type:        schema.TypeInt,
				Description: "Timestamp (unix timestamp) of when the next run will be scheduled.",
				Computed:    true,
			},
			"runtime_config": {
				Type:        schema.TypeList,
				Description: "Customer provided runtime configuration for this scheduled run.",
				Optional:    true,
				MaxItems:    1,
				Elem:        &schema.Resource{Schema: scheduledRunRuntimeConfigSchema()},
			},
		},
	}
}

func resourceScheduledRunCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	parsedScheduledRun, err := parseScheduledRunInput(d)
	if err != nil {
		return diag.Errorf("could not extract scheduled run: %s", err)
	}

	var mutation struct {
		CreateRunSchedule structs.ScheduledRun `graphql:"stackScheduledRunCreate(stack: $stack, input: $input)"`
	}

	variables := map[string]any{
		"stack": toID(d.Get("stack_id").(string)),
		"input": *parsedScheduledRun,
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ScheduledRunCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create scheduled run: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(fmt.Sprintf("%s/%s", d.Get("stack_id"), mutation.CreateRunSchedule.ID))

	return resourceScheduledRunRead(ctx, d, meta)
}

func resourceScheduledRunRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
			ScheduledRun *structs.ScheduledRun `graphql:"scheduledRun(id: $id)"`
		} `graphql:"stack(id: $stack)"`
	}

	variables := map[string]any{
		"stack": toID(stackID),
		"id":    toID(scheduleID),
	}

	if err := meta.(*internal.Client).Query(ctx, "StackScheduledRunRead", &query, variables); err != nil {
		return diag.Errorf("could not query for scheduled run config: %v", internal.FromSpaceliftError(err))
	}

	if query.Stack == nil || query.Stack.ScheduledRun == nil {
		d.SetId("")
		return nil
	}

	if err := structs.PopulateRunSchedule(d, query.Stack.ScheduledRun); err != nil {
		return diag.Errorf("could not populate scheduled run config: %v", err)
	}

	return nil
}

func resourceScheduledRunUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	scheduledRun, err := parseScheduledRunInput(d)
	if err != nil {
		return diag.Errorf("could not extract scheduled run: %s", err)
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
		UpdateRunSchedule structs.ScheduledRun `graphql:"stackScheduledRunUpdate(stack: $stack, scheduledRun: $scheduledRun, input: $input)"`
	}

	variables := map[string]any{
		"stack":        toID(stackID),
		"scheduledRun": toID(scheduleID),
		"input":        *scheduledRun,
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ScheduledRunUpdate", &mutation, variables); err != nil {
		return diag.Errorf("could not update scheduled run: %v", internal.FromSpaceliftError(err))
	}

	return resourceScheduledRunRead(ctx, d, meta)
}

func resourceScheduledRunDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
		DeleteRunSchedule structs.ScheduledRun `graphql:"stackScheduledRunDelete(stack: $stack, scheduledRun: $scheduledRun)"`
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ScheduledRunDelete", &mutation, map[string]any{
		"stack":        toID(stackID),
		"scheduledRun": toID(scheduleID),
	}); err != nil {
		return diag.Errorf("could not delete scheduled run config: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

// scheduledRunRuntimeConfigSchema returns the runtime_config schema for the
// spacelift_scheduled_run resource: the shared input fields plus the
// read-only terraform_version and terraform_workflow_tool fields populated
// from the server.
func scheduledRunRuntimeConfigSchema() map[string]*schema.Schema {
	s := runtimeConfigInputSchema(false)
	s["terraform_version"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: "Terraform version to use",
		Computed:    true,
	}
	s["terraform_workflow_tool"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: "Defines the tool that will be used to execute the workflow. This can be one of `OPEN_TOFU`, `TERRAFORM_FOSS` or `CUSTOM`. Defaults to `TERRAFORM_FOSS`.",
		Computed:    true,
	}
	return s
}

func parseScheduledRunInput(d *schema.ResourceData) (*structs.ScheduledRunInput, error) {
	cfg := &structs.ScheduledRunInput{
		RuntimeConfig: parseRuntimeConfigInput(d, "runtime_config"),
	}

	name, ok := d.GetOk("name")
	if ok && name != nil {
		cfg.Name = toString(name)
	}

	every, everyOk := d.GetOk("every")
	at, atOk := d.GetOk("at")

	if everyOk && every != nil {
		cfg.CronSchedule = listToStringList(every)

		timezone, getOk := d.GetOk("timezone")
		if getOk && timezone != nil {
			cfg.Timezone = toOptionalString(timezone)
		}
	} else if atOk && at != nil {
		cfg.TimestampSchedule = toOptionalInt(at)
	} else {
		return nil, errors.New("Either `at` or `every` attribute is required")
	}

	return cfg, nil
}
