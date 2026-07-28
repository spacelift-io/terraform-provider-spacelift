package structs

import "github.com/shurcooL/graphql"

// Repo represents a Spacelift repo, the built-in VCS provider storing source
// code directly in Spacelift. Its ID is the repo slug, not a database ID.
type Repo struct {
	ID          string   `graphql:"id"`
	Name        string   `graphql:"name"`
	Description string   `graphql:"description"`
	Labels      []string `graphql:"labels"`
	Space       struct {
		ID string `graphql:"id"`
	} `graphql:"space"`
	VCSChecks string   `graphql:"vcsChecks"`
	CreatedAt int      `graphql:"createdAt"`
	UpdatedAt int      `graphql:"updatedAt"`
	Stacks    []string `graphql:"stacks"`
}

// RepoCreateInput represents input relevant to creating a Repo. The slug is
// derived from the name by the backend and cannot be chosen.
type RepoCreateInput struct {
	SpaceID     graphql.ID        `json:"spaceID"`
	Name        graphql.String    `json:"name"`
	Description *graphql.String   `json:"description"`
	Labels      *[]graphql.String `json:"labels"`
	VCSChecks   *graphql.String   `json:"vcsChecks"`
}

// RepoUpdateInput represents input relevant to updating a Repo. There is no
// space field: a Repo cannot be moved between spaces.
type RepoUpdateInput struct {
	Name        graphql.String    `json:"name"`
	Description *graphql.String   `json:"description"`
	Labels      *[]graphql.String `json:"labels"`
	VCSChecks   *graphql.String   `json:"vcsChecks"`
}
