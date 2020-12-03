package spacelift

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourceMountedFile() *schema.Resource {
	return &schema.Resource{
		Create: resourceMountedFileCreate,
		Read:   resourceMountedFileRead,
		Delete: resourceMountedFileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"checksum": {
				Type:        schema.TypeString,
				Description: "SHA-256 checksum of the value",
				Computed:    true,
			},
			"content": {
				Type:             schema.TypeString,
				Description:      "Content of the mounted file encoded using Base-64",
				DiffSuppressFunc: suppressValueChange,
				Sensitive:        true,
				Required:         true,
				ForceNew:         true,
			},
			"context_id": {
				Type:          schema.TypeString,
				Description:   "ID of the context on which the mounted file is defined",
				Optional:      true,
				ConflictsWith: []string{"module_id", "stack_id"},
				ForceNew:      true,
			},
			"module_id": {
				Type:          schema.TypeString,
				Description:   "ID of the module on which the mounted file is defined",
				Optional:      true,
				ConflictsWith: []string{"stack_id"},
				ForceNew:      true,
			},
			"relative_path": {
				Type:        schema.TypeString,
				Description: "Relative path to the mounted file, without the /spacelift/project/ prefix",
				Required:    true,
				ForceNew:    true,
			},
			"stack_id": {
				Type:          schema.TypeString,
				Description:   "ID of the stack on which the mounted file is defined",
				Optional:      true,
				ConflictsWith: []string{"context_id"},
				ForceNew:      true,
			},
			"write_only": {
				Type:        schema.TypeBool,
				Description: "Indicates whether the content can be read back outside a Run",
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
		},
	}
}

func resourceMountedFileCreate(d *schema.ResourceData, meta interface{}) error {
	variables := map[string]interface{}{
		"config": structs.ConfigInput{
			ID:        toID(d.Get("relative_path")),
			Type:      structs.ConfigType("FILE_MOUNT"),
			Value:     toString(d.Get("content")),
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
		return resourceMountedFileCreateContext(d, meta.(*internal.Client), variables)
	}

	if moduleOK {
		return resourceMountedFileCreateModule(d, meta.(*internal.Client), variables)
	}

	return resourceMountedFileCreateStack(d, meta.(*internal.Client), variables)
}

func resourceMountedFileCreateContext(d *schema.ResourceData, client *internal.Client, variables map[string]interface{}) error {
	var mutation struct {
		AddContextConfig structs.ConfigElement `graphql:"contextConfigAdd(context: $context, config: $config)"`
	}

	if err := client.Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not create context mounted file")
	}

	d.SetId(fmt.Sprintf("context/%s/%s", d.Get("context_id"), d.Get("relative_path")))
	return resourceMountedFileRead(d, client)
}

func resourceMountedFileCreateModule(d *schema.ResourceData, client *internal.Client, variables map[string]interface{}) error {
	var mutation struct {
		AddModuleConfig structs.ConfigElement `graphql:"stackConfigAdd(stack: $stack, config: $config)"`
	}

	if err := client.Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not module mounted file")
	}

	d.SetId(fmt.Sprintf("module/%s/%s", d.Get("module_id"), d.Get("relative_path")))
	return resourceMountedFileRead(d, client)
}

func resourceMountedFileCreateStack(d *schema.ResourceData, client *internal.Client, variables map[string]interface{}) error {
	var mutation struct {
		AddStackConfig structs.ConfigElement `graphql:"stackConfigAdd(stack: $stack, config: $config)"`
	}

	if err := client.Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not create stack mounted file")
	}

	d.SetId(fmt.Sprintf("stack/%s/%s", d.Get("stack_id"), d.Get("relative_path")))
	return resourceMountedFileRead(d, client)
}

func resourceMountedFileRead(d *schema.ResourceData, meta interface{}) error {
	idParts := strings.SplitN(d.Id(), "/", 3)
	if len(idParts) != 3 {
		return errors.Errorf("unexpected resource ID: %s", d.Id())
	}

	client := meta.(*internal.Client)
	var element *structs.ConfigElement
	var err error

	switch idParts[0] {
	case "context":
		element, err = resourceMountedFileReadContext(d, client, toID(idParts[1]), toID(idParts[2]))
	case "module":
		element, err = resourceMountedFileReadModule(d, client, toID(idParts[1]), toID(idParts[2]))
	case "stack":
		element, err = resourceMountedFileReadStack(d, client, toID(idParts[1]), toID(idParts[2]))
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
		d.Set("content", *value)
	} else {
		d.Set("content", element.Checksum)
	}

	return nil
}

func resourceMountedFileReadContext(d *schema.ResourceData, client *internal.Client, context graphql.ID, ID graphql.ID) (*structs.ConfigElement, error) {
	var query struct {
		Context *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"context(id: $context)"`
	}

	if err := client.Query(&query, map[string]interface{}{"context": context, "id": ID}); err != nil {
		return nil, errors.Wrap(err, "could not query for context mounted file")
	}

	if query.Context == nil {
		return nil, nil
	}

	return query.Context.ConfigElement, nil
}

func resourceMountedFileReadModule(d *schema.ResourceData, client *internal.Client, module graphql.ID, ID graphql.ID) (*structs.ConfigElement, error) {
	var query struct {
		Module *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"module(id: $module)"`
	}

	if err := client.Query(&query, map[string]interface{}{"module": module, "id": ID}); err != nil {
		return nil, errors.Wrap(err, "could not query for module mounted file")
	}

	if query.Module == nil {
		return nil, nil
	}

	return query.Module.ConfigElement, nil
}

func resourceMountedFileReadStack(d *schema.ResourceData, client *internal.Client, stack graphql.ID, ID graphql.ID) (*structs.ConfigElement, error) {
	var query struct {
		Stack *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"stack(id: $stack)"`
	}

	if err := client.Query(&query, map[string]interface{}{"stack": stack, "id": ID}); err != nil {
		return nil, errors.Wrap(err, "could not query for stack mounted file")
	}

	if query.Stack == nil {
		return nil, nil
	}

	return query.Stack.ConfigElement, nil
}

func resourceMountedFileDelete(d *schema.ResourceData, meta interface{}) error {
	idParts := strings.SplitN(d.Id(), "/", 3)
	if len(idParts) != 3 {
		return errors.Errorf("unexpected resource ID: %s", d.Id())
	}

	client := meta.(*internal.Client)
	var err error

	switch idParts[0] {
	case "context":
		err = resourceMountedFileDeleteContext(d, client, toID(idParts[1]), toID(idParts[2]))
	case "stack", "module":
		err = resourceMountedFileDeleteStack(d, client, toID(idParts[1]), toID(idParts[2]))
	default:
		return errors.Errorf("unexpected resource type: %s", idParts[0])
	}

	if err != nil {
		return errors.Wrap(err, "could not delete mounted file")
	}

	d.SetId("")
	return nil
}

func resourceMountedFileDeleteContext(d *schema.ResourceData, client *internal.Client, context graphql.ID, ID graphql.ID) error {
	var mutation struct {
		DeleteContextConfig *structs.ConfigElement `graphql:"contextConfigDelete(context: $context, id: $id)"`
	}

	return client.Mutate(&mutation, map[string]interface{}{"context": context, "id": ID})
}

func resourceMountedFileDeleteStack(d *schema.ResourceData, client *internal.Client, stack graphql.ID, ID graphql.ID) error {
	var mutation struct {
		DeleteStackConfig *structs.ConfigElement `graphql:"stackConfigDelete(stack: $stack, id: $id)"`
	}

	return client.Mutate(&mutation, map[string]interface{}{"stack": stack, "id": ID})
}
