package structs

// PolicyAttachment is a single policy attachment embedded in a Policy.
type PolicyAttachment struct {
	ID       string `graphql:"id"`
	StackID  string `graphql:"stackId"`
	IsModule bool   `graphql:"isModule"`
}
