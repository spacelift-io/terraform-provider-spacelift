package spacelift

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Deprecated! Used for backwards compatibility.
func dataStackGCPServiceAccount() *schema.Resource {
	schema := dataGCPServiceAccount()
	schema.Description = "" +
		"~> **Note:** `spacelift_stack_gcp_service_account` is deprecated. Please use `spacelift_gcp_service_account` instead. The functionality is identical." +
		"\n\n" +
		strings.ReplaceAll(schema.Description, "spacelift_gcp_service_account", "spacelift_stack_gcp_service_account")

	return schema
}

func dataGCPServiceAccount() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_gcp_service_account` represents a Google Cloud Platform " +
			"service account that's linked to a particular Stack or Module. " +
			"These accounts are created by Spacelift on per-stack basis, and can " +
			"be added as members to as many organizations and projects as needed. " +
			"During a Run or a Task, temporary credentials for those service " +
			"accounts are injected into the environment, which allows " +
			"credential-less GCP Terraform provider setup.",

		ReadContext: dataGCPServiceAccountRead,

		Schema: map[string]*schema.Schema{
			"service_account_email": {
				Type:        schema.TypeString,
				Description: "email address of the GCP service account dedicated for this stack",
				Computed:    true,
			},
			"module_id": {
				Type:         schema.TypeString,
				Description:  "ID of the stack which uses GCP service account credentials",
				ExactlyOneOf: []string{"module_id", "stack_id"},
				Optional:     true,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack which uses GCP service account credentials",
				Optional:    true,
			},
			"token_scopes": {
				Type:        schema.TypeSet,
				Description: "list of Google API scopes",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
		},
	}
}

func dataGCPServiceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var ret diag.Diagnostics

	if stackID, ok := d.GetOk("stack_id"); ok {
		d.SetId(stackID.(string))
		ret = resourceStackGCPServiceAccountReadWithHooks(ctx, d, meta, func(message string) diag.Diagnostics {
			return diag.Errorf(message)
		})
	} else {
		d.SetId(d.Get("module_id").(string))
		ret = resourceModuleGCPServiceAccountReadWithHooks(ctx, d, meta, func(message string) diag.Diagnostics {
			return diag.Errorf(message)
		})
	}

	if ret.HasError() {
		d.SetId("")
	}

	return ret
}
