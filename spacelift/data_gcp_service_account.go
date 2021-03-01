package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Deprecated! Used for backwards compatibility.
func dataStackGCPServiceAccount() *schema.Resource {
	schema := dataGCPServiceAccount()
	schema.DeprecationMessage = "use spacelift_gcp_service_account data source instead"

	return schema
}

func dataGCPServiceAccount() *schema.Resource {
	return &schema.Resource{
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
