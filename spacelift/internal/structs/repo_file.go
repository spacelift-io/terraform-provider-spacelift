package structs

import "github.com/shurcooL/graphql"

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
