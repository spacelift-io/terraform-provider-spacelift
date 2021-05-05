package spacelift

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataMountedFile() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataMountedFileRead,

		Schema: map[string]*schema.Schema{
			"checksum": {
				Type:        schema.TypeString,
				Description: "SHA-256 checksum of the value",
				Computed:    true,
			},
			"content": {
				Type:        schema.TypeString,
				Description: "content of the mounted file encoded using Base-64",
				Sensitive:   true,
				Computed:    true,
			},
			"context_id": {
				Type:         schema.TypeString,
				Description:  "ID of the context where the mounted file is stored",
				ExactlyOneOf: []string{"context_id", "stack_id", "module_id"},
				Optional:     true,
			},
			"module_id": {
				Type:         schema.TypeString,
				Description:  "ID of the module where the mounted file is stored",
				ExactlyOneOf: []string{"context_id", "stack_id", "module_id"},
				Optional:     true,
			},
			"relative_path": {
				Type:        schema.TypeString,
				Description: "relative path to the mounted file",
				Required:    true,
			},
			"stack_id": {
				Type:         schema.TypeString,
				Description:  "ID of the stack where the mounted file is stored",
				ExactlyOneOf: []string{"context_id", "stack_id", "module_id"},
				Optional:     true,
			},
			"write_only": {
				Type:        schema.TypeBool,
				Description: "indicates whether the value can be read back outside a Run",
				Computed:    true,
			},
		},
	}
}

func dataMountedFileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if _, ok := d.GetOk("context_id"); ok {
		return dataMountedFileReadContext(ctx, d, meta)
	}

	if _, ok := d.GetOk("module_id"); ok {
		return dataMountedFileReadModule(ctx, d, meta)
	}

	return dataMountedFileReadStack(ctx, d, meta)
}

func dataMountedFileReadContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Context *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"context(id: $context)"`
	}

	contextID := d.Get("context_id")
	variableName := d.Get("relative_path")

	variables := map[string]interface{}{
		"context": toID(contextID),
		"id":      toID(variableName),
	}

	if err := meta.(*internal.Client).Query(ctx, "MountedFileReadContext", &query, variables); err != nil {
		return diag.Errorf("could not query for context mounted file: %v", err)
	}

	if query.Context == nil {
		return diag.Errorf("context not found")
	}

	configElement := query.Context.ConfigElement
	if configElement == nil {
		return diag.Errorf("mounted file not found")
	}

	if configElement.Type != "FILE_MOUNT" {
		return diag.Errorf("config element is not a mounted file")
	}

	d.SetId(fmt.Sprintf("context/%s/%s", contextID, variableName))

	populateMountedFile(d, query.Context.ConfigElement)

	return nil
}

func dataMountedFileReadModule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Module *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"module(id: $module)"`
	}

	moduleID := d.Get("module_id")
	variableName := d.Get("relative_path")

	variables := map[string]interface{}{
		"module": toID(moduleID),
		"id":     toID(variableName),
	}

	if err := meta.(*internal.Client).Query(ctx, "MountedFileReadModule", &query, variables); err != nil {
		return diag.Errorf("could not query for module mounted file: %v", err)
	}

	if query.Module == nil {
		return diag.Errorf("module not found")
	}

	if query.Module.ConfigElement == nil {
		return diag.Errorf("mounted file not found")
	}

	d.SetId(fmt.Sprintf("module/%s/%s", moduleID, variableName))

	populateMountedFile(d, query.Module.ConfigElement)

	return nil
}

func dataMountedFileReadStack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Stack *struct {
			ConfigElement *structs.ConfigElement `graphql:"configElement(id: $id)"`
		} `graphql:"stack(id: $stack)"`
	}

	stackID := d.Get("stack_id")
	variableName := d.Get("relative_path")

	variables := map[string]interface{}{
		"stack": toID(stackID),
		"id":    toID(variableName),
	}

	if err := meta.(*internal.Client).Query(ctx, "MountedFileReadStack", &query, variables); err != nil {
		return diag.Errorf("could not query for stack mounted file: %v", err)
	}

	if query.Stack == nil {
		return diag.Errorf("stack not found")
	}

	if query.Stack.ConfigElement == nil {
		return diag.Errorf("mounted file not found")
	}

	d.SetId(fmt.Sprintf("stack/%s/%s", stackID, variableName))

	populateMountedFile(d, query.Stack.ConfigElement)

	return nil
}

func populateMountedFile(d *schema.ResourceData, el *structs.ConfigElement) {
	d.Set("checksum", el.Checksum)
	d.Set("write_only", el.WriteOnly)

	if el.Value != nil {
		d.Set("content", *el.Value)
	} else {
		d.Set("content", nil)
	}
}
