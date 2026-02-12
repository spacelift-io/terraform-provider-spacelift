package spacelift

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/vcs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceGitLabIntegration() *schema.Resource {
	return &schema.Resource{
		Description:   "`spacelift_gitlab_integration` represents an integration with a GitLab instance",
		CreateContext: resourceGitLabIntegrationCreate,
		ReadContext:   resourceGitLabIntegrationRead,
		UpdateContext: resourceGitLabIntegrationUpdate,
		DeleteContext: resourceGitLabIntegrationDelete,

		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
			if diff.HasChange(gitLabIsDefault) {
				isDefault := diff.Get(gitLabIsDefault).(bool)
				spaceID := diff.Get(gitLabSpaceID).(string)
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
			gitLabID: {
				Type:        schema.TypeString,
				Description: "GitLab integration id.",
				Computed:    true,
			},
			gitLabName: {
				Type:             schema.TypeString,
				Description:      "The friendly name of the integration",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			gitLabDescription: {
				Type:             schema.TypeString,
				Description:      "Description of the integration",
				Optional:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			gitLabAPIHost: {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "API host URL",
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			gitLabUserFacingHost: {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "User facing host URL.",
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			gitLabToken: {
				Type:             schema.TypeString,
				Description:      "The GitLab API Token",
				Optional:         true,
				Sensitive:        true,
				Deprecated:       "`private_token` is deprecated. Please use private_token_wo in combination with private_token_wo_version",
				ValidateDiagFunc: validations.DisallowEmptyString,
				ConflictsWith:    []string{gitLabTokenWo, gitLabTokenWoVersion},
				AtLeastOneOf:     []string{gitLabToken, gitLabTokenWo},
			},
			gitLabTokenWo: {
				Type:          schema.TypeString,
				Description:   "`private_token_wo` the GitLab API Token .The private_token_wo is not stored in the state. Modify private_token_wo_version to trigger an update. This field requires Terraform/OpenTofu 1.11+.",
				Sensitive:     true,
				Optional:      true,
				WriteOnly:     true,
				ConflictsWith: []string{gitLabToken},
				RequiredWith:  []string{gitLabTokenWoVersion},
				AtLeastOneOf:  []string{gitLabToken, gitLabTokenWo},
			},
			gitLabTokenWoVersion: {
				Type:          schema.TypeString,
				Description:   "Used together with private_token_wo to trigger an update to the private_token. Increment this value when an update to private_token_wo is required. This field requires Terraform/OpenTofu 1.11+.",
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{gitLabToken},
				RequiredWith:  []string{gitLabTokenWo},
			},
			gitLabLabels: {
				Type:        schema.TypeSet,
				Description: "Labels to set on the integration",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			gitLabSpaceID: {
				Type:             schema.TypeString,
				Description:      "ID (slug) of the space the integration is in; Default: `root`",
				Optional:         true,
				Default:          "root",
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			gitLabIsDefault: {
				Type:        schema.TypeBool,
				Description: "Is the GitLab integration the default for all spaces? If set to `true` the space must be set to `root` in `" + gitLabSpaceID + "` or left empty which uses the default",
				Optional:    true,
				Default:     false,
				ForceNew:    true, // unable to update isDefault flag
			},
			gitLabWebhookSecret: {
				Type:        schema.TypeString,
				Description: "Secret for webhooks originating from GitLab repositories",
				Computed:    true,
				Sensitive:   true,
			},
			gitLabWebhookURL: {
				Type:        schema.TypeString,
				Description: "URL for webhooks originating from GitLab repositories",
				Computed:    true,
			},
			gitLabVCSChecks: {
				Type:        schema.TypeString,
				Description: "VCS checks configured for GitLab repositories. Possible values: INDIVIDUAL, AGGREGATED, ALL. Defaults to INDIVIDUAL.",
				Optional:    true,
				Default:     vcs.CheckTypeDefault,
			},
			gitLabUseGitCheckout: {
				Type:        schema.TypeBool,
				Description: "Indicates whether the integration should use git checkout. If false source code will be downloaded using the VCS API. Defaults to true.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func resourceGitLabIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	token, digas := internal.ExtractWriteOnlyField(gitLabToken, gitLabTokenWo, gitLabTokenWoVersion, d)
	if digas != nil {
		return digas
	}

	var mutation struct {
		CreateGitLabIntegration structs.GitLabIntegration `graphql:"gitlabIntegrationCreate(apiHost: $apiHost, userFacingHost: $userFacingHost, privateToken: $token, customInput: $customInput)"`
	}

	variables := map[string]interface{}{
		"customInput": &vcs.CustomVCSInput{
			Name:           toString(d.Get(gitLabName)),
			IsDefault:      toOptionalBool(d.Get(gitLabIsDefault)),
			SpaceID:        toString(d.Get(gitLabSpaceID)),
			Labels:         setToOptionalStringList(d.Get(gitLabLabels)),
			Description:    toOptionalString(d.Get(gitLabDescription)),
			VCSChecks:      toOptionalString(d.Get(gitLabVCSChecks)),
			UseGitCheckout: getOptionalBool(d, gitLabUseGitCheckout),
		},
		"apiHost":        toString(d.Get(gitLabAPIHost)),
		"userFacingHost": toString(d.Get(gitLabUserFacingHost)),
		"token":          toString(token),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "GitLabIntegrationCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create the GitLab integration: %v", internal.FromSpaceliftError(err))
	}

	fillGitLabIntegrationResults(d, &mutation.CreateGitLabIntegration)

	return nil
}

func resourceGitLabIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		GitLabIntegration *structs.GitLabIntegration `graphql:"gitlabIntegration(id: $id)"`
	}

	variables := map[string]interface{}{"id": d.Id()}
	if err := meta.(*internal.Client).Query(ctx, "GitLabIntegrationRead", &query, variables); err != nil {
		return diag.Errorf("could not query for the gitlab integration: %v", err)
	}

	if query.GitLabIntegration == nil {
		d.SetId("")
	} else {
		fillGitLabIntegrationResults(d, query.GitLabIntegration)
	}

	return nil
}

func resourceGitLabIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	token, digas := internal.ExtractWriteOnlyField(gitLabToken, gitLabTokenWo, gitLabTokenWoVersion, d)
	if digas != nil {
		return digas
	}

	var mutation struct {
		UpdateGitLabIntegration structs.GitLabIntegration `graphql:"gitlabIntegrationUpdate(apiHost: $apiHost, userFacingHost: $userFacingHost, privateToken: $privateToken, customInput: $customInput)"`
	}

	variables := map[string]interface{}{
		"privateToken":   toOptionalString(token),
		"apiHost":        toString(d.Get(gitLabAPIHost)),
		"userFacingHost": toString(d.Get(gitLabUserFacingHost)),
		"customInput": &vcs.CustomVCSUpdateInput{
			ID:             toID(d.Id()),
			SpaceID:        toString(d.Get(gitLabSpaceID)),
			Description:    toOptionalString(d.Get(gitLabDescription)),
			Labels:         setToOptionalStringList(d.Get(gitLabLabels)),
			VCSChecks:      toOptionalString(d.Get(gitLabVCSChecks)),
			UseGitCheckout: getOptionalBool(d, gitLabUseGitCheckout),
		},
	}

	var ret diag.Diagnostics

	if err := meta.(*internal.Client).Mutate(ctx, "GitLabIntegrationUpdate", &mutation, variables); err != nil {
		ret = append(ret, diag.Errorf("could not update the GitLab integration: %v", internal.FromSpaceliftError(err))...)
	}

	fillGitLabIntegrationResults(d, &mutation.UpdateGitLabIntegration)

	return ret
}

func resourceGitLabIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteGitLabIntegration *structs.GitLabIntegration `graphql:"gitlabIntegrationDelete(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": toID(d.Id()),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "GitLabIntegrationDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete GitLab integration: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func fillGitLabIntegrationResults(d *schema.ResourceData, gitLabIntegration *structs.GitLabIntegration) {
	d.SetId(gitLabIntegration.ID)
	d.Set(gitLabName, gitLabIntegration.Name)
	d.Set(gitLabSpaceID, gitLabIntegration.Space.ID)
	d.Set(gitLabIsDefault, gitLabIntegration.IsDefault)
	d.Set(gitLabDescription, gitLabIntegration.Description)
	d.Set(gitLabAPIHost, gitLabIntegration.APIHost)
	d.Set(gitLabUserFacingHost, gitLabIntegration.UserFacingHost)
	d.Set(gitLabWebhookURL, gitLabIntegration.WebhookURL)
	d.Set(gitLabWebhookSecret, gitLabIntegration.WebhookSecret)
	d.Set(gitLabVCSChecks, gitLabIntegration.VCSChecks)
	d.Set(gitLabUseGitCheckout, gitLabIntegration.UseGitCheckout)

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range gitLabIntegration.Labels {
		labels.Add(label)
	}
	d.Set(gitLabLabels, labels)
}
