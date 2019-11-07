package spacelift

import (
	"path"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
)

func resourceContextAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceContextAttachmentCreate,
		Read:   resourceContextAttachmentRead,
		Delete: resourceContextAttachmentDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"context_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "ID of the context to attach",
				Required:    true,
				ForceNew:    true,
			},
			"priority": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Priority of the context attachment, used in case of conflicts",
				Optional:    true,
				Default:     0,
				ForceNew:    true,
			},
			"stack_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "ID of the stack to attach the context to",
				Required:    true,
				ForceNew:    true,
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
		"stack":    toID(d.Get("stack_id")),
		"priority": graphql.Int(d.Get("priority").(int)),
	}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not attach context")
	}

	d.SetId(path.Join(contextID, mutation.AttachContext.ID))
	return nil
}

func resourceContextAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 {
		return errors.Errorf("unexpected ID: %s", d.Id())
	}

	var query struct {
		Context *struct {
			Attachment *structs.ContextAttachment `graphql:"attachedStack(id: $id)"`
		} `graphql:"context(id: $context)"`
	}

	variables := map[string]interface{}{
		"context": toID(idParts[0]),
		"id":      toID(idParts[1]),
	}

	if err := meta.(*Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for context attachment")
	}

	if query.Context == nil || query.Context.Attachment == nil {
		d.SetId("")
		return nil
	}

	attachment := query.Context.Attachment
	d.Set("context_id", idParts[0])
	d.Set("priority", attachment.Priority)
	d.Set("stack_id", attachment.StackID)

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

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not detach context")
	}

	d.SetId("")
	return nil
}
