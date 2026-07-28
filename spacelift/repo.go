package spacelift

import (
	"fmt"

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

// validateSpaceliftRepoVCS rejects VCS settings that a Spacelift repo silently ignores.
// Each comparison waits until its values are known, because Get reads a value the plan
// has not resolved yet as empty and would reject it out of hand.
func validateSpaceliftRepoVCS(diff *schema.ResourceDiff) error {
	if blocks, ok := diff.Get("spacelift").([]any); !ok || len(blocks) == 0 {
		return nil
	}

	if diff.NewValueKnown("branch") {
		// The backend hardcodes "main" when resolving the ref, so any other branch degrades run signals.
		if branch := diff.Get("branch").(string); branch != "main" {
			return fmt.Errorf("branch must be \"main\" when using a Spacelift repo, got %q", branch)
		}
	}

	if !diff.NewValueKnown("repository") || !diff.NewValueKnown("spacelift.0.id") {
		return nil
	}

	repoSlug := diff.Get("spacelift.0.id").(string)
	if repository := diff.Get("repository").(string); repository != repoSlug {
		return fmt.Errorf("repository must be the Spacelift repo ID (slug) %q when using a Spacelift repo, got %q", repoSlug, repository)
	}

	return nil
}
