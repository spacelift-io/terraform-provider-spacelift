package spacelift

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
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
							Description: "Continue on the specified states of a finished run. If not specified, the default is `[ 'finished' ]`",
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

type waitConfiguration struct {
	disabled          bool
	continueOnState   []string
	continueOnTimeout bool
}

func expandWaitConfiguration(input []interface{}) *waitConfiguration {
	if len(input) == 0 {
		return nil
	}
	v := input[0].(map[string]interface{})
	cfg := &waitConfiguration{
		disabled:          v["disabled"].(bool),
		continueOnState:   []string{},
		continueOnTimeout: v["continue_on_timeout"].(bool),
	}

	if v, ok := v["continue_on_state"]; ok {
		for _, item := range v.(*schema.Set).List() {
			str, ok := item.(string)
			if !ok {
				panic(fmt.Sprintf("continue_on_state contains a non-string element %+v", str))
			}
			cfg.continueOnState = append(cfg.continueOnState, str)
		}
	}
	if len(cfg.continueOnState) == 0 {
		cfg.continueOnState = append(cfg.continueOnState, "finished")
	}
	return cfg
}

func resourceRunCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		ID string `graphql:"runResourceCreate(stack: $stack, commitSha: $sha, proposed: $proposed)"`
	}

	stackID := d.Get("stack_id").(string)

	variables := map[string]interface{}{
		"stack":    toID(stackID),
		"sha":      (*graphql.String)(nil),
		"proposed": (*graphql.Boolean)(nil),
	}

	if sha, ok := d.GetOk("commit_sha"); ok {
		variables["sha"] = graphql.NewString(graphql.String(sha.(string)))
	}

	if proposed, ok := d.GetOk("proposed"); ok {
		variables["proposed"] = graphql.NewBoolean(graphql.Boolean(proposed.(bool)))
	}

	client := meta.(*internal.Client)
	if err := client.Mutate(ctx, "ResourceRunCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not trigger run for stack %s: %v", stackID, internal.FromSpaceliftError(err))
	}

	if waitRaw, ok := d.GetOk("wait"); ok {
		wait := expandWaitConfiguration(waitRaw.([]interface{}))

		if !wait.disabled {
			stateConf := &retry.StateChangeConf{
				ContinuousTargetOccurence: 1,
				Delay:                     10 * time.Second, // TODO: Delay must be a multiple of Timeout
				MinTimeout:                10 * time.Second,
				Pending: []string{
					"running",
				},
				Target: []string{
					"finished",
					"unconfirmed", // Let's treat unconfirmed as the target state.
					// It's not finished, but we don't want to wait for it because it requires confirmation from someone.

				},
				Refresh: checkStackStatusFunc(ctx, client, stackID, mutation.ID),
				Timeout: d.Timeout(schema.TimeoutCreate),
			}

			finalState, err := stateConf.WaitForStateContext(ctx)
			if err != nil {
				if timeoutErr, ok := internal.AsError[*retry.TimeoutError](err); ok {
					tflog.Debug(ctx, "received retry.TimeoutError from WaitForStateContext", map[string]any{
						"stackID":       stackID,
						"runID":         mutation.ID,
						"lastState":     timeoutErr.LastState,
						"expectedState": timeoutErr.ExpectedState,
					})
					finalState = "__timeout__"
				} else if err == context.DeadlineExceeded {
					tflog.Debug(ctx, "received context.DeadlineExceeded from WaitForStateContext", map[string]any{
						"stackID": stackID,
						"runID":   mutation.ID,
					})
					finalState = "__timeout__"
				} else {
					return diag.Errorf("failed waiting for run %s on stack %s to finish. error(%T): %+v ", mutation.ID, stackID, err, err)
				}
			}

			switch finalState.(string) {
			case "__timeout__":
				if !wait.continueOnTimeout {
					return diag.Errorf("run %s on stack %s has timed out", mutation.ID, stackID)
				} else {
					tflog.Info(ctx, "run timed out but continue_on_discarded=true",
						map[string]any{
							"stackID": stackID,
							"runID":   mutation.ID,
						})
				}
			default:
				if !slices.Contains[[]string](wait.continueOnState, finalState.(string)) {
					return diag.Errorf("run %s on stack %s has ended with status %s. expected %v", mutation.ID, stackID, finalState, wait.continueOnState)
				}
				tflog.Debug(ctx, "run finished", map[string]any{
					"stackID":    stackID,
					"runID":      mutation.ID,
					"finalState": finalState,
				})
			}
		}
	}

	d.SetId(mutation.ID)
	return nil
}

func checkStackStatusFunc(ctx context.Context, client *internal.Client, stackID string, runID string) retry.StateRefreshFunc {
	return func() (result any, state string, err error) {
		// instead of a resource handle we return the current state as result
		// Makes it easier to detect which end state has been reached.
		// Otherwise we would need another GraphQL query
		result, finished, err := getStackRunStateByID(ctx, client, stackID, runID)
		if err != nil {
			return
		}
		state = "running"
		if finished {
			state = "finished"
		}
		// Let's treat unconfirmed as the target state.
		// It's not finished, but we don't want to wait for it because it requires confirmation from someone.
		if result == "unconfirmed" {
			state = "unconfirmed"
		}
		return
	}
}

func getStackRunStateByID(ctx context.Context, client *internal.Client, stackID string, runID string) (string, bool, error) {
	var query struct {
		Stack struct {
			Run struct {
				ID       graphql.String
				State    graphql.String
				Finished graphql.Boolean
			} `graphql:"run(id: $runId)"`
		} `graphql:"stack(id: $stackId)"`
	}

	variables := map[string]interface{}{
		"stackId": graphql.ID(stackID),
		"runId":   graphql.ID(runID),
	}

	if err := client.Query(ctx, "StackRunRead", &query, variables); err != nil {
		return "", false, errors.Wrap(err, fmt.Sprintf("could not query for run %s of stack %s", runID, stackID))
	}

	currentState := strings.ToLower(string(query.Stack.Run.State))
	tflog.Debug(ctx, "current state of run", map[string]interface{}{
		"stackID":      stackID,
		"runID":        runID,
		"currentState": currentState,
		"finished":     query.Stack.Run.Finished,
	})
	return currentState, bool(query.Stack.Run.Finished), nil
}
