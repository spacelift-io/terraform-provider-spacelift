package spacelift

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceMountedFile() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_mounted_file` represents a file mounted in each Run's " +
			"workspace that is part of a configuration of a context (`spacelift_context`), " +
			"stack (`spacelift_stack`) or a module (`spacelift_module`). In principle, " +
			"it's very similar to an environment variable (`spacelift_environment_variable`) " +
			"except that the value is written to the filesystem rather than passed to " +
			"the environment.",

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
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
				ConflictsWith:    []string{"content_wo", "content_wo_version"},
				AtLeastOneOf:     []string{"content", "content_wo"},
			},
			"content_wo": {
				Type:          schema.TypeString,
				Description:   "Content of the mounted file encoded using Base-64. The content is not stored in the state. Modify content_wo_version to trigger an update. This field requires Terraform/OpenTofu 1.11+.",
				Sensitive:     true,
				Optional:      true,
				WriteOnly:     true,
				ConflictsWith: []string{"content"},
				RequiredWith:  []string{"content_wo_version"},
				AtLeastOneOf:  []string{"content", "content_wo"},
			},
			"content_wo_version": {
				Type:          schema.TypeString,
				Description:   "Used together with content_wo to trigger an update to the content. Increment this value when an update to content_wo is required. This field requires Terraform/OpenTofu 1.11+.",
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"content"},
				RequiredWith:  []string{"content_wo"},
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
			"description": {
				Type:        schema.TypeString,
				Description: "Free-form description of the mounted file",
				Optional:    true,
				ForceNew:    true,
			},
			"write_only": {
				Type:        schema.TypeBool,
				Description: "Indicates whether the content can be read back outside a Run. Defaults to `true`.",
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
		},
	}
}

func resourceMountedFileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var content string
	if v, ok := d.GetOk("content"); ok {
		content = v.(string)
	}

	if _, ok := d.GetOk("content_wo_version"); ok {
		p := cty.GetAttrPath("content_wo")
		woVal, d := d.GetRawConfigAt(p)
		if d.HasError() {
			return diag.FromErr(fmt.Errorf("could not get write-only content: %v", d))
		}

		if !woVal.IsNull() {
			content = woVal.AsString()
		}
	}

	variables := map[string]interface{}{
		"config": structs.ConfigInput{
			ID:          toID(d.Get("relative_path")),
			Type:        structs.ConfigType("FILE_MOUNT"),
			Value:       toString(content),
			WriteOnly:   graphql.Boolean(d.Get("write_only").(bool)),
			Description: toOptionalString(d.Get("description")),
		},
	}

	if contextID, ok := d.GetOk("context_id"); ok {
		variables["context"] = toID(contextID)
		return resourceMountedFileCreateContext(ctx, d, meta.(*internal.Client), variables)
	}

	if stackID, ok := d.GetOk("stack_id"); ok {
		if err := verifyStack(ctx, stackID.(string), meta); err != nil {
			return diag.FromErr(err)
		}

		variables["stack"] = toID(stackID)
		return resourceMountedFileCreateStack(ctx, d, meta.(*internal.Client), variables)
	}

	moduleID := d.Get("module_id").(string)
	if err := verifyModule(ctx, moduleID, meta); err != nil {
		return diag.FromErr(err)
	}

	variables["stack"] = toID(moduleID)
	return resourceMountedFileCreateModule(ctx, d, meta.(*internal.Client), variables)
}

func resourceMountedFileCreateContext(ctx context.Context, d *schema.ResourceData, client *internal.Client, variables map[string]interface{}) diag.Diagnostics {
	var mutation struct {
		AddContextConfig structs.ConfigElement `graphql:"contextConfigAdd(context: $context, config: $config)"`
	}

	if err := client.Mutate(ctx, "MountedFileCreateContext", &mutation, variables); err != nil {
		return diag.Errorf("could not create context mounted file: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(fmt.Sprintf("context/%s/%s", d.Get("context_id"), d.Get("relative_path")))

	return resourceMountedFileRead(ctx, d, client)
}

func resourceMountedFileCreateModule(ctx context.Context, d *schema.ResourceData, client *internal.Client, variables map[string]interface{}) diag.Diagnostics {
	var mutation struct {
		AddModuleConfig structs.ConfigElement `graphql:"stackConfigAdd(stack: $stack, config: $config)"`
	}

	if err := client.Mutate(ctx, "MountedFileCreateModule", &mutation, variables); err != nil {
		return diag.Errorf("could not create module mounted file: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(fmt.Sprintf("module/%s/%s", d.Get("module_id"), d.Get("relative_path")))

	return resourceMountedFileRead(ctx, d, client)
}

func resourceMountedFileCreateStack(ctx context.Context, d *schema.ResourceData, client *internal.Client, variables map[string]interface{}) diag.Diagnostics {
	var mutation struct {
		AddStackConfig structs.ConfigElement `graphql:"stackConfigAdd(stack: $stack, config: $config)"`
	}

	if err := client.Mutate(ctx, "MountedFileCreateStack", &mutation, variables); err != nil {
		return diag.Errorf("could not create stack mounted file: %v", internal.FromSpaceliftError(err))
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

	resourceType, resourceID, relativePath := idParts[0], idParts[1], idParts[2]

	switch resourceType {
	case "context":
		element, err = resourceMountedFileReadContext(ctx, d, client, resourceID, relativePath)
	case "module":
		element, err = resourceMountedFileReadModule(ctx, d, client, resourceID, relativePath)
	case "stack":
		element, err = resourceMountedFileReadStack(ctx, d, client, resourceID, relativePath)
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
	d.Set("relative_path", relativePath)
	d.Set("write_only", element.WriteOnly)

	if element.Description != nil {
		d.Set("description", *element.Description)
	}

	if _, hasValueWo := d.GetOk("content_wo_version"); !hasValueWo {
		if value := element.Value; value != nil {
			d.Set("content", *value)
		} else {
			d.Set("content", element.Checksum)
		}
	}

	return nil
}

func resourceMountedFileReadContext(ctx context.Context, d *schema.ResourceData, client *internal.Client, contextID, relativePath string) (*structs.ConfigElement, error) {
	var query struct {
		Context *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"context(id: $context)"`
	}

	variables := map[string]interface{}{"context": toID(contextID), "id": toID(relativePath)}

	if err := client.Query(ctx, "MountedFileReadContext", &query, variables); err != nil {
		return nil, errors.Wrap(err, "could not query for context mounted file")
	}

	if query.Context == nil {
		return nil, nil
	}

	d.Set("context_id", contextID)

	return query.Context.ConfigElement, nil
}

