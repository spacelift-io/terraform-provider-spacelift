package spacelift

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
)

// Deprecated! Used for backwards compatibility.
func dataStackGCPServiceAccount() *schema.Resource {
	schema := dataGCPServiceAccount()
	schema.DeprecationMessage = "use spacelift_gcp_service_account data source instead"

	return schema
}

func dataGCPServiceAccount() *schema.Resource {
	return &schema.Resource{
		Read: dataGCPServiceAccountRead,

		Schema: map[string]*schema.Schema{
			"service_account_email": {
				Type:        schema.TypeString,
				Description: "Email address of the GCP service account dedicated for this stack",
				Computed:    true,
			},
			"module_id": {
				Type:          schema.TypeString,
				Description:   "ID of the stack which uses GCP service account credentials",
				ConflictsWith: []string{"stack_id"},
				Optional:      true,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack which uses GCP service account credentials",
				Optional:    true,
			},
			"token_scopes": {
				Type:        schema.TypeSet,
				Description: "List of Google API scopes",
				Elem: &schema.Schema{
					Type:        schema.TypeString,
					Description: "Name (it's actually always a URL) of a single Google API scope",
				},
				Computed: true,
			},
		},
	}
}

func dataGCPServiceAccountRead(d *schema.ResourceData, meta interface{}) error {
	var err error
	d.SetId(d.Get("stack_id").(string))

	if stackID, ok := d.GetOk("stack_id"); ok {
		d.SetId(stackID.(string))
		err = resourceStackGCPServiceAccountReadWithHooks(d, meta, func(message string) error {
			return errors.New(message)
		})
	}

	if moduleID, ok := d.GetOk("module_id"); ok {
		d.SetId(moduleID.(string))
		err = resourceModuleGCPServiceAccountReadWithHooks(d, meta, func(message string) error {
			return errors.New(message)
		})
	}

	if err != nil {
		d.SetId("")
		return err
	}

	return nil
}
