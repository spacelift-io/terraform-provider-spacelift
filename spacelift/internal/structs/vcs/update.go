package vcs

import "github.com/shurcooL/graphql"

// CustomVCSUpdateInput represents the custom VCS update input data.
type CustomVCSUpdateInput struct {
	ID             graphql.ID        `json:"id"`
	SpaceID        graphql.ID        `json:"space"`
	Labels         *[]graphql.String `json:"labels"`
	Description    *graphql.String   `json:"description"`
	VCSChecks      *graphql.String   `json:"vcsChecks"`
	UseGitCheckout *graphql.Boolean  `json:"useGitCheckout"`
}
