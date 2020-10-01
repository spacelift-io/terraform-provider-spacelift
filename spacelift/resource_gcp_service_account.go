package spacelift

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

// Deprecated! Used for backwards compatibility.
func resourceStackGCPServiceAccount() *schema.Resource {
	schema := resourceGCPServiceAccount()
	schema.DeprecationMessage = "use spacelift_gcp_service_account resource instead"

	return schema
}

func resourceGCPServiceAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceGCPServiceAccountCreate,
		Read:   resourceGCPServiceAccountRead,
		Update: resourceGCPServiceAccountCreate,
		Delete: resourceGCPServiceAccountDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"service_account_email": {
				Type:        schema.TypeString,
				Description: "Email address of the GCP service account dedicated for this stack",
				Computed:    true,
			},
			"module_id": {
				Type:          schema.TypeString,
				Description:   "ID of the module which uses GCP service account credentials",
				Optional:      true,
				ConflictsWith: []string{"stack_id"},
				ForceNew:      true,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack which uses GCP service account credentials",
				Optional:    true,
				ForceNew:    true,
			},
			"token_scopes": {
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

func resourceGCPServiceAccountCreate(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		CreateGCPIntegration struct {
			Activated bool `graphql:"activated"`
		} `graphql:"stackIntegrationGcpCreate(id: $id, tokenScopes: $tokenScopes)"`
	}

	var tokenScopes []graphql.String
	for _, scope := range d.Get("token_scopes").(*schema.Set).List() {
		tokenScopes = append(tokenScopes, graphql.String(scope.(string)))
	}

	var id string
	if stackID, ok := d.GetOk("stack_id"); ok {
		id = stackID.(string)
	} else if moduleID, ok := d.GetOk("module_id"); ok {
		id = moduleID.(string)
	} else {
		return errors.New("either module_id or stack_id must be provided")
	}

	variables := map[string]interface{}{
		"id":          toID(id),
		"tokenScopes": tokenScopes,
	}

	if err := meta.(*internal.Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not generate dedicated GCP role account for the stack")
	}

	if !mutation.CreateGCPIntegration.Activated {
		return errors.New("GCP integration not activated")
	}

	if d.Id() == "" {
		d.SetId(id)
	}

	return resourceGCPServiceAccountRead(d, meta)
}

func resourceGCPServiceAccountRead(d *schema.ResourceData, meta interface{}) error {
	if _, ok := d.GetOk("module_id"); ok {
		return resourceModuleGCPServiceAccountReadWithHooks(d, meta, func(_ string) error {
			d.SetId("")
			return nil
		})
	}

	return resourceStackGCPServiceAccountReadWithHooks(d, meta, func(_ string) error {
		d.SetId("")
		return nil
	})
}

func resourceGCPServiceAccountDelete(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		DeleteGCPIntegration struct {
			Activated bool `graphql:"activated"`
		} `graphql:"stackIntegrationGcpDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not delete stack GCP service account")
	}

	if mutation.DeleteGCPIntegration.Activated {
		return errors.New("did not disable GCP integration, still reporting as activated")
	}

	d.SetId("")

	return nil
}

func resourceModuleGCPServiceAccountReadWithHooks(d *schema.ResourceData, meta interface{}, onNil func(message string) error) error {
	var query struct {
		Module *structs.Module `graphql:"module(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for module")
	}

	if query.Module == nil {
		return onNil("module not found")
	}

	integration := query.Module.Integrations.GCP
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

func resourceStackGCPServiceAccountReadWithHooks(d *schema.ResourceData, meta interface{}, onNil func(message string) error) error {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Query(&query, variables); err != nil {
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
