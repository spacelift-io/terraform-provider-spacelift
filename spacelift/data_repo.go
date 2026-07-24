package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataRepo() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_repo` returns details about a Spacelift repo, the built-in " +
			"VCS provider storing source code directly in Spacelift.",

		ReadContext: dataRepoRead,

		Schema: map[string]*schema.Schema{
			repoID: {
				Type:             schema.TypeString,
				Description:      "ID (slug) of the repo",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			repoName: {
				Type:        schema.TypeString,
				Description: "Name of the repo",
				Computed:    true,
			},
			repoDescription: {
				Type:        schema.TypeString,
				Description: "Free-form repo description for users",
				Computed:    true,
			},
			repoLabels: {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Labels describing the repo",
				Computed:    true,
			},
			repoSpaceID: {
				Type:        schema.TypeString,
				Description: "ID (slug) of the space the repo is in",
				Computed:    true,
			},
			repoVCSChecks: {
				Type:        schema.TypeString,
				Description: "VCS checks configured for the repo. One of `INDIVIDUAL`, `AGGREGATED` or `ALL`.",
				Computed:    true,
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

func dataRepoRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var query struct {
		Repo *structs.Repo `graphql:"repo(id: $id)"`
	}

	repoSlug := d.Get(repoID).(string)
	variables := map[string]any{"id": toID(repoSlug)}

	if err := meta.(*internal.Client).Query(ctx, "RepoRead", &query, variables); err != nil {
		return diag.Errorf("could not query for repo: %v", err)
	}

	if query.Repo == nil {
		return diag.Errorf("could not find repo %s", repoSlug)
	}

	d.SetId(query.Repo.ID)

	if err := d.Set(repoID, query.Repo.ID); err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(populateRepo(d, query.Repo))
}

func populateRepo(d *schema.ResourceData, repo *structs.Repo) error {
	for key, value := range flattenRepo(repo) {
		if err := d.Set(key, value); err != nil {
			return err
		}
	}

	return nil
}

func flattenRepo(repo *structs.Repo) map[string]any {
	labels := schema.NewSet(schema.HashString, []any{})
	for _, label := range repo.Labels {
		labels.Add(label)
	}

	stacks := make([]any, len(repo.Stacks))
	for i, stack := range repo.Stacks {
		stacks[i] = stack
	}

	return map[string]any{
		repoName:        repo.Name,
		repoDescription: repo.Description,
		repoLabels:      labels,
		repoSpaceID:     repo.Space.ID,
		repoVCSChecks:   repo.VCSChecks,
		repoCreatedAt:   repo.CreatedAt,
		repoUpdatedAt:   repo.UpdatedAt,
		repoStacks:      stacks,
	}
}
