package spacelift

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

const (
	repoID          = "repo_id"
	repoName        = "name"
	repoDescription = "description"
	repoLabels      = "labels"
	repoSpaceID     = "space_id"
	repoVCSChecks   = "vcs_checks"
	repoCreatedAt   = "created_at"
	repoUpdatedAt   = "updated_at"
	repoStacks      = "stacks"
	repos           = "repos"
)

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
