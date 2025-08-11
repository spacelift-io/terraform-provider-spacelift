package vcs

import "github.com/shurcooL/graphql"

// CustomVCSInput represents the custom VCS input data.
type CustomVCSInput struct {
	Name           graphql.String    `json:"name"`
	SpaceID        graphql.ID        `json:"spaceID"`
	Labels         *[]graphql.String `json:"labels"`
	Description    *graphql.String   `json:"description"`
	IsDefault      *graphql.Boolean  `json:"isDefault"`
	VCSChecks      *graphql.String   `json:"vcsChecks"`
	UseGitCheckout *graphql.Boolean  `json:"useGitCheckout"`
}
