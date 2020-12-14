package structs

// WorkerPool represents the WorkerPool data relevant to the provider.
type WorkerPool struct {
	ID          string  `graphql:"id"`
	Config      string  `graphql:"config"`
	Name        string  `graphql:"name"`
	Description *string `graphql:"description"`
}
