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

func resourceStackActivator() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_stack_activator` is used to to enable/disable Spacelift Stack.",
		CreateContext: resourceStackActivatorCreate,
		ReadContext:   resourceStackActivatorRead,
		UpdateContext: resourceStackActivatorUpdate,
		DeleteContext: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"stack_id": {
				Type:             schema.TypeString,
				Description:      "ID of the stack to enable/disable",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Enable/disable stack",
				Required:    true,
			},
		},
	}
}

func resourceStackActivatorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceStackActivatorUpdate(ctx, d, meta)
}

func resourceStackActivatorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	stack, err := queryStack(ctx, d, meta)
	if err != nil {
		return diag.Errorf("failed to retrieve stack: %v", internal.FromSpaceliftError(err))
	}
	if stack == nil {
		return diag.Errorf("stack not found: %v", internal.FromSpaceliftError(err))
	}
	d.SetId(fmt.Sprintf("activator-%d", time.Now().Unix()))
	d.Set("enabled", !stack.IsDisabled)
	return nil
}

func resourceStackActivatorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	stack, err := queryStack(ctx, d, meta)
	if err != nil {
		return diag.Errorf("failed to retrieve stack: %v", internal.FromSpaceliftError(err))
	}
	if stack == nil {
		return diag.Errorf("stack not found: %v", internal.FromSpaceliftError(err))
	}
	d.SetId(fmt.Sprintf("activator-%d", time.Now().Unix()))
	enabled, ok := d.Get("enabled").(bool)
	if !ok {
		return diag.Errorf("invalid enabled attribute: %v", d.Get("enabled"))
	}
	if !stack.IsDisabled == enabled {
		return nil
	}
	if enabled {
		return enableStack(ctx, d, meta)
	}
	return disableStack(ctx, d, meta)
}

func queryStack(ctx context.Context, d *schema.ResourceData, meta interface{}) (*structs.Stack, error) {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}
	variables := map[string]interface{}{"id": graphql.ID(d.Get("stack_id"))}
	if err := meta.(*internal.Client).Query(ctx, "StackActivatorRead", &query, variables); err != nil {
		return nil, err
	}
	return query.Stack, nil
}

func enableStack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		EnableStack *structs.Stack `graphql:"stackEnable(id: $id)"`
	}
	stackID, ok := d.Get("stack_id").(string)
	if !ok {
		return diag.Errorf("invalid stack ID")
	}
	variables := map[string]interface{}{"id": toID(stackID)}
	if err := meta.(*internal.Client).Mutate(ctx, "StackActivatorEnable", &mutation, variables); err != nil {
		return diag.Errorf("could not enable stack %s: %v", stackID, internal.FromSpaceliftError(err))
	}
	d.Set("enabled", true)
	return nil
}

func disableStack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		EnableStack *structs.Stack `graphql:"stackDisable(id: $id)"`
	}
	stackID, ok := d.Get("stack_id").(string)
	if !ok {
		return diag.Errorf("invalid stack ID")
	}
	variables := map[string]interface{}{"id": toID(stackID)}
	if err := meta.(*internal.Client).Mutate(ctx, "StackActivatorDisable", &mutation, variables); err != nil {
		return diag.Errorf("could not enable stack %s: %v", stackID, internal.FromSpaceliftError(err))
	}
	d.Set("enabled", false)
	return nil
}
