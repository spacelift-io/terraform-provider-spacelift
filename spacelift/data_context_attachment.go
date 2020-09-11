package spacelift

import (
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
)

func dataContextAttachment() *schema.Resource {
	return &schema.Resource{
		Read: dataContextAttachmentRead,

		Schema: map[string]*schema.Schema{
			"attachment_id": {
				Type:        schema.TypeString,
				Description: "ID of the attachment",
				Required:    true,
			},
			"context_id": {
				Type:        schema.TypeString,
				Description: "ID of the attached context",
				Computed:    true,
			},
			"module_id": {
				Type:          schema.TypeString,
				Description:   "ID of the attached module",
				ConflictsWith: []string{"stack_id"},
				Computed:      true,
			},
			"priority": {
				Type:        schema.TypeInt,
				Description: "Priority of the context attachment, used in case of conflicts",
				Computed:    true,
			},
			"stack_id": {
				Type:          schema.TypeString,
				Description:   "ID of the attached stack",
				ConflictsWith: []string{"module_id"},
				Computed:      true,
			},
		},
	}
}

func dataContextAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	attachmentID := d.Get("attachment_id")

	parts := strings.SplitN(attachmentID.(string), "/", 2)
	if len(parts) != 2 {
		return errors.Errorf("unexpected attachment ID: %s", attachmentID)
	}

	d.Set("context_id", parts[0])

	var query struct {
		Context *struct {
			Attachment *structs.ContextAttachment `graphql:"attachedStack(id: $id)"`
		} `graphql:"context(id: $context)"`
	}

	variables := map[string]interface{}{
		"context": toID(parts[0]),
		"id":      toID(parts[1]),
	}

	if err := meta.(*Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for context attachment")
	}

	if query.Context == nil {
		return errors.New("context not found")
	}

	if query.Context.Attachment == nil {
		return errors.New("context attachment not found")
	}

	attachment := query.Context.Attachment
	d.SetId(attachmentID.(string))
	d.Set("priority", attachment.Priority)

	if attachment.IsModule {
		d.Set("module_id", attachment.StackID)
	} else {
		d.Set("stack_id", attachment.StackID)
	}

	return nil
}
