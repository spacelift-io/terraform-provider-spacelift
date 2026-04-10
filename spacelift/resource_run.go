package spacelift

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceRun() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_run` allows programmatically triggering runs in response " +
			"to arbitrary changes in the keepers section.",

		CreateContext: resourceRunCreate,
		ReadContext:   schema.NoopContext,
		Delete:        schema.RemoveFromState,
		UpdateContext: schema.NoopContext,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"stack_id": {
				Type:             schema.TypeString,
				Description:      "ID of the stack on which the run is to be triggered.",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"commit_sha": {
				Description: "The commit SHA for which to trigger a run.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"keepers": {
				Description: "" +
					"Arbitrary map of values that, when changed, will trigger " +
					"recreation of the resource.",
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
			"proposed": {
				Type:        schema.TypeBool,
				Description: "Whether the run is a proposed run. Defaults to `false`.",
				Optional:    true,
				ForceNew:    true,
				Default:     false,
			},
			"id": {
				Description: "The ID of the triggered run.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"runtime_config": {
				Type:        schema.TypeList,
				Description: "Custom runtime configuration for this run.",
				Optional:    true,
				ForceNew:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project_root": {
							Type:        schema.TypeString,
							Description: "Project root is the optional directory relative to the workspace root containing the entrypoint to the Stack.",
							Optional:    true,
						},
						"runner_image": {
							Type:        schema.TypeString,
							Description: "Name of the Docker image used to process Runs.",
							Optional:    true,
						},
						"environment": {
							Type:        schema.TypeSet,
							Description: "Environment variables for the run.",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:        schema.TypeString,
										Description: "Environment variable key.",
										Required:    true,
									},
									"value": {
										Type:        schema.TypeString,
										Description: "Environment variable value.",
										Required:    true,
									},
								},
							},
						},
						"after_apply": {
							Type:        schema.TypeList,
							Description: "List of after-apply scripts.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"after_destroy": {
							Type:        schema.TypeList,
							Description: "List of after-destroy scripts.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"after_init": {
							Type:        schema.TypeList,
							Description: "List of after-init scripts.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"after_perform": {
							Type:        schema.TypeList,
							Description: "List of after-perform scripts.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"after_plan": {
							Type:        schema.TypeList,
							Description: "List of after-plan scripts.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"after_run": {
							Type:        schema.TypeList,
							Description: "List of after-run scripts.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"before_apply": {
							Type:        schema.TypeList,
							Description: "List of before-apply scripts.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"before_destroy": {
							Type:        schema.TypeList,
							Description: "List of before-destroy scripts.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"before_init": {
							Type:        schema.TypeList,
							Description: "List of before-init scripts.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"before_perform": {
							Type:        schema.TypeList,
							Description: "List of before-perform scripts.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"before_plan": {
							Type:        schema.TypeList,
							Description: "List of before-plan scripts.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
					},
				},
			},
			"wait": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Wait for the run to finish",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disabled": {
							Type:        schema.TypeBool,
							Description: "Whether waiting for a job is disabled or not. Default: `false`",
							Optional:    true,
							Default:     false,
						},
						"continue_on_state": {
							Type: schema.TypeSet,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Continue on the specified states of a finished run. If not specified, the default is `[ 'finished' ]`. You can use following states: `applying`, `canceled`, `confirmed`, `destroying`, `discarded`, `failed`, `finished`, `initializing`, `pending_review`, `performing`, `planning`, `preparing_apply`, `preparing_replan`, `preparing`, `queued`, `ready`, `replan_requested`, `skipped`, `stopped`, `unconfirmed`.",
							Optional:    true,
						},
						"continue_on_timeout": {
							Type:        schema.TypeBool,
							Description: "Continue if run timed out, i.e. did not reach any defined end state in time. Default: `false`",
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
		},
	}
}

func resourceRunCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var mutation struct {
		Run structs.Run `graphql:"runTrigger(stack: $stack, commitSha: $commitSha, runType: $runType, runtimeConfig: $runtimeConfig)"`
	}

	stackID := d.Get("stack_id").(string)

	runType := structs.RunTypeTracked
	if proposed, ok := d.GetOk("proposed"); ok && proposed.(bool) {
		runType = structs.RunTypeProposed
	}

	variables := map[string]any{
		"stack":         toID(stackID),
		"commitSha":     (*graphql.String)(nil),
		"runType":       &runType,
		"runtimeConfig": (*structs.RuntimeConfigInput)(nil),
	}

	if sha, ok := d.GetOk("commit_sha"); ok {
		variables["commitSha"] = toOptionalString(sha)
	}

	if runtimeConfig, ok := d.Get("runtime_config").([]any); ok && len(runtimeConfig) > 0 {
		variables["runtimeConfig"] = parseRuntimeConfig(runtimeConfig)
	}

	client := meta.(*internal.Client)
	if err := client.Mutate(ctx, "RunTrigger", &mutation, variables); err != nil {
		return diag.Errorf("could not trigger run for stack %s: %v", stackID, internal.FromSpaceliftError(err))
	}

	if waitRaw, ok := d.GetOk("wait"); ok {
		wait := structs.NewWaitConfiguration(waitRaw.([]any))
		if diag := wait.Wait(ctx, d, client, stackID, mutation.Run.ID); len(diag) > 0 {
			return diag
		}
	}

	d.SetId(mutation.Run.ID)
	return nil
}

func parseRuntimeConfig(runtimeConfig []any) *structs.RuntimeConfigInput {
	mapped := runtimeConfig[0].(map[string]any)

	environment := []structs.EnvVarInput{}
	for _, e := range mapped["environment"].(*schema.Set).List() {
		envMap := e.(map[string]any)
		environment = append(environment, structs.EnvVarInput{
			Key:   toString(envMap["key"]),
			Value: toString(envMap["value"]),
		})
	}

	cfg := &structs.RuntimeConfigInput{
		AfterApply:    listToOptionalStringList(mapped["after_apply"]),
		AfterDestroy:  listToOptionalStringList(mapped["after_destroy"]),
		AfterInit:     listToOptionalStringList(mapped["after_init"]),
		AfterPerform:  listToOptionalStringList(mapped["after_perform"]),
		AfterPlan:     listToOptionalStringList(mapped["after_plan"]),
		AfterRun:      listToOptionalStringList(mapped["after_run"]),
		BeforeApply:   listToOptionalStringList(mapped["before_apply"]),
		BeforeDestroy: listToOptionalStringList(mapped["before_destroy"]),
		BeforeInit:    listToOptionalStringList(mapped["before_init"]),
		BeforePerform: listToOptionalStringList(mapped["before_perform"]),
		BeforePlan:    listToOptionalStringList(mapped["before_plan"]),
		Environment:   &environment,
	}

	if projectRoot := mapped["project_root"]; len(projectRoot.(string)) > 0 {
		cfg.ProjectRoot = toOptionalString(projectRoot)
	}
	if runnerImage := mapped["runner_image"]; len(runnerImage.(string)) > 0 {
		cfg.RunnerImage = toOptionalString(runnerImage)
	}

	return cfg
}
