package spacelift

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
)

func dataStackGCPServiceAccount() *schema.Resource {
	return &schema.Resource{
		Read: dataStackGCPServiceAccountRead,

		Schema: map[string]*schema.Schema{
			"service_account_email": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Email address of the GCP service account dedicated for this stack",
				Computed:    true,
			},
			"stack_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "ID of the stack which uses GCP service account credentials",
				Required:    true,
			},
			"token_scopes": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}
}

func dataStackGCPServiceAccountRead(d *schema.ResourceData, meta interface{}) error {
	d.SetId(d.Get("stack_id").(string))

	err := resourceStackGCPServiceAccountReadWithHooks(d, meta, func(message string) error {
		return errors.New(message)
	})

	if err != nil {
		d.SetId("")
		return err
	}

	return nil
}
