package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceDriftDetection() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_drift_detection` represents a Drift Detection configuration " +
			"for a Stack. It will trigger a proposed run on the given schedule, which you can " +
			"listen for using run state webhooks. If reconcile is true, then a tracked run " +
			"will be triggered when drift is detected.",

		CreateContext: resourceDriftDetectionCreate,
		ReadContext:   resourceDriftDetectionRead,
		UpdateContext: resourceDriftDetectionUpdate,
		DeleteContext: resourceDriftDetectionDelete,

		Importer: &schema.ResourceImporter{StateContext: importIntegration},

		Schema: map[string]*schema.Schema{
			"reconcile": {
				Type:        schema.TypeBool,
				Description: "Whether a tracked run should be triggered when drift is detected.",
				Optional:    true,
			},
			"stack_id": {
				Type:             schema.TypeString,
				Description:      "ID of the stack for which to set up drift detection",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"schedule": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				MinItems:    1,
				Description: "List of cron schedule expressions based on which drift detection should be triggered.",
				Required:    true,
			},
			"timezone": {
				Type:        schema.TypeString,
				Description: "Timezone in which the schedule is expressed. Defaults to `UTC`.",
				Optional:    true,
				Default:     "UTC",
			},
		},
	}
}

func resourceDriftDetectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreateDriftDetectionIntegration struct {
			Reconcile bool     `graphql:"reconcile"`
			Schedule  []string `graphql:"schedule"`
			Timezone  string   `graphql:"timezone"`
		} `graphql:"stackIntegrationDriftDetectionCreate(stack: $stack, input: $input)"`
	}

	var scheduleExpressions []graphql.String
	for _, expr := range d.Get("schedule").([]interface{}) {
		scheduleExpressions = append(scheduleExpressions, graphql.String(expr.(string)))
	}

	stackID := d.Get("stack_id").(string)

	variables := map[string]interface{}{
		"stack": toID(stackID),
		"input": structs.DriftDetectionIntegrationInput{
			Reconcile: graphql.Boolean(d.Get("reconcile").(bool)),
			Schedule:  scheduleExpressions,
			Timezone:  graphql.NewString(graphql.String(d.Get("timezone").(string))),
		},
	}

	if err := meta.(*internal.Client).Mutate(ctx, "DriftDetectionCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create drift detection integration for the stack: %v", err)
	}

	d.SetId(stackID)

	return resourceDriftDetectionRead(ctx, d, meta)
}

func resourceDriftDetectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UpdateDriftDetectionIntegration struct {
			Reconcile bool     `graphql:"reconcile"`
			Schedule  []string `graphql:"schedule"`
			Timezone  string   `graphql:"timezone"`
		} `graphql:"stackIntegrationDriftDetectionUpdate(stack: $stack, input: $input)"`
	}

	var scheduleExpressions []graphql.String
	for _, expr := range d.Get("schedule").([]interface{}) {
		scheduleExpressions = append(scheduleExpressions, graphql.String(expr.(string)))
	}

	stackID := d.Get("stack_id").(string)

	variables := map[string]interface{}{
		"stack": toID(stackID),
		"input": structs.DriftDetectionIntegrationInput{
			Reconcile: graphql.Boolean(d.Get("reconcile").(bool)),
			Schedule:  scheduleExpressions,
			Timezone:  graphql.NewString(graphql.String(d.Get("timezone").(string))),
		},
	}

	if err := meta.(*internal.Client).Mutate(ctx, "DriftDetectionUpdate", &mutation, variables); err != nil {
		return diag.Errorf("could not update drift detection integration for the stack: %v", err)
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
		return diag.Errorf("could not delete drift detection integration for stack: %v", err)
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

	// Schedule(s) has/have to be set for the drift detection integration
	if len(integration.Schedule) == 0 {
		return onNil("drift detection integration not found")
	}

	d.Set("reconcile", integration.Reconcile)
	d.Set("timezone", integration.Timezone)

	schedule := make([]interface{}, len(integration.Schedule))
	for i, expr := range integration.Schedule {
		schedule[i] = expr
	}
	if err := d.Set("schedule", schedule); err != nil {
		return diag.Errorf("error setting schedule (resource): %v", err)
	}

	return nil
}
