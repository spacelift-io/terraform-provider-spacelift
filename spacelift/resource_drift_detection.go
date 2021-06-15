package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourceDriftDetection() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_drift_detection` represents a Drift Detection configuration " +
			"for a Stack. It will run a proposed run on the given schedule, which you can " +
			"listen on using run state webhooks. If reconcile is true, then a tracked run " +
			"will be triggered when drift is detected.",

		CreateContext: resourceDriftDetectionCreate,
		ReadContext:   resourceDriftDetectionRead,
		UpdateContext: resourceDriftDetectionCreate,
		DeleteContext: resourceDriftDetectionDelete,

		Importer: &schema.ResourceImporter{StateContext: importIntegration},

		Schema: map[string]*schema.Schema{
			"reconcile": {
				Type:        schema.TypeBool,
				Description: "Whether a tracked run should be triggered when drift is detected.",
				Optional:    true,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack for which to set up drift detection",
				Required:    true,
				ForceNew:    true,
			},
			"schedule": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    1,
				Description: "List of cron schedule expressions based on which drift detection should be triggered.",
				Required:    true,
			},
		},
	}
}

func resourceDriftDetectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreateDriftDetectionIntegration struct {
			Reconcile bool     `graphql:"reconcile"`
			Schedule  []string `graphql:"schedule"`
		} `graphql:"stackIntegrationDriftDetectionCreate(stack: $stack, input: $input)"`
	}

	var scheduleExpression []graphql.String
	for _, expr := range d.Get("schedule").([]interface{}) {
		scheduleExpression = append(scheduleExpression, graphql.String(expr.(string)))
	}

	stackID := d.Get("stack_id").(string)

	variables := map[string]interface{}{
		"stack": toID(stackID),
		"input": map[string]interface{}{
			"schedule": scheduleExpression,
		},
	}

	if err := meta.(*internal.Client).Mutate(ctx, "DriftDetectionCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not generate dedicated DriftDetection integration for the stack: %v", err)
	}

	if d.Id() == "" {
		d.SetId(stackID)
	}

	return resourceDriftDetectionRead(ctx, d, meta)
}

func resourceDriftDetectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceStackDriftDetectionReadWithHooks(ctx, d, meta, func(_ string) diag.Diagnostics {
		d.SetId("")
		return nil
	})
}

func resourceDriftDetectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteDriftDetectionIntegration struct {
			Deleted bool `graphql:"deleted"`
		} `graphql:"stackIntegrationDriftDetectionDelete(stack: $stack)"`
	}

	variables := map[string]interface{}{"stack": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "DriftDetectionDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete stack DriftDetection service account: %v", err)
	}

	d.SetId("")

	return nil
}

func resourceStackDriftDetectionReadWithHooks(ctx context.Context, d *schema.ResourceData, meta interface{}, onNil func(message string) diag.Diagnostics) diag.Diagnostics {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Query(ctx, "StackDriftDetectionRead", &query, variables); err != nil {
		return diag.Errorf("could not query for stack: %v", err)
	}

	if query.Stack == nil {
		return onNil("stack not found")
	}

	integration := query.Stack.Integrations.DriftDetection

	d.Set("reconcile", integration.Reconcile)

	schedule := make([]interface{}, len(integration.Schedule))
	for _, expr := range integration.Schedule {
		schedule = append(schedule, expr)
	}
	d.Set("schedule", schedule)

	return nil
}
