package spacelift

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/oklog/ulid/v2"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourceStackDependency() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_stack_dependency` represents a Spacelift **stack dependency** - " +
			"a dependency between two stacks. When one stack depends on another, the tracked runs " +
			"of the stack will not start until the dependent stack is successfully finished. Additionally, " +
			"changes to the dependency will trigger the dependent.\n\n" +
			"~> **Import format**: Use `terraform import spacelift_stack_dependency.example stack-id/depends-on-stack-id`. " +
			"The old format `stack-id/dependency-ulid` is deprecated but still supported for backward compatibility.",

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

	_, depID, err := parseStackDependencyID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	variables := map[string]any{"id": graphql.ID(depID)}

	if err := meta.(*internal.Client).Mutate(ctx, "StackDependencyDelete", &query, variables); err != nil {
		return diag.Errorf("could not delete stack dependency: %s", err)
	}

	d.SetId("")

	return nil
}

func resourceStackDependencyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	stackID := toID(d.Get("stack_id"))
	dependsOnStackID := toID(d.Get("depends_on_stack_id"))

	dependency, err := resourceStackDependencyFetchByStackID(ctx, meta, stackID, dependsOnStackID)
	if err != nil {
		return diag.Errorf("could not query for stack dependency: %s", err)
	}

	if dependency == nil {
		d.SetId("")
		return nil
	}

	d.Set("stack_id", dependency.Stack.ID)
	d.Set("depends_on_stack_id", dependency.DependsOnStack.ID)
	d.SetId(path.Join(dependency.Stack.ID, dependency.ID))

	return nil
}

func resourceStackDependencyImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	stackID, secondPart, err := parseStackDependencyID(d.Id())
	if err != nil {
		return nil, fmt.Errorf("invalid import ID format: %w", err)
	}

	var dependency *structs.StackDependency

	// Try to parse as ULID to detect old format (stackID/dependencyULID)
	_, err = ulid.Parse(secondPart)
	if err == nil {
		dependency, err = resourceStackDependencyFetchByID(ctx, meta, toID(stackID), toID(secondPart))
	} else {
		dependency, err = resourceStackDependencyFetchByStackID(ctx, meta, toID(stackID), toID(secondPart))
	}

	if err != nil {
		return nil, err
	}

	if dependency == nil {
		return nil, fmt.Errorf("stack dependency not found for import ID: %s", path.Join(stackID, secondPart))
	}

	d.Set("stack_id", dependency.Stack.ID)
	d.Set("depends_on_stack_id", dependency.DependsOnStack.ID)
	d.SetId(path.Join(dependency.Stack.ID, dependency.ID))

	return []*schema.ResourceData{d}, nil
}

// resourceStackDependencyFetchByID fetches a dependency using ULID (old format)
func resourceStackDependencyFetchByID(ctx context.Context, meta any, stackID, dependencyID graphql.ID) (*structs.StackDependency, error) {
	variables := map[string]any{
		"id":               dependencyID,
		"stackId":          stackID,
		"dependsOnStackId": toID(""),
	}
	return resourceStackDependencyFetch(ctx, meta, variables)
}

// resourceStackDependencyFetchByStackID fetches a dependency using stack ID (new format)
func resourceStackDependencyFetchByStackID(ctx context.Context, meta any, stackID, dependsOnStackID graphql.ID) (*structs.StackDependency, error) {
	variables := map[string]any{
		"id":               toID(""),
		"stackId":          stackID,
		"dependsOnStackId": dependsOnStackID,
	}
	return resourceStackDependencyFetch(ctx, meta, variables)
}

func resourceStackDependencyFetch(ctx context.Context, meta any, variables map[string]any) (*structs.StackDependency, error) {
	var query struct {
		Stack *struct {
			Dependency *structs.StackDependency `graphql:"dependency(id: $id, dependsOnStackId: $dependsOnStackId)"`
		} `graphql:"stack(id: $stackId)"`
	}

	if err := meta.(*internal.Client).Query(ctx, "StackDependencyRead", &query, variables); err != nil {
		return nil, err
	}

	if query.Stack == nil {
		return nil, fmt.Errorf("stack not found")
	}

	return query.Stack.Dependency, nil
}

func parseStackDependencyID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("unexpected resource ID format: %s", id)
	}

	return idParts[0], idParts[1], nil
}

func stackDependencyCreateInput(d *schema.ResourceData) structs.StackDependencyInput {
	return structs.StackDependencyInput{
		StackID:          toID(d.Get("stack_id")),
		DependsOnStackID: toID(d.Get("depends_on_stack_id")),
	}
}
