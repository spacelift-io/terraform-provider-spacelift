package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceAzureIntegrationConsent() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_azure_integration_consent` records externally granted Azure admin consent in Spacelift. " +
			"Create this resource after granting consent in Azure, for example with " +
			"`azuread_service_principal_delegated_permission_grant`. It does not grant or verify permissions in Azure. " +
			"Deleting this resource clears Spacelift's consent confirmation and prevents the integration from being used; " +
			"it does not revoke the Azure consent grant or delete the service principal. " +
			"The caller needs permission to confirm or clear consent in the integration's space, or the Space admin role. " +
			"This resource requires a Spacelift version supporting the `azureIntegrationConsentConfirm` and " +
			"`azureIntegrationConsentClear` GraphQL mutations. " +
			"Use this resource's `admin_consent_provided` attribute for outputs during the first apply, because " +
			"the parent integration's consent status may not refresh until the next plan.",
		CreateContext: resourceAzureIntegrationConsentCreate,
		ReadContext:   resourceAzureIntegrationConsentRead,
		DeleteContext: resourceAzureIntegrationConsentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"integration_id": {
				Type:             schema.TypeString,
				Description:      "ID of the Azure integration whose admin consent has been granted",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"admin_consent_provided": {
				Type:        schema.TypeBool,
				Description: "Indicates whether Spacelift has recorded admin consent for the integration",
				Computed:    true,
			},
		},
	}
}

func resourceAzureIntegrationConsentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var mutation struct {
		Confirm struct {
			AdminConsentProvided bool `graphql:"adminConsentProvided"`
		} `graphql:"azureIntegrationConsentConfirm(id: $id)"`
	}
	variables := map[string]any{"id": toID(d.Get("integration_id"))}
	if err := meta.(*internal.Client).Mutate(ctx, "AzureIntegrationConsentConfirm", &mutation, variables); err != nil {
		return diag.Errorf("could not confirm admin consent for the Azure integration: %v", internal.FromSpaceliftError(err))
	}
	if !mutation.Confirm.AdminConsentProvided {
		return diag.Errorf("Spacelift did not confirm admin consent for the Azure integration")
	}

	d.SetId(d.Get("integration_id").(string))
	return diag.FromErr(d.Set("admin_consent_provided", mutation.Confirm.AdminConsentProvided))
}

func resourceAzureIntegrationConsentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var query struct {
		AzureIntegration *struct {
			AdminConsentProvided bool `graphql:"adminConsentProvided"`
		} `graphql:"azureIntegration(id: $id)"`
	}
	variables := map[string]any{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "AzureIntegrationConsentRead", &query, variables); err != nil {
		return diag.Errorf("could not query admin consent for the Azure integration: %v", err)
	}
	if query.AzureIntegration == nil || !query.AzureIntegration.AdminConsentProvided {
		return removeFromState(d, "Azure integration consent confirmation")
	}

	if err := d.Set("integration_id", d.Id()); err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(d.Set("admin_consent_provided", query.AzureIntegration.AdminConsentProvided))
}

func resourceAzureIntegrationConsentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var mutation struct {
		Clear *struct {
			AdminConsentProvided bool `graphql:"adminConsentProvided"`
		} `graphql:"azureIntegrationConsentClear(id: $id)"`
	}
	variables := map[string]any{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Mutate(ctx, "AzureIntegrationConsentClear", &mutation, variables); err != nil {
		return diag.Errorf("could not clear admin consent confirmation for the Azure integration: %v", internal.FromSpaceliftError(err))
	}
	if mutation.Clear != nil && mutation.Clear.AdminConsentProvided {
		return diag.Errorf("Spacelift did not clear admin consent confirmation for the Azure integration")
	}

	d.SetId("")
	return nil
}
