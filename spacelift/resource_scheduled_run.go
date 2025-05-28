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
				Type:          schema.TypeList,
				Description:   "Customer provided runtime configuration for this scheduled run.",
				Optional:      true,
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// "yaml": {
						// 	Type:        schema.TypeString,
						// 	Description: "YAML representation of the runtime configuration.",
						// 	Optional:    true,
						// },
						"project_root": {
							Type:        schema.TypeString,
							Description: "Project root is the optional directory relative to the workspace root containing the entrypoint to the Stack.",
							Optional:    true,
						},
						"runner_image": {
							Type:        schema.TypeString,
							Description: "Name of the Docker image used to process Runs",
							Optional:    true,
						},
						"terraform_version": {
							Type:        schema.TypeString,
							Description: "Terraform version to use",
							Optional:    true,
						},
						"terraform_workflow_tool": {
							Type:        schema.TypeString,
							Description: "Defines the tool that will be used to execute the workflow. This can be one of `OPEN_TOFU`, `TERRAFORM_FOSS` or `CUSTOM`. Defaults to `TERRAFORM_FOSS`.",
							Optional:    true,
						},
						"environment": {
							Type:        schema.TypeList,
							Description: "Environment variables for the run",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:        schema.TypeString,
										Description: "Environment variable key",
										Required:    true,
									},
									"value": {
										Type:        schema.TypeString,
										Description: "Environment variable value",
										Required:    true,
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
						},
						"after_destroy": {
							Type:        schema.TypeList,
							Description: "List of after-destroy scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
						"after_init": {
							Type:        schema.TypeList,
							Description: "List of after-init scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
						"after_perform": {
							Type:        schema.TypeList,
							Description: "List of after-perform scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
						"after_plan": {
							Type:        schema.TypeList,
							Description: "List of after-plan scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
						"after_run": {
							Type:        schema.TypeList,
							Description: "List of after-run scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
						"before_apply": {
							Type:        schema.TypeList,
							Description: "List of before-apply scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
						"before_destroy": {
							Type:        schema.TypeList,
							Description: "List of before-destroy scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
						"before_init": {
							Type:        schema.TypeList,
							Description: "List of before-init scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
						"before_perform": {
							Type:        schema.TypeList,
							Description: "List of before-perform scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
						"before_plan": {
							Type:        schema.TypeList,
							Description: "List of before-plan scripts",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
						"terragrunt": {
							Type:        schema.TypeList,
							Description: "Terragrunt-specific configuration",
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"terraform_version": {
										Type:        schema.TypeString,
										Description: "Terraform version to use with Terragrunt",
										Optional:    true,
									},
									"terragrunt_version": {
										Type:        schema.TypeString,
										Description: "Terragrunt version to use",
										Optional:    true,
									},
									"use_run_all": {
										Type:        schema.TypeBool,
										Description: "Whether to use `terragrunt run-all` for execution",
										Optional:    true,
									},
									"tool": {
										Type:        schema.TypeString,
										Description: "Tool to use for Terragrunt execution (TERRAFORM_FOSS, OPEN_TOFU, MANUALLY_PROVISIONED)",
										Optional:    true,
									},
								},
							},
						},
						"terraform": {
							Type:        schema.TypeList,
							Description: "Terraform-specific configuration",
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"workflow_tool": {
										Type:        schema.TypeString,
										Description: "Workflow tool to use (TERRAFORM_FOSS, OPEN_TOFU, CUSTOM)",
										Optional:    true,
									},
									"version": {
										Type:        schema.TypeString,
										Description: "Terraform version to use",
										Optional:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceScheduledRunCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	parsedScheduledRun, err := parseScheduledRunInput(d)
	if err != nil {
		return diag.Errorf("could not extract scheduled run: %s", err)
	}

	var mutation struct {
		CreateRunSchedule structs.ScheduledRun `graphql:"stackScheduledRunCreate(stack: $stack, input: $input)"`
	}

	variables := map[string]interface{}{
		"stack": toID(d.Get("stack_id").(string)),
		"input": *parsedScheduledRun,
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ScheduledRunCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create scheduled run: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(fmt.Sprintf("%s/%s", d.Get("stack_id"), mutation.CreateRunSchedule.ID))

	return resourceScheduledRunRead(ctx, d, meta)
}

func resourceScheduledRunRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	variables := map[string]interface{}{"stack": toID(stackID), "id": toID(scheduleID)}

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

func resourceScheduledRunUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	input := structs.ScheduledRunInput{
		Name: graphql.String(scheduledRun.Name),
	}

	// if scheduledRun.At != nil {
	// 	input.TimestampSchedule = graphql.NewInt(graphql.Int(*scheduledRun.At)) //nolint:gosec // safe: value known to fit in int32
	// } else {
	// 	var scheduleExpressions []graphql.String
	// 	for _, expr := range scheduledRun.Every {
	// 		scheduleExpressions = append(scheduleExpressions, graphql.String(expr))
	// 	}
	//
	// 	input.CronSchedule = scheduleExpressions
	// 	input.Timezone = graphql.NewString(graphql.String(scheduledRun.Timezone))
	// }
	//
	// if scheduledRun.CustomRuntimeConfig != nil {
	// 	input.RuntimeConfig = &structs.RuntimeConfigInput{
	// 		Yaml: graphql.NewString(graphql.String(*scheduledRun.CustomRuntimeConfig)),
	// 	}
	// }

	variables := map[string]interface{}{
		"stack":        toID(stackID),
		"scheduledRun": toID(scheduleID),
		"input":        input,
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ScheduledRunUpdate", &mutation, variables); err != nil {
		return diag.Errorf("could not update scheduled run: %v", internal.FromSpaceliftError(err))
	}

	return resourceScheduledRunRead(ctx, d, meta)
}

func resourceScheduledRunDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	if err := meta.(*internal.Client).Mutate(ctx, "ScheduledRunDelete", &mutation, map[string]interface{}{
		"stack":        toID(stackID),
		"scheduledRun": toID(scheduleID),
	}); err != nil {
		return diag.Errorf("could not delete scheduled run config: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func parseScheduledRunInput(d *schema.ResourceData) (*structs.ScheduledRunInput, error) {
	cfg := &structs.ScheduledRunInput{}

	name, ok := d.GetOk("name")
	if ok && name != nil {
		cfg.Name = *graphql.NewString(toString(name))
	}

	runtimeConfig, ok := d.Get("runtime_config").([]interface{})
	if ok && len(runtimeConfig) > 0 {
		mapped := runtimeConfig[0].(map[string]interface{})

		environment := []structs.EnvVarInput{}
		for _, e := range mapped["environment"].([]interface{}){
			envMap := e.(map[string]interface{})
			environment = append(environment, structs.EnvVarInput{
				Key:   toString(envMap["key"]),
				Value: toString(envMap["value"]),
			})
		}


		cfg.RuntimeConfig = &structs.RuntimeConfigInput{
			Yaml: toOptionalNullableString(mapped["yaml"]),
			ProjectRoot: toOptionalNullableString(mapped["project_root"]),
			RunnerImage: toOptionalNullableString(mapped["runner_image"]),
			AfterApply: toOptionalStringList(mapped["after_apply"]),
			AfterDestroy: toOptionalStringList(mapped["after_destroy"]),
			AfterInit: toOptionalStringList(mapped["after_init"]),
			AfterPerform: toOptionalStringList(mapped["after_perform"]),
			AfterPlan: toOptionalStringList(mapped["after_plan"]),
			AfterRun: toOptionalStringList(mapped["after_run"]),
			BeforeApply: toOptionalStringList(mapped["before_apply"]),
			BeforeDestroy: toOptionalStringList(mapped["before_destroy"]),
			BeforeInit: toOptionalStringList(mapped["before_init"]),
			BeforePerform: toOptionalStringList(mapped["before_perform"]),
			BeforePlan: toOptionalStringList(mapped["before_plan"]),
			Environment: &environment,
		}
	}

	every, everyOk := d.GetOk("every")
	at, atOk := d.GetOk("at")

	if everyOk && every != nil {
		v := every.([]interface{})
		if len(v) > 0 {
			var scheduleExpressions []graphql.String
			for _, expr := range v {
				scheduleExpressions = append(scheduleExpressions, toString(expr))
			}
			cfg.CronSchedule = scheduleExpressions
		}

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
