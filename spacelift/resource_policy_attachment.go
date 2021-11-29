package spacelift

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourcePolicyAttachment() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_policy_attachment` represents a relationship between a " +
			"policy (`spacelift_policy`) and a stack (`spacelift_stack`) or " +
			"module (`spacelift_module`). Each policy can only be attached to a " +
			"stack/module once. `LOGIN` policies are the exception because they " +
			"apply globally and not to individual stacks/modules. An attempt to " +
			"attach one will fail.",

		CreateContext: resourcePolicyAttachmentCreate,
		ReadContext:   resourcePolicyAttachmentRead,
		DeleteContext: resourcePolicyAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourcePolicyAttachmentImport,
		},

		Schema: map[string]*schema.Schema{
			"policy_id": {
				Type:        schema.TypeString,
				Description: "ID of the policy to attach",
				Required:    true,
				ForceNew:    true,
			},
			"module_id": {
				Type:         schema.TypeString,
				Description:  "ID of the module to attach the policy to",
				ExactlyOneOf: []string{"module_id", "stack_id"},
				Optional:     true,
				ForceNew:     true,
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

func resourcePolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		AttachPolicy structs.PolicyAttachment `graphql:"policyAttach(id: $id, stack: $stack)"`
	}

	policyID := d.Get("policy_id").(string)

	variables := map[string]interface{}{"id": toID(policyID)}

	if stackID, ok := d.GetOk("stack_id"); ok {
		if err := verifyStack(ctx, stackID.(string), meta); err != nil {
			return diag.FromErr(err)
		}

		variables["stack"] = toID(stackID)
	} else {
		moduleID := d.Get("module_id").(string)
		if err := verifyModule(ctx, moduleID, meta); err != nil {
			return diag.FromErr(err)
		}

		variables["stack"] = toID(moduleID)
	}

	if err := meta.(*internal.Client).Mutate(ctx, "PolicyAttachmentCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not attach policy: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(path.Join(policyID, mutation.AttachPolicy.ID))

	return nil
}

func resourcePolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	policyID := d.Get("policy_id").(string)

	var projectID string

	if stackID, ok := d.GetOk("stack_id"); ok {
		projectID = stackID.(string)
	} else {
		projectID = d.Get("module_id").(string)
	}

	if attachment, err := resourcePolicyAttachmentFetch(ctx, policyID, projectID, meta); err != nil {
		return diag.FromErr(err)
	} else if attachment == nil {
		d.SetId("")
	}

	return nil
}

func resourcePolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	variables, err := resourcePolicyAttachmentVariables(d)
	if err != nil {
		return diag.FromErr(err)
	}

	var mutation struct {
		DetachPolicy *structs.PolicyAttachment `graphql:"policyDetach(id: $id)"`
	}

	if err := meta.(*internal.Client).Mutate(ctx, "PolicyAttachmentDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not detach policy: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func resourcePolicyAttachmentImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	input := d.Id()

	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("expecting attachment ID as $policyId/$projectId")
	}

	policyID, projectID := parts[0], parts[1]

	attachment, err := resourcePolicyAttachmentFetch(ctx, policyID, projectID, meta)
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

	d.SetId(path.Join(policyID, attachment.ID))
	d.Set("policy_id", policyID)

	return []*schema.ResourceData{d}, nil
}

func resourcePolicyAttachmentVariables(d *schema.ResourceData) (map[string]interface{}, error) {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 {
		return nil, errors.Errorf("unexpected ID: %s", d.Id())
	}

	return map[string]interface{}{"id": toID(idParts[1])}, nil
}

func resourcePolicyAttachmentFetch(ctx context.Context, policyID, projectID string, meta interface{}) (*structs.PolicyAttachment, error) {
	var query struct {
		Policy *struct {
			Attachment *structs.PolicyAttachment `graphql:"attachedStack(id: $id)"`
		} `graphql:"policy(id: $policy)"`
	}

	variables := map[string]interface{}{
		"policy": policyID,
		"id":     toID(projectID),
	}

	if err := meta.(*internal.Client).Query(ctx, "PolicyAttachmentFetch", &query, variables); err != nil {
		return nil, errors.Wrap(err, "could not query for policy attachment")
	}

	if query.Policy == nil {
		return nil, nil
	}

	return query.Policy.Attachment, nil
}
