package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataContextAttachment() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataContextAttachmentRead,

		Schema: map[string]*schema.Schema{
			"context_id": {
				Type:        schema.TypeString,
				Description: "ID of the attached context",
				Required:    true,
			},
			"module_id": {
				Type:         schema.TypeString,
				Description:  "ID of the attached module",
				ExactlyOneOf: []string{"module_id", "stack_id"},
				Optional:     true,
			},
			"priority": {
				Type:        schema.TypeInt,
				Description: "priority of the context attachment, used in case of conflicts",
				Computed:    true,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the attached stack",
				Optional:    true,
			},
		},
	}
}

func dataContextAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	variables := map[string]interface{}{
		"context": toID(d.Get("context_id").(string)),
	}

	if moduleID, ok := d.GetOk("module_id"); ok {
		variables["id"] = toID(moduleID)
	} else {
		variables["id"] = toID(d.Get("stack_id"))
	}

	var query struct {
		Context *struct {
			Attachment *structs.ContextAttachment `graphql:"attachedStack(id: $id)"`
		} `graphql:"context(id: $context)"`
	}

	if err := meta.(*internal.Client).Query(ctx, &query, variables); err != nil {
		d.SetId("")
		return diag.Errorf("could not query for context attachment: %v", err)
	}

	if query.Context == nil {
		d.SetId("")
		return diag.Errorf("context not found")
	}

	if query.Context.Attachment == nil {
		d.SetId("")
		return diag.Errorf("context attachment not found")
	}

	attachment := query.Context.Attachment

	d.SetId(attachment.ID)
	d.Set("priority", attachment.Priority)

	return nil
}
