package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

func dataRoleActions() *schema.Resource {
	return &schema.Resource{
		Description: "The `spacelift_role_actions` data source provides a list of all valid actions " +
			"that can be assigned to roles in Spacelift.",

		ReadContext: dataRoleActionsRead,

		Schema: map[string]*schema.Schema{
			"actions": {
				Type:        schema.TypeList,
				Description: "List of all valid actions that can be assigned to roles in Spacelift.",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataRoleActionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	introspectionClient := internal.NewIntrospectionClient(meta.(*internal.Client))

	enumValues, err := introspectionClient.GetEnumValues(ctx, "Action")
	if err != nil {
		return diag.Errorf("could not fetch role actions: %v", err)
	}

	d.SetId("role_actions")
	d.Set("actions", enumValues)

	return nil
}
