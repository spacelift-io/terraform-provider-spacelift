package spacelift

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataEnvironmentVariable() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_environment_variable` defines an environment variable on " +
			"the context (`spacelift_context`), stack (`spacelift_stack`) or a " +
			"module (`spacelift_module`), thereby allowing to pass and share " +
			"various secrets and configuration with Spacelift stacks.",

		ReadContext: dataEnvironmentVariableRead,

		Schema: map[string]*schema.Schema{
			"checksum": {
				Type:        schema.TypeString,
				Description: "SHA-256 checksum of the value",
				Computed:    true,
			},
			"context_id": {
				Type:         schema.TypeString,
				Description:  "ID of the context on which the environment variable is defined",
				ExactlyOneOf: []string{"context_id", "stack_id", "module_id"},
				Optional:     true,
			},
			"module_id": {
				Type:        schema.TypeString,
				Description: "ID of the module on which the environment variable is defined",
				Optional:    true,
			},
			"name": {
				Type:             schema.TypeString,
				Description:      "name of the environment variable",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack on which the environment variable is defined",
				Optional:    true,
			},
			"value": {
				Type:        schema.TypeString,
				Description: "value of the environment variable",
				Sensitive:   true,
				Computed:    true,
			},
			"write_only": {
				Type:        schema.TypeBool,
				Description: "indicates whether the value can be read back outside a Run",
				Computed:    true,
			},
		},
	}
}

func dataEnvironmentVariableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if _, ok := d.GetOk("context_id"); ok {
		return dataEnvironmentVariableReadContext(ctx, d, meta)
	}

	if _, ok := d.GetOk("module_id"); ok {
		return dataEnvironmentVariableReadModule(ctx, d, meta)
	}

	return dataEnvironmentVariableReadStack(ctx, d, meta)
}

func dataEnvironmentVariableReadContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Context *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"context(id: $context)"`
	}

	contextID := d.Get("context_id")
	variableName := d.Get("name")

	variables := map[string]interface{}{
		"context": toID(contextID),
		"id":      toID(variableName),
	}

	if err := meta.(*internal.Client).Query(ctx, "EnvironmentVariableReadContext", &query, variables); err != nil {
		return diag.Errorf("could not query for context environment variable: %v", err)
	}

	if query.Context == nil {
		return diag.Errorf("context not found")
	}

	configElement := query.Context.ConfigElement
	if configElement == nil {
		return diag.Errorf("environment variable not found")
	}

	if configElement.Type != "ENVIRONMENT_VARIABLE" {
		return diag.Errorf("config element is not an environment variable")
	}

	d.SetId(fmt.Sprintf("context/%s/%s", contextID, variableName))

	populateEnvironmentVariable(d, configElement)

	return nil
}

func dataEnvironmentVariableReadModule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Module *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"module(id: $module)"`
	}

	moduleID := d.Get("module_id")
	variableName := d.Get("name")

	variables := map[string]interface{}{
		"module": toID(moduleID),
		"id":     toID(variableName),
	}

	if err := meta.(*internal.Client).Query(ctx, "EnvironmentVariableReadModule", &query, variables); err != nil {
		return diag.Errorf("could not query for module environment variable: %v", err)
	}

	if query.Module == nil {
		return diag.Errorf("module not found")
	}

	if query.Module.ConfigElement == nil {
		return diag.Errorf("environment variable not found")
	}

	d.SetId(fmt.Sprintf("module/%s/%s", moduleID, variableName))

	populateEnvironmentVariable(d, query.Module.ConfigElement)

	return nil
}

func dataEnvironmentVariableReadStack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Stack *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"stack(id: $stack)"`
	}

	stackID := d.Get("stack_id")
	variableName := d.Get("name")

	variables := map[string]interface{}{
		"stack": toID(stackID),
		"id":    toID(variableName),
	}

	if err := meta.(*internal.Client).Query(ctx, "EnvironmentVariableReadStack", &query, variables); err != nil {
		return diag.Errorf("could not query for stack environment variable: %v", err)
	}

	if query.Stack == nil {
		return diag.Errorf("stack not found")
	}

	if query.Stack.ConfigElement == nil {
		return diag.Errorf("environment variable not found")
	}

	d.SetId(fmt.Sprintf("stack/%s/%s", stackID, variableName))

	populateEnvironmentVariable(d, query.Stack.ConfigElement)

	return nil
}

func populateEnvironmentVariable(d *schema.ResourceData, el *structs.ConfigElement) {
	d.Set("checksum", el.Checksum)
	d.Set("write_only", el.WriteOnly)

	if el.Value != nil {
		d.Set("value", *el.Value)
	} else {
		d.Set("value", nil)
	}
}
