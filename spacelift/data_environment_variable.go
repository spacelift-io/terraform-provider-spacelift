package spacelift

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
)

func dataEnvironmentVariable() *schema.Resource {
	return &schema.Resource{
		Read: dataEnvironmentVariableRead,

		Schema: map[string]*schema.Schema{
			"checksum": &schema.Schema{
				Type:        schema.TypeString,
				Description: "SHA-256 checksum of the value",
				Computed:    true,
			},
			"context_id": &schema.Schema{
				Type:          schema.TypeString,
				Description:   "ID of the context on which the environment variable is defined",
				Optional:      true,
				ConflictsWith: []string{"stack_id"},
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Name of the environment variable",
				Required:    true,
			},
			"stack_id": &schema.Schema{
				Type:          schema.TypeString,
				Description:   "ID of the stack on which the environment variable is defined",
				Optional:      true,
				ConflictsWith: []string{"context_id", "module_id"},
			},
			"module_id": &schema.Schema{
				Type:          schema.TypeString,
				Description:   "ID of the module on which the environment variable is defined",
				Optional:      true,
				ConflictsWith: []string{"context_id", "stack_id"},
			},
			"value": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Value of the environment variable",
				Sensitive:   true,
				Computed:    true,
			},
			"write_only": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Indicates whether the value can be read back outside a Run",
				Computed:    true,
			},
		},
	}
}

func dataEnvironmentVariableRead(d *schema.ResourceData, meta interface{}) error {
	_, contextOK := d.GetOk("context_id")
	_, moduleOK := d.GetOk("module_id")
	_, stackOK := d.GetOk("stack_id")

	if !(contextOK || moduleOK || stackOK) {
		return errors.New("either context_id or stack_id/module_id must be provided")
	}

	if contextOK {
		return dataEnvironmentVariableReadContext(d, meta)
	}

	if moduleOK {
		return dataEnvironmentVariableReadModule(d, meta)
	}

	return dataEnvironmentVariableReadStack(d, meta)
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

	if err := meta.(*Client).Query(&query, variables); err != nil {
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

	if err := meta.(*Client).Query(&query, variables); err != nil {
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

	if err := meta.(*Client).Query(&query, variables); err != nil {
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
