package spacelift

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
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
				ForceNew:      true,
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Name of the environment variable",
				Required:    true,
				ForceNew:    true,
			},
			"stack_id": &schema.Schema{
				Type:          schema.TypeString,
				Description:   "ID of the stack on which the environment variable is defined",
				Optional:      true,
				ConflictsWith: []string{"context_id"},
				ForceNew:      true,
			},
			"value": &schema.Schema{
				Type:             schema.TypeString,
				Description:      "Value of the environment variable",
				DiffSuppressFunc: suppressValueChange,
				Sensitive:        true,
				Required:         true,
				ForceNew:         true,
			},
			"write_only": &schema.Schema{
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

	if contextOK == stackOK {
		return errors.New("either context_id or stack_id must be provided")
	}

	if contextOK {
		return resourceEnvironmentVariableCreateContext(d, meta.(*Client), variables)
	}

	return resourceEnvironmentVariableCreateStack(d, meta.(*Client), variables)
}

func resourceEnvironmentVariableCreateContext(d *schema.ResourceData, client *Client, variables map[string]interface{}) error {
	var mutation struct {
		AddContextConfig structs.ConfigElement `graphql:"contextConfigAdd(context: $id, config: $input)"`
	}

	if err := client.Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not create context environment variable")
	}

	if d.Get("write_only").(bool) {
		d.Set("value", "")
	}

	d.SetId(fmt.Sprintf("context/%s/%s", d.Get("context_id"), d.Get("name")))
	return resourceEnvironmentVariableRead(d, client)
}

func resourceEnvironmentVariableCreateStack(d *schema.ResourceData, client *Client, variables map[string]interface{}) error {
	var mutation struct {
		AddStackConfig structs.ConfigElement `graphql:"stackConfigAdd(stack: $id, config: $input)"`
	}

	if err := client.Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not create stack environment variable")
	}

	if d.Get("write_only").(bool) {
		d.Set("value", "")
	}

	d.SetId(fmt.Sprintf("stack/%s/%s", d.Get("stack_id"), d.Get("name")))
	return resourceEnvironmentVariableRead(d, client)
}

func resourceEnvironmentVariableRead(d *schema.ResourceData, meta interface{}) error {
	idParts := strings.SplitN(d.Id(), "/", 3)
	if len(idParts) != 3 {
		return errors.Errorf("unexpected resource ID: %s", d.Id())
	}

	client := meta.(*Client)
	var element *structs.ConfigElement
	var err error

	switch idParts[0] {
	case "context":
		element, err = resourceEnvironmentVariableReadContext(d, client, toID(idParts[1]), toID(idParts[2]))
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

	if element.WriteOnly {
		d.Set("value", nil)
	}

	d.Set("checksum", element.Checksum)
	return nil
}

func resourceEnvironmentVariableReadContext(d *schema.ResourceData, client *Client, context graphql.ID, ID graphql.ID) (*structs.ConfigElement, error) {
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

func resourceEnvironmentVariableReadStack(d *schema.ResourceData, client *Client, stack graphql.ID, ID graphql.ID) (*structs.ConfigElement, error) {
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

	client := meta.(*Client)
	var err error

	switch idParts[0] {
	case "context":
		err = resourceEnvironmentVariableDeleteContext(d, client, toID(idParts[1]), toID(idParts[2]))
	case "stack":
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

func resourceEnvironmentVariableDeleteContext(d *schema.ResourceData, client *Client, context graphql.ID, ID graphql.ID) error {
	var mutation struct {
		DeleteContextConfig *structs.ConfigElement `graphql:"contextConfigDelete(context: $context, id: $id)"`
	}

	return client.Mutate(&mutation, map[string]interface{}{"context": context, "id": ID})
}

func resourceEnvironmentVariableDeleteStack(d *schema.ResourceData, client *Client, stack graphql.ID, ID graphql.ID) error {
	var mutation struct {
		DeleteStackConfig *structs.ConfigElement `graphql:"stackConfigDelete(stack: $stack, id: $id)"`
	}

	return client.Mutate(&mutation, map[string]interface{}{"stack": stack, "id": ID})
}

func suppressValueChange(_, _, value string, d *schema.ResourceData) bool {
	checksum, ok := d.GetOk("checksum")
	if !ok {
		return false
	}

	oldValueChecksum, _ := hex.DecodeString(checksum.(string))
	newValueChecksum := sha256.Sum256([]byte(value))

	return subtle.ConstantTimeCompare(newValueChecksum[:], oldValueChecksum) == 1
}
