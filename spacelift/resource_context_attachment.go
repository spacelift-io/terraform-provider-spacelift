package spacelift

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourceContextAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceContextAttachmentCreate,
		Read:   resourceContextAttachmentRead,
		Delete: resourceContextAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceContextAttachmentImport,
		},

		Schema: map[string]*schema.Schema{
			"context_id": {
				Type:        schema.TypeString,
				Description: "ID of the context to attach",
				Required:    true,
				ForceNew:    true,
			},
			"module_id": {
				Type:          schema.TypeString,
				Description:   "ID of the module to attach the context to",
				ConflictsWith: []string{"stack_id"},
				Optional:      true,
				ForceNew:      true,
			},
			"priority": {
				Type:        schema.TypeInt,
				Description: "Priority of the context attachment, used in case of conflicts",
				Optional:    true,
				Default:     0,
				ForceNew:    true,
			},
			"stack_id": {
				Type:          schema.TypeString,
				Description:   "ID of the stack to attach the context to",
				ConflictsWith: []string{"module_id"},
				Optional:      true,
				ForceNew:      true,
			},
		},
	}
}

func resourceContextAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		AttachContext structs.ContextAttachment `graphql:"contextAttach(id: $id, stack: $stack, priority: $priority)"`
	}

	contextID := d.Get("context_id").(string)

	variables := map[string]interface{}{
		"id":       toID(contextID),
		"priority": graphql.Int(d.Get("priority").(int)),
	}

	if stackID, ok := d.GetOk("stack_id"); ok {
		variables["stack"] = toID(stackID)
	} else if moduleID, ok := d.GetOk("module_id"); ok {
		variables["stack"] = toID(moduleID)
	} else {
		return errors.New("either module_id or stack_id must be provided")
	}

	if err := meta.(*internal.Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not attach context")
	}

	d.SetId(path.Join(contextID, mutation.AttachContext.ID))
	return nil
}

func resourceContextAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	contextID := d.Get("context_id").(string)
	var projectID string

	if stackID, ok := d.GetOk("stack_id"); ok {
		projectID = stackID.(string)
	} else if moduleID, ok := d.GetOk("module_id"); ok {
		projectID = moduleID.(string)
	} else {
		return errors.New("either module_id or stack_id must be provided")
	}

	if attachment, err := resourceContextAttachmentFetch(contextID, projectID, meta); err != nil {
		return err
	} else if attachment == nil {
		d.SetId("")
	} else {
		d.Set("priority", attachment.Priority)
	}

	return nil
}

func resourceContextAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 {
		return errors.Errorf("unexpected ID: %s", d.Id())
	}

	var mutation struct {
		DetachContext *structs.ContextAttachment `graphql:"contextDetach(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(idParts[1])}

	if err := meta.(*internal.Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not detach context")
	}

	d.SetId("")
	return nil
}

func resourceContextAttachmentImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	input := d.Id()

	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("expecting attachment ID as $contextId/$projectId")
	}

	contextID, projectID := parts[0], parts[1]

	attachment, err := resourceContextAttachmentFetch(contextID, projectID, meta)
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

func resourceContextAttachmentFetch(contextID, projectID string, meta interface{}) (*structs.ContextAttachment, error) {
	var query struct {
		Context *struct {
			Attachment *structs.ContextAttachment `graphql:"attachedStack(id: $project)"`
		} `graphql:"context(id: $context)"`
	}

	variables := map[string]interface{}{
		"context": contextID,
		"project": toID(projectID),
	}

	if err := meta.(*internal.Client).Query(&query, variables); err != nil {
		return nil, errors.Wrap(err, "could not query for context attachment")
	}

	if query.Context == nil {
		return nil, nil
	}

	return query.Context.Attachment, nil
}
