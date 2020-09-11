package structs

// PolicyType represents a policy type.
type PolicyType string

// Policy represents Policy data relevant to the provider.
type Policy struct {
	ID   string `graphql:"id"`
	Name string `graphql:"name"`
	Body string `graphql:"body"`
	Type string `graphql:"type"`
}
