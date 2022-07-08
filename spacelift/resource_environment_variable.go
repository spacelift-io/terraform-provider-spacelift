package spacelift

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourceEnvironmentVariable() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_environment_variable` defines an environment variable on " +
			"the context (`spacelift_context`), stack (`spacelift_stack`) or a " +
			"module (`spacelift_module`), thereby allowing to pass and share " +
			"various secrets and configuration with Spacelift stacks.",

		CreateContext: resourceEnvironmentVariableCreate,
		ReadContext:   resourceEnvironmentVariableRead,
		DeleteContext: resourceEnvironmentVariableDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"checksum": {
				Type:        schema.TypeString,
				Description: "SHA-256 checksum of the value",
				Computed:    true,
			},
			"context_id": {
				Type:         schema.TypeString,
				Description:  "ID of the context on which the environment variable is defined",
				Optional:     true,
				ExactlyOneOf: []string{"context_id", "stack_id", "module_id"},
				ForceNew:     true,
			},
			"module_id": {
				Type:        schema.TypeString,
				Description: "ID of the module on which the environment variable is defined",
				Optional:    true,
				ForceNew:    true,
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

func resourceEnvironmentVariableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	variables := map[string]interface{}{
		"config": structs.ConfigInput{
			ID:        toID(d.Get("name")),
			Type:      structs.ConfigType("ENVIRONMENT_VARIABLE"),
			Value:     toString(d.Get("value")),
			WriteOnly: graphql.Boolean(d.Get("write_only").(bool)),
		},
	}

	if contextID, ok := d.GetOk("context_id"); ok {
		variables["context"] = toID(contextID)
		return resourceEnvironmentVariableCreateContext(ctx, d, meta.(*internal.Client), variables)
	}

	if stackID, ok := d.GetOk("stack_id"); ok {
		if err := verifyStack(ctx, stackID.(string), meta); err != nil {
			return diag.FromErr(err)
		}

		variables["stack"] = toID(stackID)
		return resourceEnvironmentVariableCreateStack(ctx, d, meta.(*internal.Client), variables)
	}

	moduleID := d.Get("module_id").(string)
	if err := verifyModule(ctx, moduleID, meta); err != nil {
		return diag.FromErr(err)
	}

	variables["stack"] = toID(moduleID)

	return resourceEnvironmentVariableCreateModule(ctx, d, meta.(*internal.Client), variables)
}

func resourceEnvironmentVariableCreateContext(ctx context.Context, d *schema.ResourceData, client *internal.Client, variables map[string]interface{}) diag.Diagnostics {
	var mutation struct {
		AddContextConfig structs.ConfigElement `graphql:"contextConfigAdd(context: $context, config: $config)"`
	}

	if err := client.Mutate(ctx, "EnvironmentVariableCreateContext", &mutation, variables); err != nil {
		return diag.Errorf("could not create context environment variable: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(fmt.Sprintf("context/%s/%s", d.Get("context_id"), d.Get("name")))

	return resourceEnvironmentVariableRead(ctx, d, client)
}

func resourceEnvironmentVariableCreateModule(ctx context.Context, d *schema.ResourceData, client *internal.Client, variables map[string]interface{}) diag.Diagnostics {
	var mutation struct {
		AddModuleConfig structs.ConfigElement `graphql:"stackConfigAdd(stack: $stack, config: $config)"`
	}

	if err := client.Mutate(ctx, "EnvironmentVariableCreateModule", &mutation, variables); err != nil {
		return diag.Errorf("could not create module environment variable: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(fmt.Sprintf("module/%s/%s", d.Get("module_id"), d.Get("name")))

	return resourceEnvironmentVariableRead(ctx, d, client)
}

func resourceEnvironmentVariableCreateStack(ctx context.Context, d *schema.ResourceData, client *internal.Client, variables map[string]interface{}) diag.Diagnostics {
	var mutation struct {
		AddStackConfig structs.ConfigElement `graphql:"stackConfigAdd(stack: $stack, config: $config)"`
	}

	if err := client.Mutate(ctx, "EnvironmentVariableCreateStack", &mutation, variables); err != nil {
		return diag.Errorf("could not create stack environment variable: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(fmt.Sprintf("stack/%s/%s", d.Get("stack_id"), d.Get("name")))

	return resourceEnvironmentVariableRead(ctx, d, client)
}

func resourceEnvironmentVariableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	idParts := strings.SplitN(d.Id(), "/", 3)
	if len(idParts) != 3 {
		return diag.Errorf("unexpected resource ID: %s", d.Id())
	}

	client := meta.(*internal.Client)
	var element *structs.ConfigElement
	var err error

	resourceType, resourceID, variableName := idParts[0], idParts[1], idParts[2]

	switch resourceType {
	case "context":
		element, err = resourceEnvironmentVariableReadContext(ctx, d, client, resourceID, variableName)
	case "module":
		element, err = resourceEnvironmentVariableReadModule(ctx, d, client, resourceID, variableName)
	case "stack":
		element, err = resourceEnvironmentVariableReadStack(ctx, d, client, resourceID, variableName)
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
	d.Set("name", variableName)
	d.Set("write_only", element.WriteOnly)

	if value := element.Value; value != nil {
		d.Set("value", *value)
	} else {
		d.Set("value", element.Checksum)
	}

	return nil
}

func resourceEnvironmentVariableReadContext(ctx context.Context, d *schema.ResourceData, client *internal.Client, contextID, variableName string) (*structs.ConfigElement, error) {
	var query struct {
		Context *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"context(id: $context)"`
	}

	if err := client.Query(ctx, "EnvironmentVariableReadContext", &query, map[string]interface{}{"context": toID(contextID), "id": toID(variableName)}); err != nil {
		return nil, errors.Wrap(err, "could not query for context environment variable")
	}

	if query.Context == nil {
		return nil, nil
	}

	d.Set("context_id", contextID)

	return query.Context.ConfigElement, nil
}

