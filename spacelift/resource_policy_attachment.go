package spacelift

import (
	"path"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourcePolicyAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourcePolicyAttachmentCreate,
		Read:   resourcePolicyAttachmentRead,
		Delete: resourcePolicyAttachmentDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"policy_id": {
				Type:        schema.TypeString,
				Description: "ID of the policy to attach",
				Required:    true,
				ForceNew:    true,
			},
			"module_id": {
				Type:          schema.TypeString,
				Description:   "ID of the module to attach the policy to",
				ConflictsWith: []string{"stack_id"},
				Optional:      true,
				ForceNew:      true,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack to attach the policy to",
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourcePolicyAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		AttachPolicy structs.PolicyAttachment `graphql:"policyAttach(id: $id, stack: $stack)"`
	}

	policyID := d.Get("policy_id").(string)

	variables := map[string]interface{}{
		"id": toID(policyID),
	}

	if stackID, ok := d.GetOk("stack_id"); ok {
		variables["stack"] = toID(stackID)
	} else if moduleID, ok := d.GetOk("module_id"); ok {
		variables["stack"] = toID(moduleID)
	} else {
		return errors.New("either module_id or stack_id must be provided")
	}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not attach policy")
	}

	d.SetId(path.Join(policyID, mutation.AttachPolicy.ID))

	return nil
}

func resourcePolicyAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 {
		return errors.Errorf("unexpected ID: %s", d.Id())
	}

	var query struct {
		Policy *struct {
			Attachment *structs.PolicyAttachment `graphql:"attachedStack(id: $id)"`
		} `graphql:"policy(id: $policy)"`
	}

	variables := map[string]interface{}{
		"policy": toID(idParts[0]),
		"id":     toID(idParts[1]),
	}

	if err := meta.(*Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for policy attachment")
	}

	if query.Policy == nil || query.Policy.Attachment == nil {
		d.SetId("")
		return nil
	}

	attachment := query.Policy.Attachment
	d.Set("policy_id", idParts[0])

	if attachment.IsModule {
		d.Set("module_id", attachment.StackID)
	} else {
		d.Set("stack_id", attachment.StackID)
	}

	return nil
}

func resourcePolicyAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 {
		return errors.Errorf("unexpected ID: %s", d.Id())
	}

	var mutation struct {
		DetachPolicy *structs.PolicyAttachment `graphql:"policyDetach(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(idParts[1])}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not detach policy")
	}

	d.SetId("")

	return nil
}
