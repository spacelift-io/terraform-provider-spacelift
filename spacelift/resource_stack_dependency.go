package spacelift

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourceStackDependency() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_stack_dependency` represents a Spacelift **stack dependency** - " +
			"a dependency between two stacks. When one stack depends on another, the tracked runs " +
			"of the stack will not start until the dependent stack is successfully finished.",

		CreateContext: resourceStackDependencyCreate,
		ReadContext:   resourceStackDependencyRead,
		UpdateContext: resourceStackDependencyUpdate,
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
			"triggers": {
				Type:        schema.TypeBool,
				Description: "describes whether we should trigger the dependent if it's not triggered by the push, but the current stack has changed. Defaults to `true`.",
				Required:    true,
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

	d.SetId(fmt.Sprintf("%s|%s", query.StackDependency.ID, query.StackDependency.StackID))

	return resourceStackDependencyRead(ctx, d, meta)
}

func resourceStackDependencyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		StackDependency structs.StackDependency `graphql:"stackDependencyUpdate(id: $id, triggers: $triggers)"`
	}

	variables := map[string]interface{}{
		"id":       getStackDependencyId(d),
		"triggers": toBool(d.Get("triggers")),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "StackDependencyUpdate", &query, variables); err != nil {
		return diag.Errorf("could not update stack dependency: %s", err)
	}

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

	d.Set("stack_id", query.Stack.Dependency.StackID)
	d.Set("depends_on_stack_id", query.Stack.Dependency.DependsOnStackID)
	d.Set("triggers", query.Stack.Dependency.Triggers)

	return nil
}

func resourceStackDependencyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "|")
	d.Set("stack_id", idParts[1])

	resourceStackDependencyRead(ctx, d, meta)

	return []*schema.ResourceData{d}, nil
}

func getStackDependencyId(d *schema.ResourceData) graphql.ID {
	idParts := strings.Split(d.Id(), "|")

	return graphql.ID(idParts[0])
}

func stackDependencyCreateInput(d *schema.ResourceData) structs.StackDependencyInput {
	return structs.StackDependencyInput{
		StackID:          toID(d.Get("stack_id")),
		DependsOnStackID: toID(d.Get("depends_on_stack_id")),
		Triggers:         toBool(d.Get("triggers")),
	}
}
