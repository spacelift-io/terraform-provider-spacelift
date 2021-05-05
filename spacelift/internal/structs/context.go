package structs

// Context represents the Context data relevant to the provider.
type Context struct {
	ID          string  `graphql:"id"`
	Description *string `graphql:"description"`
	Name        string  `graphql:"name"`
}
