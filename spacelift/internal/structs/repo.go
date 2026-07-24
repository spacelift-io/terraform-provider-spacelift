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

// RepoFile represents a single version of a file in a Repo.
type RepoFile struct {
	FilePath  string  `graphql:"filePath"`
	IsDeleted bool    `graphql:"isDeleted"`
	SizeBytes *int    `graphql:"sizeBytes"`
	FileMode  *string `graphql:"fileMode"`
	Revision  struct {
		SHA string `graphql:"sha"`
	} `graphql:"revision"`
	Content *struct {
		SHA256Hash  string  `graphql:"sha256Hash"`
		Content     *string `graphql:"content"`
		SizeBytes   int     `graphql:"sizeBytes"`
		IsEncrypted bool    `graphql:"isEncrypted"`
	} `graphql:"content"`
}

// Revision represents a commit in a Repo.
type Revision struct {
	ID         string `graphql:"id"`
	SHA        string `graphql:"sha"`
	AuthorName string `graphql:"authorName"`
}

// RevisionCreateInput represents input relevant to committing files to a Repo.
type RevisionCreateInput struct {
	RepoID      graphql.ID          `json:"repoID"`
	Message     graphql.String      `json:"message"`
	Description *graphql.String     `json:"description"`
	AuthorName  graphql.String      `json:"authorName"`
	AuthorEmail *graphql.String     `json:"authorEmail"`
	Files       []RevisionFileInput `json:"files"`
}

// RevisionFileInput represents a single file within a commit. Content is
// base64-encoded, and files left out of a commit carry forward unchanged.
type RevisionFileInput struct {
	Path     graphql.String   `json:"path"`
	Content  *graphql.String  `json:"content"`
	Encrypt  *graphql.Boolean `json:"encrypt"`
	FileMode *graphql.String  `json:"fileMode"`
	Action   *graphql.String  `json:"action"`
}
