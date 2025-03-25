package structs

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
)

type WaitConfiguration struct {
	disabled          bool
	continueOnState   []string
	continueOnTimeout bool
}

func NewWaitConfiguration(input []interface{}) *WaitConfiguration {
	if len(input) == 0 {
		return nil
	}
	v := input[0].(map[string]interface{})
	cfg := &WaitConfiguration{
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

func (wait *WaitConfiguration) Wait(ctx context.Context, d *schema.ResourceData, client *internal.Client, stackID, mutationID string) diag.Diagnostics {
	if wait.disabled {
		return nil
	}

	stateConf := &retry.StateChangeConf{
		ContinuousTargetOccurence: 1,
		Delay:                     10 * time.Second,
		MinTimeout:                10 * time.Second,
		Pending: []string{
			"running",
		},
		Target: []string{
			"finished",
			"unconfirmed", // Let's treat unconfirmed as the target state.
			// It's not finished, but we don't want to wait for it because it requires confirmation from someone.
		},
		Refresh: checkStackStatusFunc(ctx, client, stackID, mutationID),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	finalState, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		if timeoutErr, ok := internal.AsError[*retry.TimeoutError](err); ok {
			tflog.Debug(ctx, "received retry.TimeoutError from WaitForStateContext", map[string]any{
				"stackID":       stackID,
				"runID":         mutationID,
				"lastState":     timeoutErr.LastState,
				"expectedState": timeoutErr.ExpectedState,
			})
			finalState = "__timeout__"
		} else if err == context.DeadlineExceeded {
			tflog.Debug(ctx, "received context.DeadlineExceeded from WaitForStateContext", map[string]any{
				"stackID": stackID,
				"runID":   mutationID,
			})
			finalState = "__timeout__"
		} else {
			return diag.Errorf("failed waiting for run %s on stack %s to finish. error(%T): %+v ", mutationID, stackID, err, err)
		}
	}

	switch finalState.(string) {
	case "__timeout__":
		if !wait.continueOnTimeout {
			return diag.Errorf("run %s on stack %s has timed out", mutationID, stackID)
		}
		tflog.Info(ctx, "run timed out but continue_on_timeout=true",
			map[string]any{
				"stackID": stackID,
				"runID":   mutationID,
			})
	default:
		if !slices.Contains[[]string](wait.continueOnState, finalState.(string)) {
			return diag.Errorf("run %s on stack %s has ended with status %s. expected %v", mutationID, stackID, finalState, wait.continueOnState)
		}
		tflog.Debug(ctx, "run finished", map[string]any{
			"stackID":    stackID,
			"runID":      mutationID,
			"finalState": finalState,
		})
	}

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
			RunResourceState struct {
				ID       graphql.String
				State    graphql.String
				Finished graphql.Boolean
			} `graphql:"runResourceState(id: $runId)"`
		} `graphql:"stack(id: $stackId)"`
	}

	variables := map[string]interface{}{
		"stackId": graphql.ID(stackID),
		"runId":   graphql.ID(runID),
	}

	if err := client.Query(ctx, "StackRunRead", &query, variables); err != nil {
		return "", false, errors.Wrap(err, fmt.Sprintf("could not query for run %s of stack %s", runID, stackID))
	}

	rrs := query.Stack.RunResourceState

	currentState := strings.ToLower(string(rrs.State))
	tflog.Debug(ctx, "current state of run", map[string]interface{}{
		"stackID":      stackID,
		"runID":        runID,
		"currentState": currentState,
		"finished":     rrs.Finished,
	})
	return currentState, bool(rrs.Finished), nil
}
