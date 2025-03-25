package spacelift

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceTask() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_task` represents a task in Spacelift.",

		CreateContext: resourceTaskCreate,
		ReadContext:   schema.NoopContext,
		DeleteContext: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack for which to run the task",
				Required:    true,
				ForceNew:    true,
			},
			"command": {
				Type:             schema.TypeString,
				Description:      "Command that will be run.",
				ValidateDiagFunc: validations.DisallowEmptyString,
				Required:         true,
				ForceNew:         true,
			},
			"init": {
				Type:        schema.TypeBool,
				Description: "Whether to initialize the stack or not. Default: `true`",
				Optional:    true,
				Default:     true,
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
			"wait": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Wait for the run to finish",
				MaxItems:    1,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disabled": {
							Type:        schema.TypeBool,
							Description: "Whether waiting for the task is disabled or not. Default: `false`",
							Optional:    true,
							Default:     false,
							ForceNew:    true,
						},
						"continue_on_state": {
							Type: schema.TypeSet,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Continue on the specified states of a finished run. If not specified, the default is `[ 'finished' ]`. You can use following states: `applying`, `canceled`, `confirmed`, `destroying`, `discarded`, `failed`, `finished`, `initializing`, `pending_review`, `performing`, `planning`, `preparing_apply`, `preparing_replan`, `preparing`, `queued`, `ready`, `replan_requested`, `skipped`, `stopped`, `unconfirmed`.",
							Optional:    true,
							ForceNew:    true,
						},
						"continue_on_timeout": {
							Type:        schema.TypeBool,
							Description: "Continue if task timed out, i.e. did not reach any defined end state in time. Default: `false`",
							Optional:    true,
							Default:     false,
							ForceNew:    true,
						},
					},
				},
			},
		},
	}
}

func resourceTaskCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	task, err := structs.NewTaskInput(d)
	if err != nil {
		return diag.Errorf("could not extract task: %s", err)
	}

	var mutation struct {
		CreateTask structs.Task `graphql:"taskCreate(stack: $stack, command: $command, skipInitialization: $skipInitialization)"`
	}

	variables := map[string]interface{}{
		"stack":              graphql.ID(task.StackID),
		"command":            graphql.String(task.Command),
		"skipInitialization": graphql.Boolean(!task.Init),
	}

	client := meta.(*internal.Client)
	if err := client.Mutate(ctx, "TaskCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create task: %v", internal.FromSpaceliftError(err))
	}

	if waitRaw, ok := d.GetOk("wait"); ok {
		wait := structs.NewWaitConfiguration(waitRaw.([]interface{}))
		if d := wait.Wait(ctx, d, client, task.StackID, mutation.CreateTask.ID.(string)); len(d) > 0 {
			return d
		}
	}

	d.SetId(fmt.Sprintf("%s/%s", d.Get("stack_id"), mutation.CreateTask.ID))

	return nil
}
