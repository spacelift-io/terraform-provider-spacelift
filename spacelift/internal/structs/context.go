package structs

// Context represents the Context data relevant to the provider.
type Context struct {
	ID          string   `graphql:"id"`
	Description *string  `graphql:"description"`
	Labels      []string `graphql:"labels"`
	Name        string   `graphql:"name"`
	Space       string   `graphql:"space"`
}
