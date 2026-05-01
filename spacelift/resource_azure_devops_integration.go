package spacelift

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/vcs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceAzureDevopsIntegration() *schema.Resource {
	return &schema.Resource{
		Description:   "`spacelift_azure_devops_integration` represents an integration with an Azure DevOps organization",
		CreateContext: resourceAzureDevopsIntegrationCreate,
		ReadContext:   resourceAzureDevopsIntegrationRead,
		UpdateContext: resourceAzureDevopsIntegrationUpdate,
		DeleteContext: resourceAzureDevopsIntegrationDelete,

		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, meta any) error {
			if diff.HasChange(azureDevopsIsDefault) {
				isDefault := diff.Get(azureDevopsIsDefault).(bool)
				spaceID := diff.Get(azureDevopsSpaceID).(string)
				if isDefault && spaceID != "root" {
					return fmt.Errorf(`the default integration must be in the space "root" not in %q`, spaceID)
				}
			}
			return nil
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			azureDevopsID: {
				Type:        schema.TypeString,
				Description: "Azure DevOps integration id.",
				Computed:    true,
			},
			azureDevopsName: {
				Type:             schema.TypeString,
				Description:      "The friendly name of the integration",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			azureDevopsDescription: {
				Type:             schema.TypeString,
				Description:      "Description of the integration",
				Optional:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			azureDevopsOrganizationURL: {
				Type:             schema.TypeString,
				Description:      "Organization URL where API requests will be sent",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			azureDevopsUserFacingHost: {
				Type:             schema.TypeString,
				Description:      "User facing host URL. Defaults to the organization URL if not set. Set this when using VCS agents.",
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			azureDevopsPersonalAccessToken: {
				Type:             schema.TypeString,
				Description:      "The Azure DevOps personal access token",
				Optional:         true,
				Sensitive:        true,
				Deprecated:       "`personal_access_token` is deprecated. Please use personal_access_token_wo in combination with personal_access_token_wo_version",
				ValidateDiagFunc: validations.DisallowEmptyString,
				ConflictsWith:    []string{azureDevopsPersonalAccessTokenWo, azureDevopsPersonalAccessTokenWoVer},
				AtLeastOneOf:     []string{azureDevopsPersonalAccessToken, azureDevopsPersonalAccessTokenWo},
			},
			azureDevopsPersonalAccessTokenWo: {
				Type:          schema.TypeString,
				Description:   "`personal_access_token_wo` the Azure DevOps personal access token. The personal_access_token_wo is not stored in the state. Modify personal_access_token_wo_version to trigger an update. This field requires Terraform/OpenTofu 1.11+.",
				Optional:      true,
				Sensitive:     true,
				WriteOnly:     true,
				ConflictsWith: []string{azureDevopsPersonalAccessToken},
				RequiredWith:  []string{azureDevopsPersonalAccessTokenWoVer},
				AtLeastOneOf:  []string{azureDevopsPersonalAccessToken, azureDevopsPersonalAccessTokenWo},
			},
			azureDevopsPersonalAccessTokenWoVer: {
				Type:          schema.TypeString,
				Description:   "Used together with personal_access_token_wo to trigger an update to the personal access token. Increment this value when an update to personal_access_token_wo is required. This field requires Terraform/OpenTofu 1.11+.",
				Optional:      true,
				ConflictsWith: []string{azureDevopsPersonalAccessToken},
				RequiredWith:  []string{azureDevopsPersonalAccessTokenWo},
			},
			azureDevopsLabels: {
				Type:        schema.TypeSet,
				Description: "Labels to set on the integration",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			azureDevopsSpaceID: {
				Type:             schema.TypeString,
				Description:      "ID (slug) of the space the integration is in; Default: `root`",
				Optional:         true,
				Default:          "root",
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			azureDevopsIsDefault: {
				Type:        schema.TypeBool,
				Description: "Is the Azure DevOps integration the default for all spaces? If set to `true` the space must be set to `root` in `space_id` or left empty which uses the default",
				Optional:    true,
				Default:     false,
				ForceNew:    true,
			},
			azureDevopsWebhookPassword: {
				Type:        schema.TypeString,
				Description: "Password for webhooks originating from Azure DevOps repositories",
				Computed:    true,
				Sensitive:   true,
			},
			azureDevopsWebhookURL: {
				Type:        schema.TypeString,
				Description: "URL for webhooks originating from Azure DevOps repositories",
				Computed:    true,
			},
			azureDevopsVCSChecks: {
				Type:        schema.TypeString,
				Description: "VCS checks configured for Azure DevOps repositories. Possible values: INDIVIDUAL, AGGREGATED, ALL. Defaults to INDIVIDUAL.",
				Optional:    true,
				Default:     vcs.CheckTypeDefault,
			},
			azureDevopsUseGitCheckout: {
				Type:        schema.TypeBool,
				Description: "Indicates whether the integration should use git checkout. If false source code will be downloaded using the VCS API. Defaults to true.",
				Optional:    true,
				Computed:    true,
			},
			azureDevopsAccessibleProjects: {
				Type:        schema.TypeSet,
				Description: "Restrict the integration to specific Azure DevOps projects. Leave empty to allow all projects in the organization.",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
		},
	}
}

func resourceAzureDevopsIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	token, diags := internal.ExtractWriteOnlyField(azureDevopsPersonalAccessToken, azureDevopsPersonalAccessTokenWo, azureDevopsPersonalAccessTokenWoVer, d)
	if diags != nil {
		return diags
	}

	var mutation struct {
		CreateAzureDevOpsRepoIntegration structs.AzureDevOpsRepoIntegration `graphql:"azureDevOpsRepoIntegrationCreate(organizationURL: $organizationURL, userFacingHost: $userFacingHost, personalAccessToken: $personalAccessToken, customInput: $customInput, accessibleProjects: $accessibleProjects)"`
	}

	variables := map[string]any{
		"organizationURL":     toString(d.Get(azureDevopsOrganizationURL)),
		"userFacingHost":      optionalAzureDevopsUserFacingHost(d),
		"personalAccessToken": toString(token),
		"customInput": &vcs.CustomVCSInput{
			Name:           toString(d.Get(azureDevopsName)),
			IsDefault:      toOptionalBool(d.Get(azureDevopsIsDefault)),
			SpaceID:        toID(d.Get(azureDevopsSpaceID)),
			Labels:         setToOptionalStringList(d.Get(azureDevopsLabels)),
			Description:    toOptionalString(d.Get(azureDevopsDescription)),
			VCSChecks:      toOptionalString(d.Get(azureDevopsVCSChecks)),
			UseGitCheckout: getOptionalBool(d, azureDevopsUseGitCheckout),
		},
		"accessibleProjects": setToOptionalStringList(d.Get(azureDevopsAccessibleProjects)),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "AzureDevOpsRepoIntegrationCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create the Azure DevOps integration: %v", internal.FromSpaceliftError(err))
	}

	fillAzureDevopsIntegrationResults(d, &mutation.CreateAzureDevOpsRepoIntegration)

	return nil
}

func resourceAzureDevopsIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var query struct {
		AzureDevOpsIntegration *structs.AzureDevOpsRepoIntegration `graphql:"azureDevOpsRepoIntegration(id: $id)"`
	}

	variables := map[string]any{"id": d.Id()}
	if err := meta.(*internal.Client).Query(ctx, "AzureDevOpsIntegrationRead", &query, variables); err != nil {
		return diag.Errorf("could not query for the Azure DevOps integration: %v", err)
	}

	if query.AzureDevOpsIntegration == nil {
		d.SetId("")
	} else {
		fillAzureDevopsIntegrationResults(d, query.AzureDevOpsIntegration)
	}

	return nil
}

func resourceAzureDevopsIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	token, diags := internal.ExtractWriteOnlyField(azureDevopsPersonalAccessToken, azureDevopsPersonalAccessTokenWo, azureDevopsPersonalAccessTokenWoVer, d)
	if diags != nil {
		return diags
	}

	var mutation struct {
		UpdateAzureDevOpsRepoIntegration structs.AzureDevOpsRepoIntegration `graphql:"azureDevOpsRepoIntegrationUpdate(organizationURL: $organizationURL, personalAccessToken: $personalAccessToken, customInput: $customInput, accessibleProjects: $accessibleProjects)"`
	}

	variables := map[string]any{
		"organizationURL":     toString(d.Get(azureDevopsOrganizationURL)),
		"personalAccessToken": optionalStringFromValue(token),
		"customInput": &vcs.CustomVCSUpdateInput{
			ID:             toID(d.Id()),
			SpaceID:        toID(d.Get(azureDevopsSpaceID)),
			Description:    toOptionalString(d.Get(azureDevopsDescription)),
			Labels:         setToOptionalStringList(d.Get(azureDevopsLabels)),
			VCSChecks:      toOptionalString(d.Get(azureDevopsVCSChecks)),
			UseGitCheckout: getOptionalBool(d, azureDevopsUseGitCheckout),
		},
		"accessibleProjects": setToOptionalStringList(d.Get(azureDevopsAccessibleProjects)),
	}

	var ret diag.Diagnostics

	if err := meta.(*internal.Client).Mutate(ctx, "AzureDevOpsRepoIntegrationUpdate", &mutation, variables); err != nil {
		ret = append(ret, diag.Errorf("could not update the Azure DevOps integration: %v", internal.FromSpaceliftError(err))...)
	}

	fillAzureDevopsIntegrationResults(d, &mutation.UpdateAzureDevOpsRepoIntegration)

	return ret
}

func resourceAzureDevopsIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var mutation struct {
		DeleteAzureDevOpsRepoIntegration *structs.AzureDevOpsRepoIntegration `graphql:"azureDevOpsRepoIntegrationDelete(id: $id)"`
	}

	variables := map[string]any{
		"id": toID(d.Id()),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "AzureDevOpsRepoIntegrationDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete Azure DevOps integration: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func fillAzureDevopsIntegrationResults(d *schema.ResourceData, integration *structs.AzureDevOpsRepoIntegration) {
	d.SetId(integration.ID)
	d.Set(azureDevopsID, integration.ID)
	d.Set(azureDevopsName, integration.Name)
	d.Set(azureDevopsDescription, integration.Description)
	d.Set(azureDevopsIsDefault, integration.IsDefault)
	d.Set(azureDevopsSpaceID, integration.Space.ID)
	d.Set(azureDevopsOrganizationURL, integration.OrganizationURL)
	d.Set(azureDevopsUserFacingHost, integration.UserFacingHost)
	d.Set(azureDevopsWebhookPassword, integration.WebhookPassword)
	d.Set(azureDevopsWebhookURL, integration.WebhookURL)
	d.Set(azureDevopsLabels, integration.Labels)
	d.Set(azureDevopsVCSChecks, integration.VCSChecks)
	d.Set(azureDevopsUseGitCheckout, integration.UseGitCheckout)
	d.Set(azureDevopsAccessibleProjects, integration.AccessibleProjects)
}

func optionalAzureDevopsUserFacingHost(d *schema.ResourceData) *graphql.String {
	if value, ok := d.GetOk(azureDevopsUserFacingHost); ok {
		return toOptionalString(value)
	}

	return nil
}

func optionalStringFromValue(value string) *graphql.String {
	if value == "" {
		return nil
	}

	graphqlValue := graphql.String(value)
	return &graphqlValue
}
