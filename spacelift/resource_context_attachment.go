package spacelift

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceContextAttachment() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_context_attachment` represents a Spacelift attachment of a " +
			"single context to a single stack or module, with a predefined priority.",

		CreateContext: resourceContextAttachmentCreate,
		ReadContext:   resourceContextAttachmentRead,
		DeleteContext: resourceContextAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceContextAttachmentImport,
		},

		Schema: map[string]*schema.Schema{
			"context_id": {
				Type:             schema.TypeString,
				Description:      "ID of the context to attach",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"module_id": {
				Type:         schema.TypeString,
				Description:  "ID of the module to attach the context to",
				ExactlyOneOf: []string{"module_id", "stack_id"},
				Optional:     true,
				ForceNew:     true,
			},
			"priority": {
				Type:        schema.TypeInt,
				Description: "Priority of the context attachment. All the contexts attached to a stack are sorted by priority (lowest first), though values don't need to be unique. This ordering establishes precedence rules between contexts should there be a conflict and multiple contexts define the same value. Defaults to `0`.",
				Optional:    true,
				Default:     0,
				ForceNew:    true,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack to attach the context to",
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceContextAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		AttachContext structs.ContextAttachment `graphql:"contextAttach(id: $id, stack: $stack, priority: $priority)"`
	}

	contextID := d.Get("context_id").(string)

	variables := map[string]interface{}{
		"id":       toID(contextID),
		"priority": graphql.Int(d.Get("priority").(int)), //nolint:gosec // safe: value known to fit in int32
	}

	if stackID, ok := d.GetOk("stack_id"); ok {
		if err := verifyStack(ctx, stackID.(string), meta); err != nil {
			return diag.FromErr(err)
		}

		variables["stack"] = toID(stackID)
	} else {
		if err := verifyModule(ctx, d.Get("module_id").(string), meta); err != nil {
			return diag.FromErr(err)
		}

		variables["stack"] = toID(d.Get("module_id"))
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ContextAttachmentCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not attach context: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(path.Join(contextID, mutation.AttachContext.ID))

	return nil
}

func resourceContextAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	contextID := d.Get("context_id").(string)
	var projectID string

	if stackID, ok := d.GetOk("stack_id"); ok {
		projectID = stackID.(string)
	} else {
		projectID = d.Get("module_id").(string)
	}

	if attachment, err := resourceContextAttachmentFetch(ctx, contextID, projectID, meta); err != nil {
		return diag.FromErr(err)
	} else if attachment == nil {
		d.SetId("")
	} else {
		d.Set("priority", attachment.Priority)
	}

	return nil
}

func resourceContextAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 {
		return diag.Errorf("unexpected ID: %s", d.Id())
	}

	var mutation struct {
		DetachContext *structs.ContextAttachment `graphql:"contextDetach(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(idParts[1])}

	if err := meta.(*internal.Client).Mutate(ctx, "ContextAttachmentDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not detach context: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func resourceContextAttachmentImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	input := d.Id()

	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("expecting attachment ID as $contextId/$projectId")
	}

	contextID, projectID := parts[0], parts[1]

	attachment, err := resourceContextAttachmentFetch(ctx, contextID, projectID, meta)
	if err != nil {
		return nil, err
	} else if attachment == nil {
		return nil, errors.New("attachment not found")
	}

	if attachment.IsModule {
		d.Set("module_id", projectID)
	} else {
		d.Set("stack_id", projectID)
	}

	d.SetId(path.Join(contextID, attachment.ID))
	d.Set("context_id", contextID)

	return []*schema.ResourceData{d}, nil
}

func resourceContextAttachmentFetch(ctx context.Context, contextID, projectID string, meta interface{}) (*structs.ContextAttachment, error) {
	var query struct {
		Context *struct {
			Attachment *structs.ContextAttachment `graphql:"attachedStack(id: $project)"`
		} `graphql:"context(id: $context)"`
	}

	variables := map[string]interface{}{
		"context": contextID,
		"project": toID(projectID),
	}

	if err := meta.(*internal.Client).Query(ctx, "ContextAttachmentFetch", &query, variables); err != nil {
		return nil, errors.Wrap(err, "could not query for context attachment")
	}

	if query.Context == nil {
		return nil, nil
	}

	return query.Context.Attachment, nil
}
