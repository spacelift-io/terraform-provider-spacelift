package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/vcs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceGitLabIntegration() *schema.Resource {
	return &schema.Resource{
		Description:   "`spacelift_gitlab_integration` represents an integration with an GitLab instance",
		CreateContext: resourceGitLabIntegrationCreate,
		ReadContext:   resourceGitLabIntegrationRead,
		UpdateContext: resourceGitLabIntegrationUpdate,
		DeleteContext: resourceGitLabIntegrationDelete,

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
				Description:      "The friendly name of the integration",
				Optional:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			gitLabAPIHost: {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "API host URL",
				ValidateFunc: validation.IsURLWithHTTPS,
			},
			gitLabUserFacingHost: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "User facing host URL. If not set defaults to `" + gitLabAPIHost + "`",
				ValidateFunc: validation.IsURLWithHTTPS,
			},
			gitLabToken: {
				Type:             schema.TypeString,
				Description:      "The GitLab API Token",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
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
				Description:      "ID (slug) of the space the integration is in.",
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			gitLabIsDefault: {
				Type:        schema.TypeBool,
				Description: "Is the GitLab integration the default for all spaces? If set to `true` a a space must be specified in `" + gitLabSpaceID + "`",
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
		},
	}
}

func resourceGitLabIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreateGitLabIntegration structs.GitLabIntegration `graphql:"gitlabIntegrationCreate(apiHost: $apiHost, userFacingHost: $userFacingHost, privateToken: $token, customInput: $customInput)"`
	}

	spaceID, hasSpaceID := d.GetOk(gitLabSpaceID)
	if !hasSpaceID {
		spaceID = "root"
	}
	userFacingHost, hasUserFacingHost := d.GetOk(gitLabUserFacingHost)
	if !hasUserFacingHost {
		userFacingHost = d.Get(gitLabAPIHost)
	}
	variables := map[string]interface{}{
		"customInput": &vcs.CustomVCSInput{
			Name:        toString(d.Get(gitLabName)),
			IsDefault:   toOptionalBool(d.Get(gitLabIsDefault)),
			SpaceID:     toString(spaceID),
			Labels:      toOptionalStringList(d.Get(gitLabLabels)),
			Description: toOptionalString(d.Get(gitLabDescription)),
		},
		"apiHost":        toString(d.Get(gitLabAPIHost)),
		"userFacingHost": toString(userFacingHost),
		"token":          toString(d.Get(gitLabToken)),
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
		return diag.Errorf("could not query for the bitbucket datacenter integration: %v", err)
	}

	if query.GitLabIntegration == nil {
		d.SetId("")
	} else {
		fillGitLabIntegrationResults(d, query.GitLabIntegration)
	}

	return nil
}

func resourceGitLabIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UpdateGitLabIntegration structs.GitLabIntegration `graphql:"gitlabIntegrationUpdate(apiHost: $apiHost, userFacingHost: $userFacingHost, privateToken: $privateToken, customInput: $customInput)"`
	}

	userFacingHost, hasUserFacingHost := d.GetOk(gitLabUserFacingHost)
	if !hasUserFacingHost {
		userFacingHost = d.Get(gitLabAPIHost)
	}

	variables := map[string]interface{}{
		"privateToken":   toOptionalString(d.Get(gitLabToken)),
		"apiHost":        toString(d.Get(gitLabAPIHost)),
		"userFacingHost": toString(userFacingHost),
		"customInput": &vcs.CustomVCSUpdateInput{
			ID:          toID(d.Id()),
			SpaceID:     toString(d.Get(gitLabSpaceID)),
			Description: toOptionalString(d.Get(gitLabDescription)),
			Labels:      toOptionalStringList(d.Get(gitLabLabels)),
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

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range gitLabIntegration.Labels {
		labels.Add(label)
	}
	d.Set(gitLabLabels, labels)
}
