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
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the scheduled run",
				Optional:    true,
				Computed:    true,
			},
			"at": {
				Type:        schema.TypeInt,
				Description: "Timestamp (unix timestamp) at which time the scheduled run should happen.",
				Optional:    true,
				Computed:    true,
			},
			"every": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of cron schedule expressions based on which the scheduled run should be triggered.",
				Optional:    true,
				Computed:    true,
			},
			"timezone": {
				Type:        schema.TypeString,
				Description: "Timezone in which the schedule is expressed. Defaults to `UTC`.",
				Optional:    true,
				Computed:    true,
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
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"yaml": {
							Type:        schema.TypeString,
							Description: "YAML representation of the runtime configuration.",
							Optional:    true,
							Computed:    true,
						},
						"project_root": {
							Type:        schema.TypeString,
							Description: "Project root is the optional directory relative to the workspace root containing the entrypoint to the Stack.",
							Optional:    true,
							Computed:    true,
						},
						"runner_image": {
							Type:        schema.TypeString,
							Description: "Name of the Docker image used to process Runs",
							Optional:    true,
							Computed:    true,
						},
						"terraform_version": {
							Type:        schema.TypeString,
							Description: "Terraform version to use",
							Computed:    true,
						},
						"terraform_workflow_tool": {
							Type:        schema.TypeString,
							Description: "Defines the tool that will be used to execute the workflow. This can be one of `OPEN_TOFU`, `TERRAFORM_FOSS` or `CUSTOM`. Defaults to `TERRAFORM_FOSS`.",
							Computed:    true,
						},
						"environment": {
							Type:        schema.TypeSet,
							Description: "Environment variables for the run",
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:        schema.TypeString,
										Description: "Environment variable key",
										Computed:    true,
									},
									"value": {
										Type:        schema.TypeString,
										Description: "Environment variable value",
										Computed:    true,
									},
								},
							},
						},
						"after_apply": {
							Type:        schema.TypeList,
							Description: "List of after-apply scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
							Computed: true,
						},
						"after_destroy": {
							Type:        schema.TypeList,
							Description: "List of after-destroy scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
							Computed: true,
						},
						"after_init": {
							Type:        schema.TypeList,
							Description: "List of after-init scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
							Computed: true,
						},
						"after_perform": {
							Type:        schema.TypeList,
							Description: "List of after-perform scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
							Computed: true,
						},
						"after_plan": {
							Type:        schema.TypeList,
							Description: "List of after-plan scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
							Computed: true,
						},
						"after_run": {
							Type:        schema.TypeList,
							Description: "List of after-run scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
							Computed: true,
						},
						"before_apply": {
							Type:        schema.TypeList,
							Description: "List of before-apply scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
							Computed: true,
						},
						"before_destroy": {
							Type:        schema.TypeList,
							Description: "List of before-destroy scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
							Computed: true,
						},
						"before_init": {
							Type:        schema.TypeList,
							Description: "List of before-init scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
							Computed: true,
						},
						"before_perform": {
							Type:        schema.TypeList,
							Description: "List of before-perform scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
							Computed: true,
						},
						"before_plan": {
							Type:        schema.TypeList,
							Description: "List of before-plan scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
							Computed: true,
						},
					},
				},
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

	variables := map[string]interface{}{
		"stack": toID(stackID),
		"id":    toID(scheduleID),
	}

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
