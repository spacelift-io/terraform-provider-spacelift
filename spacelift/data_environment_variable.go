package spacelift

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataEnvironmentVariable() *schema.Resource {
	return &schema.Resource{
		Read: dataEnvironmentVariableRead,

		Schema: map[string]*schema.Schema{
			"checksum": {
				Type:        schema.TypeString,
				Description: "SHA-256 checksum of the value",
				Computed:    true,
			},
			"context_id": {
				Type:          schema.TypeString,
				Description:   "ID of the context on which the environment variable is defined",
				Optional:      true,
				ConflictsWith: []string{"stack_id", "module_id"},
			},
			"module_id": {
				Type:          schema.TypeString,
				Description:   "ID of the module on which the environment variable is defined",
				Optional:      true,
				ConflictsWith: []string{"stack_id"},
			},
			"name": {
				Type:        schema.TypeString,
				Description: "name of the environment variable",
				Required:    true,
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

func dataEnvironmentVariableRead(d *schema.ResourceData, meta interface{}) error {
	if _, ok := d.GetOk("context_id"); ok {
		return dataEnvironmentVariableReadContext(d, meta)
	}

	if _, ok := d.GetOk("module_id"); ok {
		return dataEnvironmentVariableReadModule(d, meta)
	}

	if _, ok := d.GetOk("stack_id"); ok {
		return dataEnvironmentVariableReadStack(d, meta)
	}

	return errors.New("context_id, module_id or stack_id must be provided")
}

func dataEnvironmentVariableReadContext(d *schema.ResourceData, meta interface{}) error {
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

	if err := meta.(*internal.Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for context environment variable")
	}

	if query.Context == nil {
		return errors.New("context not found")
	}

	configElement := query.Context.ConfigElement
	if configElement == nil {
		return errors.New("environment variable not found")
	}

	if configElement.Type != "ENVIRONMENT_VARIABLE" {
		return errors.New("config element is not an environment variable")
	}

	d.SetId(fmt.Sprintf("context/%s/%s", contextID, variableName))

	populateEnvironmentVariable(d, configElement)

	return nil
}

func dataEnvironmentVariableReadModule(d *schema.ResourceData, meta interface{}) error {
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

	if err := meta.(*internal.Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for module environment variable")
	}

	if query.Module == nil {
		return errors.New("module not found")
	}

	if query.Module.ConfigElement == nil {
		return errors.New("environment variable not found")
	}

	d.SetId(fmt.Sprintf("module/%s/%s", moduleID, variableName))

	populateEnvironmentVariable(d, query.Module.ConfigElement)

	return nil
}
func dataEnvironmentVariableReadStack(d *schema.ResourceData, meta interface{}) error {
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

	if err := meta.(*internal.Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for stack environment variable")
	}

	if query.Stack == nil {
		return errors.New("stack not found")
	}

	if query.Stack.ConfigElement == nil {
		return errors.New("environment variable not found")
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
