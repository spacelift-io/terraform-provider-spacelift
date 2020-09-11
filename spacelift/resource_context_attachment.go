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

	if attachment.IsModule {
		d.Set("module_id", attachment.StackID)
	} else {
		d.Set("stack_id", attachment.StackID)
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

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not detach context")
	}

	d.SetId("")
	return nil
}
