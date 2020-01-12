package spacelift

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
)

func resourceStackGCPServiceAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceStackGCPServiceAccountCreate,
		Read:   resourceStackGCPServiceAccountRead,
		Update: resourceStackGCPServiceAccountCreate,
		Delete: resourceStackGCPServiceAccountDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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
				ForceNew:    true,
			},
			"token_scopes": &schema.Schema{
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    1,
				Description: "List of scopes that will be requested when generating temporary GCP service account credentials",
				Required:    true,
				Set:         schema.HashString,
			},
		},
	}
}

func resourceStackGCPServiceAccountCreate(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		CreateGCPIntegration struct {
			Activated bool `graphql:"activated"`
		} `graphql:"stackIntegrationGcpCreate(id: $id, tokenScopes: $tokenScopes)"`
	}

	var tokenScopes []graphql.String
	for _, scope := range d.Get("token_scopes").(*schema.Set).List() {
		tokenScopes = append(tokenScopes, graphql.String(scope.(string)))
	}

	stackID := d.Get("stack_id").(string)

	variables := map[string]interface{}{
		"id":          toID(stackID),
		"tokenScopes": tokenScopes,
	}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not generate dedicated GCP role account for the stack")
	}

	if !mutation.CreateGCPIntegration.Activated {
		return errors.New("GCP integration not activated")
	}

	if d.Id() == "" {
		d.SetId(stackID)
	}

	return resourceStackGCPServiceAccountRead(d, meta)
}

func resourceStackGCPServiceAccountRead(d *schema.ResourceData, meta interface{}) error {
	return resourceStackGCPServiceAccountReadWithHooks(d, meta, func(_ string) error {
		d.SetId("")
		return nil
	})
}

func resourceStackGCPServiceAccountDelete(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		DeleteGCPIntegration struct {
			Activated bool `graphql:"activated"`
		} `graphql:"stackIntegrationGcpDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not delete stack GCP service account")
	}

	if mutation.DeleteGCPIntegration.Activated {
		return errors.New("did not disable GCP integration, still reporting as activated")
	}

	d.SetId("")

	return nil
}

func resourceStackGCPServiceAccountReadWithHooks(d *schema.ResourceData, meta interface{}, onNil func(message string) error) error {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for stack")
	}

	if query.Stack == nil {
		return onNil("stack not found")
	}

	integration := query.Stack.Integrations.GCP
	serviceAccountEmail := integration.ServiceAccountEmail

	if serviceAccountEmail == nil {
		return onNil("GCP integration not activated")
	}

	d.Set("service_account_email", *serviceAccountEmail)

	tokenScopes := schema.NewSet(schema.HashString, []interface{}{})
	for _, scope := range integration.TokenScopes {
		tokenScopes.Add(scope)
	}
	d.Set("token_scopes", tokenScopes)

	return nil
}
