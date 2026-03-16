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
		ID string `graphql:"runResourceCreate(stack: $stack, commitSha: $sha, proposed: $proposed)"`
	}

	stackID := d.Get("stack_id").(string)

	variables := map[string]any{
		"stack":    toID(stackID),
		"sha":      (*graphql.String)(nil),
		"proposed": (*graphql.Boolean)(nil),
	}

	if sha, ok := d.GetOk("commit_sha"); ok {
		variables["sha"] = new(graphql.String(sha.(string)))
	}

	if proposed, ok := d.GetOk("proposed"); ok {
		variables["proposed"] = new(graphql.Boolean(proposed.(bool)))
	}

	client := meta.(*internal.Client)
	if err := client.Mutate(ctx, "ResourceRunCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not trigger run for stack %s: %v", stackID, internal.FromSpaceliftError(err))
	}

	if waitRaw, ok := d.GetOk("wait"); ok {
		wait := structs.NewWaitConfiguration(waitRaw.([]any))
		if diag := wait.Wait(ctx, d, client, stackID, mutation.ID); len(diag) > 0 {
			return diag
		}
	}

	d.SetId(mutation.ID)
	return nil
}
