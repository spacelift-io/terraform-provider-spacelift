package spacelift

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

func resourceSecurityEmail() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_security_email` represents an email address that " +
			"receives notifications about security issues in Spacelift.",

		CreateContext: resourceSecurityEmailCreate,
		ReadContext:   resourceSecurityEmailRead,
		UpdateContext: resourceSecurityEmailUpdate,
		DeleteContext: resourceSecurityEmailDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"email": {
				Type:        schema.TypeString,
				Description: "Email address to which the security notifications are sent",
				Required:    true,
			},
		},
	}
}

func resourceSecurityEmailCreate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var mutation struct {
		SecurityEmail *string `graphql:"accountUpdateSecurityEmail(securityEmail: $securityEmail)"`
	}

	variables := map[string]interface{}{"securityEmail": toString(data.Get("email"))}
	if err := i.(*internal.Client).Mutate(ctx, "AccountUpdateSecurityEmail", &mutation, variables); err != nil {
		return diag.Errorf("could not create security email: %v", err)
	}

	data.SetId(time.Now().String())

	return resourceSecurityEmailRead(ctx, data, i)
}

func resourceSecurityEmailRead(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var query struct {
		SecurityEmail *string `graphql:"securityEmail"`
	}
	if err := i.(*internal.Client).Query(ctx, "SecurityEmail", &query, nil); err != nil {
		return diag.Errorf("could not query for security email: %v", err)
	}

	if query.SecurityEmail == nil {
		data.SetId("")
		return nil
	}

	data.Set("email", query.SecurityEmail)

	return nil
}

func resourceSecurityEmailUpdate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var mutation struct {
		SecurityEmail *string `graphql:"accountUpdateSecurityEmail(securityEmail: $email)"`
	}
	variables := map[string]interface{}{
		"email": toString(data.Get("email")),
	}
	if err := i.(*internal.Client).Mutate(ctx, "AccountUpdateSecurityEmail", &mutation, variables); err != nil {
		return diag.Errorf("could not create security email: %v", err)
	}

	return resourceSecurityEmailRead(ctx, data, i)
}

func resourceSecurityEmailDelete(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	data.SetId("")
	return diag.Diagnostics{{
		Severity: diag.Warning,
		Summary:  "deleting security email is not supported, the resource has been removed from the state, but is left configured in Spacelift",
	}}
}
