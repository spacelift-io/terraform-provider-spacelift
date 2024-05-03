package spacelift

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourceStackDependencyReference() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_stack_dependency_reference` represents a Spacelift **stack dependency reference** - " +
			"a reference matches a stack's output to another stack's input. It is similar to an environment variable " +
			"(`spacelift_environment_variable`), except that value is provided by another stack's output.",

		CreateContext: resourceStackDependencyReferenceCreate,
		ReadContext:   resourceStackDependencyReferenceRead,
		UpdateContext: resourceStackDependencyReferenceUpdate,
		DeleteContext: resourceStackDependencyReferenceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"stack_dependency_id": {
				Type:        schema.TypeString,
				Description: "Immutable ID of stack dependency",
				Required:    true,
				ForceNew:    true,
			},
			"output_name": {
				Type:        schema.TypeString,
				Description: "Name of the output of stack to depend on",
				Required:    true,
			},
			"input_name": {
				Type:        schema.TypeString,
				Description: "Name of the input of the stack dependency reference",
				Required:    true,
			},
			"trigger_always": {
				Type:        schema.TypeBool,
				Description: "Whether the dependents should be triggered even if the value of the reference did not change.",
				Default:     false,
				Optional:    true,
			},
		},
	}
}

func resourceStackDependencyReferenceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		StackDependencyReference structs.StackDependencyReference `graphql:"stackDependenciesAddReference(stackDependencyID: $stackDependencyID, reference: $reference)"`
	}

	stackID, depID, diags := getStackDependencyIDParts(d)
	if diags != nil {
		return diags
	}

	variables := map[string]interface{}{
		"stackDependencyID": toID(depID),
		"reference": structs.StackDependencyReferenceInput{
			OutputName:    toString(d.Get("output_name")),
			InputName:     toString(d.Get("input_name")),
			Type:          toString("ENVIRONMENT_VARIABLE"),
			TriggerAlways: toBool(d.Get("trigger_always")),
		},
	}

	if err := meta.(*internal.Client).Mutate(ctx, "StackDependenciesAddReference", &query, variables); err != nil {
		return diag.Errorf("could not create stack dependency reference: %s", err)
	}

	d.SetId(path.Join(stackID, depID, query.StackDependencyReference.ID))
	d.Set("output_name", query.StackDependencyReference.OutputName)
	d.Set("input_name", query.StackDependencyReference.InputName)
	d.Set("trigger_always", query.StackDependencyReference.TriggerAlways)

	return nil
}

func resourceStackDependencyReferenceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Stack *struct {
			Dependency *struct {
				Reference *structs.StackDependencyReference `graphql:"reference(id: $reference_id)"`
			} `graphql:"dependency(id: $dependency_id)"`
		} `graphql:"stack(id: $stack_id)"`
	}

	stackID, depID, refID, diags := getStackDependencyReferenceIDParts(d)
	if diags != nil {
		return diags
	}

	variables := map[string]interface{}{
		"stack_id":      toID(stackID),
		"dependency_id": toID(depID),
		"reference_id":  toID(refID),
	}

	if err := meta.(*internal.Client).Query(ctx, "StackDependenciesReferenceRead", &query, variables); err != nil {
		return diag.Errorf("could not query for stack dependency reference: %s", err)
	}

	var nonExistenceWarning string

	if query.Stack == nil {
		nonExistenceWarning = fmt.Sprintf("could not find stack (%s), maybe it was deleted manually", stackID)
	} else if query.Stack.Dependency == nil {
		nonExistenceWarning = fmt.Sprintf("could not find stack dependency (%s), maybe it was deleted manually", depID)
	} else if query.Stack.Dependency.Reference == nil {
		nonExistenceWarning = fmt.Sprintf("could not find stack dependency reference (%s), maybe it was deleted manually", refID)
	}

	if nonExistenceWarning != "" {
		d.SetId("")

		return diag.Diagnostics{{
			Severity: diag.Warning,
			Summary:  nonExistenceWarning,
		}}
	}

	d.Set("stack_dependency_id", path.Join(stackID, depID))
	d.Set("output_name", query.Stack.Dependency.Reference.OutputName)
	d.Set("input_name", query.Stack.Dependency.Reference.InputName)
	d.Set("trigger_always", query.Stack.Dependency.Reference.TriggerAlways)

	return nil
}

func resourceStackDependencyReferenceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		StackDependencyReference structs.StackDependencyReference `graphql:"stackDependenciesUpdateReference(reference: $reference)"`
	}

	stackID, depID, refID, diags := getStackDependencyReferenceIDParts(d)
	if diags != nil {
		return diags
	}

	variables := map[string]interface{}{
		"reference": structs.StackDependencyReferenceUpdateInput{
			ID:            toID(refID),
			OutputName:    toString(d.Get("output_name")),
			InputName:     toString(d.Get("input_name")),
			Type:          toString("ENVIRONMENT_VARIABLE"),
			TriggerAlways: toBool(d.Get("trigger_always")),
		},
	}

	if err := meta.(*internal.Client).Mutate(ctx, "StackDependenciesUpdateReference", &query, variables); err != nil {
		return diag.Errorf("could not update stack dependency reference: %s", err)
	}

	d.Set("stack_dependency_id", path.Join(stackID, depID))
	d.Set("output_name", query.StackDependencyReference.OutputName)
	d.Set("input_name", query.StackDependencyReference.InputName)
	d.Set("trigger_always", query.StackDependencyReference.TriggerAlways)

	return nil
}

func resourceStackDependencyReferenceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		StackDependencyReference *structs.StackDependencyReference `graphql:"stackDependenciesDeleteReference(id: $id)"`
	}

	_, _, refID, diags := getStackDependencyReferenceIDParts(d)
	if diags != nil {
		return diags
	}

	variables := map[string]interface{}{
		"id": toID(refID),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "StackDependenciesDeleteReference", &query, variables); err != nil {
		return diag.Errorf("could not delete stack dependency reference: %s", err)
	}

	d.SetId("")
	return nil
}

func getStackDependencyIDParts(d *schema.ResourceData) (string, string, diag.Diagnostics) {
	stackDependencyID := string(toString(d.Get("stack_dependency_id")))

	idParts := strings.SplitN(stackDependencyID, "/", 2)
	if len(idParts) != 2 {
		return "", "", diag.Errorf("unexpected stack_dependency_id: %s", stackDependencyID)
	}

	return idParts[0], idParts[1], nil
}

func getStackDependencyReferenceIDParts(d *schema.ResourceData) (string, string, string, diag.Diagnostics) {
	idParts := strings.SplitN(d.Id(), "/", 3)
	if len(idParts) != 3 {
		return "", "", "", diag.Errorf("unexpected resource ID: %s", d.Id())
	}

	return idParts[0], idParts[1], idParts[2], nil
}