func resourceMountedFileReadModule(ctx context.Context, d *schema.ResourceData, client *internal.Client, moduleID, relativePath string) (*structs.ConfigElement, error) {
	var query struct {
		Module *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"module(id: $module)"`
	}

	variables := map[string]interface{}{"module": toID(moduleID), "id": toID(relativePath)}

	if err := client.Query(ctx, "MountedFileReadModule", &query, variables); err != nil {
		return nil, errors.Wrap(err, "could not query for module mounted file")
	}

	if query.Module == nil {
		return nil, nil
	}

	d.Set("module_id", moduleID)

	return query.Module.ConfigElement, nil
}

func resourceMountedFileReadStack(ctx context.Context, d *schema.ResourceData, client *internal.Client, stackID, relativePath string) (*structs.ConfigElement, error) {
	var query struct {
		Stack *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"stack(id: $stack)"`
	}

	variables := map[string]interface{}{"stack": toID(stackID), "id": toID(relativePath)}

	if err := client.Query(ctx, "MountedFileReadStack", &query, variables); err != nil {
		return nil, errors.Wrap(err, "could not query for stack mounted file")
	}

	if query.Stack == nil {
		return nil, nil
	}

	d.Set("stack_id", stackID)

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
		return diag.Errorf("could not delete mounted file: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func resourceMountedFileDeleteContext(ctx context.Context, d *schema.ResourceData, client *internal.Client, context graphql.ID, id graphql.ID) error {
	var mutation struct {
		DeleteContextConfig *structs.ConfigElement `graphql:"contextConfigDelete(context: $context, id: $id)"`
	}

	return client.Mutate(ctx, "MountedFileDeleteContext", &mutation, map[string]interface{}{"context": context, "id": id})
}

func resourceMountedFileDeleteStack(ctx context.Context, d *schema.ResourceData, client *internal.Client, stack graphql.ID, id graphql.ID) error {
	var mutation struct {
		DeleteStackConfig *structs.ConfigElement `graphql:"stackConfigDelete(stack: $stack, id: $id)"`
	}

	return client.Mutate(ctx, "MountedFileDeleteStack", &mutation, map[string]interface{}{"stack": stack, "id": id})
}
