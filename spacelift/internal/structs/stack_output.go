package structs

// StackOutput represents the metadata of a stack output.
type StackOutput struct {
	ID          string `graphql:"id"`
	Description string `graphql:"description"`
	Sensitive   bool   `graphql:"sensitive"`
}
