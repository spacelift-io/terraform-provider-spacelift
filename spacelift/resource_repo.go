package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/vcs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceRepo() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_repo` represents a Spacelift repo - a built-in version " +
			"control system storing infrastructure code directly in Spacelift, " +
			"without an external provider like GitHub or GitLab.\n\n" +
			"Repos have no branches: a stack attached to one always tracks the " +
			"latest commit.",

		CreateContext: resourceRepoCreate,
		ReadContext:   resourceRepoRead,
		UpdateContext: resourceRepoUpdate,
		DeleteContext: resourceRepoDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			repoName: {
				Type:             schema.TypeString,
				Description:      "Name of the repo. The repo's ID (slug) is derived from the name when the repo is created, and does not change when the name changes.",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			repoSpaceID: {
				Type:        schema.TypeString,
				Description: "ID (slug) of the space the repo is in. A repo cannot be moved between spaces.",
				Required:    true,
				ForceNew:    true,
			},
			repoDescription: {
				Type:        schema.TypeString,
				Description: "Free-form repo description for users",
				Optional:    true,
			},
			repoLabels: {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Labels describing the repo",
				Optional:    true,
			},
			repoVCSChecks: {
				Type:        schema.TypeString,
				Description: "VCS checks configured for the repo. Possible values: `INDIVIDUAL`, `AGGREGATED`, `ALL`. Defaults to `INDIVIDUAL`.",
				Optional:    true,
				Default:     vcs.CheckTypeDefault,
			},
			repoCreatedAt: {
				Type:        schema.TypeInt,
				Description: "Unix timestamp of when the repo was created",
				Computed:    true,
			},
			repoUpdatedAt: {
				Type:        schema.TypeInt,
				Description: "Unix timestamp of when the repo was last updated",
				Computed:    true,
			},
			repoStacks: {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IDs (slugs) of the stacks using this repo as their source code provider",
				Computed:    true,
			},
		},
	}
}

func resourceRepoCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var mutation struct {
		CreateRepo structs.Repo `graphql:"repoCreate(input: $input)"`
	}

	variables := map[string]any{
		"input": structs.RepoCreateInput{
			SpaceID:     toID(d.Get(repoSpaceID)),
			Name:        toString(d.Get(repoName)),
			Description: toOptionalString(d.Get(repoDescription)),
			Labels:      setToOptionalStringList(d.Get(repoLabels)),
			VCSChecks:   toOptionalString(d.Get(repoVCSChecks)),
		},
	}

	if err := meta.(*internal.Client).Mutate(ctx, "RepoCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create the repo: %v", internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.CreateRepo.ID)

	return diag.FromErr(populateRepo(d, &mutation.CreateRepo))
}

func resourceRepoRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var query struct {
		Repo *structs.Repo `graphql:"repo(id: $id)"`
	}

	variables := map[string]any{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Query(ctx, "RepoRead", &query, variables); err != nil {
		return diag.Errorf("could not query for repo: %v", err)
	}

	if query.Repo == nil {
		d.SetId("")
		return nil
	}

	return diag.FromErr(populateRepo(d, query.Repo))
}

func resourceRepoUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var mutation struct {
		UpdateRepo structs.Repo `graphql:"repoUpdate(id: $id, input: $input)"`
	}

	variables := map[string]any{
		"id": toID(d.Id()),
		"input": structs.RepoUpdateInput{
			Name:        toString(d.Get(repoName)),
			Description: toOptionalString(d.Get(repoDescription)),
			Labels:      setToOptionalStringList(d.Get(repoLabels)),
			VCSChecks:   toOptionalString(d.Get(repoVCSChecks)),
		},
	}

	if err := meta.(*internal.Client).Mutate(ctx, "RepoUpdate", &mutation, variables); err != nil {
		return diag.Errorf("could not update the repo: %v", internal.FromSpaceliftError(err))
	}

	return diag.FromErr(populateRepo(d, &mutation.UpdateRepo))
}

func resourceRepoDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var mutation struct {
		DeleteRepo graphql.Boolean `graphql:"repoDelete(id: $id)"`
	}

	variables := map[string]any{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "RepoDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete the repo: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}
