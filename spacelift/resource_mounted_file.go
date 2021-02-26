package spacelift

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourceMountedFile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMountedFileCreate,
		ReadContext:   resourceMountedFileRead,
		DeleteContext: resourceMountedFileDelete,

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
				Type:         schema.TypeString,
				Description:  "ID of the context on which the mounted file is defined",
				ExactlyOneOf: []string{"context_id", "module_id", "stack_id"},
				Optional:     true,
				ForceNew:     true,
			},
			"module_id": {
				Type:        schema.TypeString,
				Description: "ID of the module on which the mounted file is defined",
				Optional:    true,
				ForceNew:    true,
			},
			"relative_path": {
				Type:        schema.TypeString,
				Description: "Relative path to the mounted file, without the /mnt/workspace/ prefix",
				Required:    true,
				ForceNew:    true,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack on which the mounted file is defined",
				Optional:    true,
				ForceNew:    true,
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

func resourceMountedFileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	variables := map[string]interface{}{
		"config": structs.ConfigInput{
			ID:        toID(d.Get("relative_path")),
			Type:      structs.ConfigType("FILE_MOUNT"),
			Value:     toString(d.Get("content")),
			WriteOnly: graphql.Boolean(d.Get("write_only").(bool)),
		},
	}

	if contextID, ok := d.GetOk("context_id"); ok {
		variables["context"] = toID(contextID)
		return resourceMountedFileCreateContext(ctx, d, meta.(*internal.Client), variables)
	}

	if stackID, ok := d.GetOk("stack_id"); ok {
		variables["stack"] = toID(stackID)
		return resourceMountedFileCreateStack(ctx, d, meta.(*internal.Client), variables)
	}

	variables["stack"] = toID(d.Get("module_id"))
	return resourceMountedFileCreateModule(ctx, d, meta.(*internal.Client), variables)
}

func resourceMountedFileCreateContext(ctx context.Context, d *schema.ResourceData, client *internal.Client, variables map[string]interface{}) diag.Diagnostics {
	var mutation struct {
		AddContextConfig structs.ConfigElement `graphql:"contextConfigAdd(context: $context, config: $config)"`
	}

	if err := client.Mutate(ctx, &mutation, variables); err != nil {
		return diag.Errorf("could not create context mounted file: %v", err)
	}

	d.SetId(fmt.Sprintf("context/%s/%s", d.Get("context_id"), d.Get("relative_path")))

	return resourceMountedFileRead(ctx, d, client)
}

func resourceMountedFileCreateModule(ctx context.Context, d *schema.ResourceData, client *internal.Client, variables map[string]interface{}) diag.Diagnostics {
	var mutation struct {
		AddModuleConfig structs.ConfigElement `graphql:"stackConfigAdd(stack: $stack, config: $config)"`
	}

	if err := client.Mutate(ctx, &mutation, variables); err != nil {
		return diag.Errorf("could not module mounted file: %v", err)
	}

	d.SetId(fmt.Sprintf("module/%s/%s", d.Get("module_id"), d.Get("relative_path")))

	return resourceMountedFileRead(ctx, d, client)
}

func resourceMountedFileCreateStack(ctx context.Context, d *schema.ResourceData, client *internal.Client, variables map[string]interface{}) diag.Diagnostics {
	var mutation struct {
		AddStackConfig structs.ConfigElement `graphql:"stackConfigAdd(stack: $stack, config: $config)"`
	}

	if err := client.Mutate(ctx, &mutation, variables); err != nil {
		return diag.Errorf("could not create stack mounted file: %v", err)
	}

	d.SetId(fmt.Sprintf("stack/%s/%s", d.Get("stack_id"), d.Get("relative_path")))

	return resourceMountedFileRead(ctx, d, client)
}

func resourceMountedFileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	idParts := strings.SplitN(d.Id(), "/", 3)
	if len(idParts) != 3 {
		return diag.Errorf("unexpected resource ID: %s", d.Id())
	}

	client := meta.(*internal.Client)
	var element *structs.ConfigElement
	var err error

	switch resourceType, resourceID, filePath := idParts[0], idParts[1], idParts[2]; resourceType {
	case "context":
		element, err = resourceMountedFileReadContext(ctx, d, client, resourceID, filePath)
	case "module":
		element, err = resourceMountedFileReadModule(ctx, d, client, resourceID, filePath)
	case "stack":
		element, err = resourceMountedFileReadStack(ctx, d, client, resourceID, filePath)
	default:
		return diag.Errorf("unexpected resource type: %s", idParts[0])
	}

	if err != nil {
		return diag.FromErr(err)
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

func resourceMountedFileReadContext(ctx context.Context, d *schema.ResourceData, client *internal.Client, context, ID string) (*structs.ConfigElement, error) {
	var query struct {
		Context *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"context(id: $context)"`
	}

	if err := client.Query(ctx, &query, map[string]interface{}{"context": toID(context), "id": toID(ID)}); err != nil {
		return nil, errors.Wrap(err, "could not query for context mounted file")
	}

	if query.Context == nil {
		return nil, nil
	}

	return query.Context.ConfigElement, nil
}

func resourceMountedFileReadModule(ctx context.Context, d *schema.ResourceData, client *internal.Client, module, ID string) (*structs.ConfigElement, error) {
	var query struct {
		Module *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"module(id: $module)"`
	}

	if err := client.Query(ctx, &query, map[string]interface{}{"module": toID(module), "id": toID(ID)}); err != nil {
		return nil, errors.Wrap(err, "could not query for module mounted file")
	}

	if query.Module == nil {
		return nil, nil
	}

	return query.Module.ConfigElement, nil
}

func resourceMountedFileReadStack(ctx context.Context, d *schema.ResourceData, client *internal.Client, stack, ID string) (*structs.ConfigElement, error) {
	var query struct {
		Stack *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"stack(id: $stack)"`
	}

	if err := client.Query(ctx, &query, map[string]interface{}{"stack": toID(stack), "id": toID(ID)}); err != nil {
		return nil, errors.Wrap(err, "could not query for stack mounted file")
	}

	if query.Stack == nil {
		return nil, nil
	}

	return query.Stack.ConfigElement, nil
}

func resourceMountedFileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	idParts := strings.SplitN(d.Id(), "/", 3)
	if len(idParts) != 3 {
		return diag.Errorf("unexpected resource ID: %s", d.Id())
	}

	client := meta.(*internal.Client)
	var err error

	switch resourceType, contextID, fileID := idParts[0], idParts[1], idParts[2]; resourceType {
	case "context":
		err = resourceMountedFileDeleteContext(ctx, d, client, toID(contextID), toID(fileID))
	case "stack", "module":
		err = resourceMountedFileDeleteStack(ctx, d, client, toID(contextID), toID(fileID))
	default:
		return diag.Errorf("unexpected resource type: %s", resourceType)
	}

	if err != nil {
		return diag.Errorf("could not delete mounted file: %v", err)
	}

	d.SetId("")

	return nil
}

func resourceMountedFileDeleteContext(ctx context.Context, d *schema.ResourceData, client *internal.Client, context graphql.ID, ID graphql.ID) error {
	var mutation struct {
		DeleteContextConfig *structs.ConfigElement `graphql:"contextConfigDelete(context: $context, id: $id)"`
	}

	return client.Mutate(ctx, &mutation, map[string]interface{}{"context": context, "id": ID})
}

func resourceMountedFileDeleteStack(ctx context.Context, d *schema.ResourceData, client *internal.Client, stack graphql.ID, ID graphql.ID) error {
	var mutation struct {
		DeleteStackConfig *structs.ConfigElement `graphql:"stackConfigDelete(stack: $stack, id: $id)"`
	}

	return client.Mutate(ctx, &mutation, map[string]interface{}{"stack": stack, "id": ID})
}