func resourceEnvironmentVariableReadModule(ctx context.Context, d *schema.ResourceData, client *internal.Client, moduleID, variableName string) (*structs.ConfigElement, error) {
	var query struct {
		Module *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"module(id: $module)"`
	}

	if err := client.Query(ctx, "EnvironmentVariableReadModule", &query, map[string]interface{}{"module": toID(moduleID), "id": toID(variableName)}); err != nil {
		return nil, errors.Wrap(err, "could not query for module environment variable")
	}

	if query.Module == nil {
		return nil, nil
	}

	d.Set("module_id", moduleID)

	return query.Module.ConfigElement, nil
}

func resourceEnvironmentVariableReadStack(ctx context.Context, d *schema.ResourceData, client *internal.Client, stackID, variableName string) (*structs.ConfigElement, error) {
	var query struct {
		Stack *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"stack(id: $stack)"`
	}

	if err := client.Query(ctx, "EnvironmentVariableReadStack", &query, map[string]interface{}{"stack": toID(stackID), "id": toID(variableName)}); err != nil {
		return nil, errors.Wrap(err, "could not query for stack environment variable")
	}

	if query.Stack == nil {
		return nil, nil
	}

	d.Set("stack_id", stackID)

	return query.Stack.ConfigElement, nil
}

func resourceEnvironmentVariableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	idParts := strings.SplitN(d.Id(), "/", 3)
	if len(idParts) != 3 {
		return diag.Errorf("unexpected resource ID: %s", d.Id())
	}

	client := meta.(*internal.Client)
	var err error

	switch resourceType, resourceID, variableName := idParts[0], idParts[1], idParts[2]; resourceType {
	case "context":
		err = resourceEnvironmentVariableDeleteContext(ctx, d, client, resourceID, variableName)
	case "module", "stack":
		err = resourceEnvironmentVariableDeleteStack(ctx, d, client, resourceID, variableName)
	default:
		return diag.Errorf("unexpected resource type: %s", idParts[0])
	}

	if err != nil {
		return diag.Errorf("could not delete environment variable: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")
	return nil
}

func resourceEnvironmentVariableDeleteContext(ctx context.Context, d *schema.ResourceData, client *internal.Client, context, id string) error {
	var mutation struct {
		DeleteContextConfig *structs.ConfigElement `graphql:"contextConfigDelete(context: $context, id: $id)"`
	}

	return client.Mutate(ctx, "EnvironmentVariableDeleteContext", &mutation, map[string]interface{}{"context": toID(context), "id": toID(id)})
}

func resourceEnvironmentVariableDeleteStack(ctx context.Context, d *schema.ResourceData, client *internal.Client, stack, id string) error {
	var mutation struct {
		DeleteStackConfig *structs.ConfigElement `graphql:"stackConfigDelete(stack: $stack, id: $id)"`
	}

	return client.Mutate(ctx, "EnvironmentVariableDeleteStack", &mutation, map[string]interface{}{"stack": toID(stack), "id": toID(id)})
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
