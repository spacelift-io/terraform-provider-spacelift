package structs

// WorkerPool represents the WorkerPoool data relevant to the provider.
type WorkerPool struct {
	ID          string  `graphql:"id"`
	Name        string  `graphql:"name"`
	Description *string `graphql:"description"`
}
