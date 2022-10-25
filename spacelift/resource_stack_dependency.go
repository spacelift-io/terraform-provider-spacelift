package spacelift

import (
	"context"
	"path"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourceStackDependency() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"**Note:** This resource is under development. Please do not use it yet. " +
			"\n\n" +
			"`spacelift_stack_dependency` represents a Spacelift **stack dependency** - " +
			"a dependency between two stacks. When one stack depends on another, the tracked runs " +
			"of the stack will not start until the dependent stack is successfully finished. Additionally, " +
			"changes to the dependency will trigger the dependent.",

		CreateContext: resourceStackDependencyCreate,
		ReadContext:   resourceStackDependencyRead,
		DeleteContext: resourceStackDependencyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceStackDependencyImport,
		},

		Schema: map[string]*schema.Schema{
			"stack_id": {
				Type:        schema.TypeString,
				Description: "immutable ID (slug) of stack which has a dependency.",
				Required:    true,
				ForceNew:    true,
			},
			"depends_on_stack_id": {
				Type:        schema.TypeString,
				Description: "immutable ID (slug) of stack to depend on.",
				Required:    true,
				ForceNew:    true,
			},
		},
	}

}

func resourceStackDependencyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		StackDependency structs.StackDependency `graphql:"stackDependencyCreate(input: $input)"`
	}

	variables := map[string]interface{}{"input": stackDependencyCreateInput(d)}

	if err := meta.(*internal.Client).Mutate(ctx, "StackDependencyCreate", &query, variables); err != nil {
		return diag.Errorf("could not create stack dependency: %s", err)
	}

	d.SetId(path.Join(query.StackDependency.Stack.ID, query.StackDependency.ID))

	return resourceStackDependencyRead(ctx, d, meta)
}

func resourceStackDependencyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		StackDependency *structs.StackDependency `graphql:"stackDependencyDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": getStackDependencyId(d)}

	if err := meta.(*internal.Client).Mutate(ctx, "StackDependencyDelete", &query, variables); err != nil {
		return diag.Errorf("could not delete stack dependency: %s", err)
	}

	d.SetId("")

	return nil
}

func resourceStackDependencyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Stack *struct {
			Dependency *structs.StackDependency `graphql:"dependency(id: $id)"`
		} `graphql:"stack(id: $stack)"`
	}

	variables := map[string]interface{}{
		"id":    getStackDependencyId(d),
		"stack": toID(d.Get("stack_id")),
	}

	if err := meta.(*internal.Client).Query(ctx, "StackDependencyRead", &query, variables); err != nil {
		return diag.Errorf("could not query for stack dependency: %s", err)
	}

	if query.Stack == nil {
		return nil
	}

	d.Set("stack_id", query.Stack.Dependency.Stack.ID)
	d.Set("depends_on_stack_id", query.Stack.Dependency.DependsOnStack.ID)

	return nil
}

func resourceStackDependencyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "/")
	d.Set("stack_id", idParts[0])

	resourceStackDependencyRead(ctx, d, meta)

	return []*schema.ResourceData{d}, nil
}

func getStackDependencyId(d *schema.ResourceData) graphql.ID {
	idParts := strings.Split(d.Id(), "/")

	return graphql.ID(idParts[1])
}

func stackDependencyCreateInput(d *schema.ResourceData) structs.StackDependencyInput {
	return structs.StackDependencyInput{
		StackID:          toID(d.Get("stack_id")),
		DependsOnStackID: toID(d.Get("depends_on_stack_id")),
	}
}
