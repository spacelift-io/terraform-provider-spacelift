package spacelift

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourceStackDestructor() *schema.Resource {
	deleteTimeout := time.Hour * 2

	return &schema.Resource{
		Description: "" +
			"`spacelift_stack_destructor` is used to destroy the resources of a " +
			"Stack before deleting it. `depends_on` should be used to make sure " +
			"that all necessery resources (environment variables, roles, " +
			"integrations, etc.) are still in place when the destruction run is " +
			"executed.",

		CreateContext: resourceStackDestructorCreate,
		ReadContext:   resourceStackDestructorRead,
		UpdateContext: resourceStackDestructorUpdate,
		DeleteContext: resourceStackDestructorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack to delete and destroy on destruction",
				Required:    true,
				ForceNew:    true,
			},
			"deactivated": {
				Type:        schema.TypeBool,
				Description: "If set to true, destruction won't delete the stack",
				Optional:    true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: &deleteTimeout,
		},
	}
}

func resourceStackDestructorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId(d.Get("stack_id").(string))
	d.Set("deactivated", d.Get("deactivated"))
	d.Set("stack_id", d.Get("stack_id"))

	return resourceStackDestructorRead(ctx, d, meta)
}

func resourceStackDestructorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Get("stack_id").(string))}

	if err := meta.(*internal.Client).Query(ctx, "StackDestructorRead", &query, variables); err != nil {
		return diag.Errorf("could not query for stack: %v", err)
	}

	stack := query.Stack
	if stack == nil {
		d.SetId("")
		return nil
	}

	return nil
}

func resourceStackDestructorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.Set("deactivated", d.Get("deactivated"))

	return resourceStackRead(ctx, d, meta)
}

func resourceStackDestructorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if deactivated := d.Get("deactivated"); deactivated != nil && deactivated.(bool) {
		d.SetId("")
		return nil
	}

	var mutation struct {
		DeleteStack *structs.Stack `graphql:"stackDelete(id: $id, destroyResources: true)"`
	}

	variables := map[string]interface{}{"id": toID(d.Get("stack_id").(string))}

	if err := meta.(*internal.Client).Mutate(ctx, "StackDestructorDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete stack: %v", internal.FromSpaceliftError(err))
	}

	if mutation.DeleteStack != nil && mutation.DeleteStack.Deleting {
		if diagnostics := waitForDestroy(ctx, meta.(*internal.Client), d.Get("stack_id").(string)); diagnostics.HasError() {
			return diagnostics
		}
	}

	d.SetId("")

	return nil
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
			return diag.Errorf("could not query for stack: %v", err)
		}

		stack := query.Stack
		if stack == nil {
			return nil
		}

		if !stack.Deleting {
			return diag.Errorf("destruction of Stack unsuccessful, please check the destruction run logs")
		}
	}
}
