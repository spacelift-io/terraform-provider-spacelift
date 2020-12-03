package spacelift

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourceEnvironmentVariable() *schema.Resource {
	return &schema.Resource{
		Create: resourceEnvironmentVariableCreate,
		Read:   resourceEnvironmentVariableRead,
		Delete: resourceEnvironmentVariableDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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
				ForceNew:      true,
			},
			"module_id": {
				Type:          schema.TypeString,
				Description:   "ID of the module on which the environment variable is defined",
				Optional:      true,
				ConflictsWith: []string{"stack_id"},
				ForceNew:      true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the environment variable",
				Required:    true,
				ForceNew:    true,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack on which the environment variable is defined",
				Optional:    true,
				ForceNew:    true,
			},
			"value": {
				Type:             schema.TypeString,
				Description:      "Value of the environment variable",
				DiffSuppressFunc: suppressValueChange,
				Sensitive:        true,
				Required:         true,
				ForceNew:         true,
			},
			"write_only": {
				Type:        schema.TypeBool,
				Description: "Indicates whether the value can be read back outside a Run",
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
		},
	}
}

func resourceEnvironmentVariableCreate(d *schema.ResourceData, meta interface{}) error {
	variables := map[string]interface{}{
		"config": structs.ConfigInput{
			ID:        toID(d.Get("name")),
			Type:      structs.ConfigType("ENVIRONMENT_VARIABLE"),
			Value:     toString(d.Get("value")),
			WriteOnly: graphql.Boolean(d.Get("write_only").(bool)),
		},
	}

	contextID, contextOK := d.GetOk("context_id")
	if contextOK {
		variables["context"] = toID(contextID)
	}

	stackID, stackOK := d.GetOk("stack_id")
	if stackOK {
		variables["stack"] = toID(stackID)
	}

	moduleID, moduleOK := d.GetOk("module_id")
	if moduleOK {
		variables["stack"] = toID(moduleID)
	}

	if !(contextOK || stackOK || moduleOK) {
		return errors.New("either context_id or stack_id/module_id must be provided")
	}

	if contextOK {
		return resourceEnvironmentVariableCreateContext(d, meta.(*internal.Client), variables)
	}

	if moduleOK {
		return resourceEnvironmentVariableCreateModule(d, meta.(*internal.Client), variables)
	}

	return resourceEnvironmentVariableCreateStack(d, meta.(*internal.Client), variables)
}

func resourceEnvironmentVariableCreateContext(d *schema.ResourceData, client *internal.Client, variables map[string]interface{}) error {
	var mutation struct {
		AddContextConfig structs.ConfigElement `graphql:"contextConfigAdd(context: $context, config: $config)"`
	}

	if err := client.Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not create context environment variable")
	}

	d.SetId(fmt.Sprintf("context/%s/%s", d.Get("context_id"), d.Get("name")))

	return resourceEnvironmentVariableRead(d, client)
}

func resourceEnvironmentVariableCreateModule(d *schema.ResourceData, client *internal.Client, variables map[string]interface{}) error {
	var mutation struct {
		AddModuleConfig structs.ConfigElement `graphql:"stackConfigAdd(stack: $stack, config: $config)"`
	}

	if err := client.Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not create module environment variable")
	}

	d.SetId(fmt.Sprintf("module/%s/%s", d.Get("module_id"), d.Get("name")))

	return resourceEnvironmentVariableRead(d, client)
}

func resourceEnvironmentVariableCreateStack(d *schema.ResourceData, client *internal.Client, variables map[string]interface{}) error {
	var mutation struct {
		AddStackConfig structs.ConfigElement `graphql:"stackConfigAdd(stack: $stack, config: $config)"`
	}

	if err := client.Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not create stack environment variable")
	}

	d.SetId(fmt.Sprintf("stack/%s/%s", d.Get("stack_id"), d.Get("name")))

	return resourceEnvironmentVariableRead(d, client)
}

