package spacelift

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataContextAttachment() *schema.Resource {
	return &schema.Resource{
		Read: dataContextAttachmentRead,

		Schema: map[string]*schema.Schema{
			"context_id": {
				Type:        schema.TypeString,
				Description: "ID of the attached context",
				Required:    true,
			},
			"module_id": {
				Type:          schema.TypeString,
				Description:   "ID of the attached module",
				ConflictsWith: []string{"stack_id"},
				Optional:      true,
			},
			"priority": {
				Type:        schema.TypeInt,
				Description: "priority of the context attachment, used in case of conflicts",
				Computed:    true,
			},
			"stack_id": {
				Type:          schema.TypeString,
				Description:   "ID of the attached stack",
				ConflictsWith: []string{"module_id"},
				Optional:      true,
			},
		},
	}
}

func dataContextAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	variables := map[string]interface{}{
		"context": toID(d.Get("context_id").(string)),
	}

	if ID, ok := d.GetOk("module_id"); ok {
		variables["id"] = toID(ID.(string))
	} else if ID, ok := d.GetOk("stack_id"); ok {
		variables["id"] = toID(ID.(string))
	} else {
		return errors.Errorf("either module_id or stack_id must be present")
	}

	var query struct {
		Context *struct {
			Attachment *structs.ContextAttachment `graphql:"attachedStack(id: $id)"`
		} `graphql:"context(id: $context)"`
	}

	if err := meta.(*internal.Client).Query(&query, variables); err != nil {
		d.SetId("")
		return errors.Wrap(err, "could not query for context attachment")
	}

	if query.Context == nil {
		d.SetId("")
		return errors.New("context not found")
	}

	if query.Context.Attachment == nil {
		d.SetId("")
		return errors.New("context attachment not found")
	}

	attachment := query.Context.Attachment

	d.SetId(attachment.ID)
	d.Set("priority", attachment.Priority)

	return nil
}
