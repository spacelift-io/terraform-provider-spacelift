package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/search"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

const reposPageSize = 50

func dataRepos() *schema.Resource {
	repoSchema := dataRepo().Schema

	repoSchema[repoID] = &schema.Schema{
		Type:        schema.TypeString,
		Description: "ID (slug) of the repo",
		Computed:    true,
	}

	return &schema.Resource{
		Description: "" +
			"`spacelift_repos` returns the Spacelift repos in a single space. " +
			"Repos are not inherited by child spaces, so only repos created " +
			"directly in the given space are returned.",

		ReadContext: dataReposRead,

		Schema: map[string]*schema.Schema{
			repoSpaceID: {
				Type:             schema.TypeString,
				Description:      "ID (slug) of the space to list repos from",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			repoLabels: {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Required labels to match",
				Optional:    true,
			},
			repos: {
				Type:        schema.TypeList,
				Description: "List of repos in the space",
				Elem:        &schema.Resource{Schema: repoSchema},
				Computed:    true,
			},
		},
	}
}

func dataReposRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var query struct {
		Repos struct {
			Edges []struct {
				Node structs.Repo `graphql:"node"`
			} `graphql:"edges"`
			PageInfo search.PageInfo `graphql:"pageInfo"`
		} `graphql:"repos(spaceID: $spaceID, first: $first, after: $after)"`
	}

	spaceID := d.Get(repoSpaceID).(string)
	first := graphql.Int(reposPageSize)

	variables := map[string]any{
		"spaceID": toID(spaceID),
		"first":   &first,
		"after":   (*graphql.String)(nil),
	}

	var found []structs.Repo

	for {
		if err := meta.(*internal.Client).Query(ctx, "ReposPage", &query, variables); err != nil {
			return diag.Errorf("could not query for repos: %v", err)
		}

		for _, edge := range query.Repos.Edges {
			found = append(found, edge.Node)
		}

		if !query.Repos.PageInfo.HasNextPage {
			break
		}

		after := graphql.String(query.Repos.PageInfo.EndCursor)
		variables["after"] = &after
	}

	matching := internal.FilterByRequiredLabels(d, found, func(repo structs.Repo) []string { return repo.Labels })

	result := make([]any, len(matching))
	for i, repo := range matching {
		flattened := flattenRepo(&repo)
		flattened[repoID] = repo.ID
		result[i] = flattened
	}

	d.SetId("spacelift-repos-" + spaceID)

	if err := d.Set(repos, result); err != nil {
		return diag.Errorf("could not set repos: %v", err)
	}

	return nil
}
