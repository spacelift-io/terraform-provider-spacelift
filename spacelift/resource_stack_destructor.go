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

func resourceStackDestructor() *schema.Resource {
	deleteTimeout := time.Hour * 2

	return &schema.Resource{
		Description: "" +
			"`spacelift_stack_destructor` is used to destroy the resources of a " +
			"Stack before deleting it. `depends_on` should be used to make sure " +
			"that all necessary resources (environment variables, roles, " +
			"integrations, etc.) are still in place when the destruction run is " +
			"executed. **Note:** Destroying this resource will delete the " +
			"resources in the stack. If this resource needs to be deleted and " +
			"the resources in the stacks are to be preserved, ensure that the " +
			"`deactivated` attribute is set to `true`.",

		CreateContext: resourceStackDestructorCreate,
		ReadContext:   resourceStackDestructorRead,
		UpdateContext: schema.NoopContext,
		DeleteContext: resourceStackDestructorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"stack_id": {
				Type:             schema.TypeString,
				Description:      "ID of the stack to delete and destroy on destruction",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"deactivated": {
				Type:        schema.TypeBool,
				Description: "If set to true, destruction won't delete the resources (e.g. AWS/Azure/GCP infrastructure) managed by the stack",
				Optional:    true,
			},
			"discard_runs": {
				Type:        schema.TypeBool,
				Description: "If set to true, destruction will also discard all runs on the stack that are eligible for discarding (e.g. not in progress runs)",
				Optional:    true,
				Default:     false,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: &deleteTimeout,
		},
	}
}

func resourceStackDestructorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId(fmt.Sprintf("destructor-%d", time.Now().Unix()))
	return resourceStackDestructorRead(ctx, d, meta)
}

func resourceStackDestructorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Get("stack_id"))}

	if err := meta.(*internal.Client).Query(ctx, "StackDestructorRead", &query, variables); err != nil {
		return diag.Errorf("could not query for stack: %v", err)
	}

	if query.Stack == nil {
		d.SetId("")
	}

	return nil
}

func resourceStackDestructorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if deactivated := d.Get("deactivated"); deactivated != nil && deactivated.(bool) {
		d.SetId("")
		return nil
	}

	stackID := d.Get("stack_id").(string)
	variables := map[string]interface{}{"id": toID(stackID)}

	if discardRuns := d.Get("discard_runs"); discardRuns != nil && discardRuns.(bool) {
		var mutation struct {
			RunDiscard *structs.RunDiscardAll `graphql:"runDiscardAll(stack: $id)"`
		}
		if err := meta.(*internal.Client).Mutate(ctx, "StackDestructorDiscardRuns", &mutation, variables); err != nil {
			return diag.Errorf("could not discard runs on stack %s: %v", stackID, internal.FromSpaceliftError(err))
		}
		if len(mutation.RunDiscard.FailedDiscarding) > 0 {
			return diag.Errorf("could not discard runs on stack %s: %v", stackID, mutation.RunDiscard.FailedDiscarding)
		}

		// Wait for the stack to be unblocked after discarding runs.
		// The backend clears blockers asynchronously when runs reach terminal states.
		if diagnostics := waitForStackUnblocked(ctx, meta.(*internal.Client), stackID); diagnostics.HasError() {
			return diagnostics
		}
	}

	var mutation struct {
		DeleteStack *structs.Stack `graphql:"stackDelete(id: $id, destroyResources: true)"`
	}
	if err := meta.(*internal.Client).Mutate(ctx, "StackDestructorDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete stack %s: %v", stackID, internal.FromSpaceliftError(err))
	}

	if mutation.DeleteStack != nil && mutation.DeleteStack.Deleting {
		if diagnostics := waitForDestroy(ctx, meta.(*internal.Client), stackID); diagnostics.HasError() {
			return diagnostics
		}
	}

	d.SetId("")

	return nil
}

func waitForStackUnblocked(ctx context.Context, client *internal.Client, id string) diag.Diagnostics {
	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()

	// Wait up to 2 minutes for the stack blocker to be cleared
	timeout := time.After(time.Minute * 2)

	for {
		select {
		case <-ctx.Done():
			return diag.FromErr(ctx.Err())
		case <-timeout:
			return diag.Errorf("timeout waiting for stack %s to be unblocked after discarding runs", id)
		case <-ticker.C:
		}

		// Use a custom struct to query only the blocker field
		var query struct {
			Stack *struct {
				Blocker *struct {
					ID string `graphql:"id"`
				} `graphql:"blocker"`
			} `graphql:"stack(id: $id)"`
		}

		variables := map[string]interface{}{"id": graphql.ID(id)}

		if err := client.Query(ctx, "StackCheckBlocker", &query, variables); err != nil {
			return diag.Errorf("could not query for stack %s blocker status: %v", id, err)
		}

		stack := query.Stack
		if stack == nil {
			return diag.Errorf("stack %s not found while waiting for unblock", id)
		}

		// Check if the stack is no longer blocked
		if stack.Blocker == nil {
			return nil
		}

		// Stack still has a blocker, continue waiting
	}
}

func waitForDestroy(ctx context.Context, client *internal.Client, id string) diag.Diagnostics {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return diag.FromErr(ctx.Err())
		case <-ticker.C:
		}

		var query struct {
			Stack *structs.Stack `graphql:"stack(id: $id)"`
		}

		variables := map[string]interface{}{"id": graphql.ID(id)}

		if err := client.Query(ctx, "StackCheckState", &query, variables); err != nil {
			return diag.Errorf("could not query for stack %s: %v", id, err)
		}

		stack := query.Stack
		if stack == nil {
			return nil
		}

		if !stack.Deleting {
			return diag.Errorf("destruction of stack %s unsuccessful, please check the destruction run logs", id)
		}
	}
}