func resourceEnvironmentVariableRead(d *schema.ResourceData, meta interface{}) error {
	idParts := strings.SplitN(d.Id(), "/", 3)
	if len(idParts) != 3 {
		return errors.Errorf("unexpected resource ID: %s", d.Id())
	}

	client := meta.(*internal.Client)
	var element *structs.ConfigElement
	var err error

	switch idParts[0] {
	case "context":
		element, err = resourceEnvironmentVariableReadContext(d, client, toID(idParts[1]), toID(idParts[2]))
	case "module":
		element, err = resourceEnvironmentVariableReadModule(d, client, toID(idParts[1]), toID(idParts[2]))
	case "stack":
		element, err = resourceEnvironmentVariableReadStack(d, client, toID(idParts[1]), toID(idParts[2]))
	default:
		return errors.Errorf("unexpected resource type: %s", idParts[0])
	}

	if err != nil {
		return err
	}

	if element == nil {
		d.SetId("")
		return nil
	}

	d.Set("checksum", element.Checksum)

	if value := element.Value; value != nil {
		d.Set("value", *value)
	} else {
		d.Set("value", element.Checksum)
	}

	return nil
}

func resourceEnvironmentVariableReadContext(d *schema.ResourceData, client *internal.Client, context graphql.ID, ID graphql.ID) (*structs.ConfigElement, error) {
	var query struct {
		Context *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"context(id: $context)"`
	}

	if err := client.Query(&query, map[string]interface{}{"context": context, "id": ID}); err != nil {
		return nil, errors.Wrap(err, "could not query for context environment variable")
	}

	if query.Context == nil {
		return nil, nil
	}

	return query.Context.ConfigElement, nil
}

func resourceEnvironmentVariableReadModule(d *schema.ResourceData, client *internal.Client, module graphql.ID, ID graphql.ID) (*structs.ConfigElement, error) {
	var query struct {
		Module *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"module(id: $module)"`
	}

	if err := client.Query(&query, map[string]interface{}{"module": module, "id": ID}); err != nil {
		return nil, errors.Wrap(err, "could not query for module environment variable")
	}

	if query.Module == nil {
		return nil, nil
	}

	return query.Module.ConfigElement, nil
}

func resourceEnvironmentVariableReadStack(d *schema.ResourceData, client *internal.Client, stack graphql.ID, ID graphql.ID) (*structs.ConfigElement, error) {
	var query struct {
		Stack *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"stack(id: $stack)"`
	}

	if err := client.Query(&query, map[string]interface{}{"stack": stack, "id": ID}); err != nil {
		return nil, errors.Wrap(err, "could not query for stack environment variable")
	}

	if query.Stack == nil {
		return nil, nil
	}

	return query.Stack.ConfigElement, nil
}

func resourceEnvironmentVariableDelete(d *schema.ResourceData, meta interface{}) error {
	idParts := strings.SplitN(d.Id(), "/", 3)
	if len(idParts) != 3 {
		return errors.Errorf("unexpected resource ID: %s", d.Id())
	}

	client := meta.(*internal.Client)
	var err error

	switch idParts[0] {
	case "context":
		err = resourceEnvironmentVariableDeleteContext(d, client, toID(idParts[1]), toID(idParts[2]))
	case "module", "stack":
		err = resourceEnvironmentVariableDeleteStack(d, client, toID(idParts[1]), toID(idParts[2]))
	default:
		return errors.Errorf("unexpected resource type: %s", idParts[0])
	}

	if err != nil {
		return errors.Wrap(err, "could not delete environment variable")
	}

	d.SetId("")
	return nil
}

func resourceEnvironmentVariableDeleteContext(d *schema.ResourceData, client *internal.Client, context graphql.ID, ID graphql.ID) error {
	var mutation struct {
		DeleteContextConfig *structs.ConfigElement `graphql:"contextConfigDelete(context: $context, id: $id)"`
	}

	return client.Mutate(&mutation, map[string]interface{}{"context": context, "id": ID})
}

func resourceEnvironmentVariableDeleteStack(d *schema.ResourceData, client *internal.Client, stack graphql.ID, ID graphql.ID) error {
	var mutation struct {
		DeleteStackConfig *structs.ConfigElement `graphql:"stackConfigDelete(stack: $stack, id: $id)"`
	}

	return client.Mutate(&mutation, map[string]interface{}{"stack": stack, "id": ID})
}

func suppressValueChange(_, old, new string, d *schema.ResourceData) bool {
	oldValueChecksum, err := hex.DecodeString(old)
	if err != nil {
		// This is normal if the value is plaintext.
		return false
	}

	newValueChecksum := sha256.Sum256([]byte(new))

	return subtle.ConstantTimeCompare(newValueChecksum[:], oldValueChecksum) == 1
}
