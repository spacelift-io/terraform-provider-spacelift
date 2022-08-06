package spacelift

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

// Deprecated: Used for backwards compatibility.
func resourceStackGCPServiceAccount() *schema.Resource {
	schema := resourceGCPServiceAccount()
	schema.Description = "" +
		"~> **Note:** `spacelift_stack_gcp_service_account` is deprecated. Please use `spacelift_gcp_service_account` instead. The functionality is identical." +
		"\n\n" +
		strings.ReplaceAll(schema.Description, "spacelift_gcp_service_account", "spacelift_stack_gcp_service_account")

	return schema
}

func resourceGCPServiceAccount() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_gcp_service_account` represents a Google Cloud Platform " +
			"service account that's linked to a particular Stack or Module. " +
			"These accounts are created by Spacelift on per-stack basis, and can " +
			"be added as members to as many organizations and projects as needed. " +
			"During a Run or a Task, temporary credentials for those service " +
			"accounts are injected into the environment, which allows " +
			"credential-less GCP Terraform provider setup.",

		CreateContext: resourceGCPServiceAccountCreate,
		ReadContext:   resourceGCPServiceAccountRead,
		UpdateContext: resourceGCPServiceAccountCreate,
		DeleteContext: resourceGCPServiceAccountDelete,

		Importer: &schema.ResourceImporter{StateContext: importIntegration},

		Schema: map[string]*schema.Schema{
			"service_account_email": {
				Type:        schema.TypeString,
				Description: "Email address of the GCP service account dedicated for this stack",
				Computed:    true,
			},
			"module_id": {
				Type:         schema.TypeString,
				Description:  "ID of the module which uses GCP service account credentials",
				ExactlyOneOf: []string{"module_id", "stack_id"},
				Optional:     true,
				ForceNew:     true,
			},
			"stack_id": {
				Type:        schema.TypeString,
				Description: "ID of the stack which uses GCP service account credentials",
				Optional:    true,
				ForceNew:    true,
			},
			"token_scopes": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				MinItems:    1,
				Description: "List of scopes that will be requested when generating temporary GCP service account credentials",
				Required:    true,
				Set:         schema.HashString,
			},
		},
	}
}

func resourceGCPServiceAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreateGCPIntegration struct {
			Activated bool `graphql:"activated"`
		} `graphql:"stackIntegrationGcpCreate(id: $id, tokenScopes: $tokenScopes)"`
	}

	var tokenScopes []graphql.String
	for _, scope := range d.Get("token_scopes").(*schema.Set).List() {
		tokenScopes = append(tokenScopes, graphql.String(scope.(string)))
	}

	var ID string
	if stackID, ok := d.GetOk("stack_id"); ok {
		if err := verifyStack(ctx, stackID.(string), meta); err != nil {
			return diag.FromErr(err)
		}

		ID = stackID.(string)
	} else {
		moduleID := d.Get("module_id").(string)
		if err := verifyModule(ctx, moduleID, meta); err != nil {
			return diag.FromErr(err)
		}

		ID = moduleID
	}

	variables := map[string]interface{}{
		"id":          toID(ID),
		"tokenScopes": tokenScopes,
	}

	if err := meta.(*internal.Client).Mutate(ctx, "GCPServiceAccountCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not generate dedicated GCP role account for the stack: %v", err)
	}

	if !mutation.CreateGCPIntegration.Activated {
		return diag.Errorf("GCP integration not activated")
	}

	if d.Id() == "" {
		d.SetId(ID)
	}

	return resourceGCPServiceAccountRead(ctx, d, meta)
}

func resourceGCPServiceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if _, ok := d.GetOk("module_id"); ok {
		return resourceModuleGCPServiceAccountReadWithHooks(ctx, d, meta, func(_ string) diag.Diagnostics {
			d.SetId("")
			return nil
		})
	}

	return resourceStackGCPServiceAccountReadWithHooks(ctx, d, meta, func(_ string) diag.Diagnostics {
		d.SetId("")
		return nil
	})
}

func resourceGCPServiceAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteGCPIntegration struct {
			Activated bool `graphql:"activated"`
		} `graphql:"stackIntegrationGcpDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "GCPServiceAccountDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete stack GCP service account: %v", err)
	}

	if mutation.DeleteGCPIntegration.Activated {
		return diag.Errorf("did not disable GCP integration, still reporting as activated")
	}

	d.SetId("")

	return nil
}

func resourceModuleGCPServiceAccountReadWithHooks(ctx context.Context, d *schema.ResourceData, meta interface{}, onNil func(message string) diag.Diagnostics) diag.Diagnostics {
	var query struct {
		Module *structs.Module `graphql:"module(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Query(ctx, "ModuleGCPServiceAccountRead", &query, variables); err != nil {
		return diag.Errorf("could not query for module: %v", err)
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

func resourceStackGCPServiceAccountReadWithHooks(ctx context.Context, d *schema.ResourceData, meta interface{}, onNil func(message string) diag.Diagnostics) diag.Diagnostics {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Query(ctx, "StackGCPServiceAccountRead", &query, variables); err != nil {
		return diag.Errorf("could not query for stack: %v", err)
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
