package structs

// ContextAttachment is a single context attachment embedded in a Context.
type ContextAttachment struct {
	ID       string `graphql:"id"`
	StackID  string `graphql:"stackId"`
	IsModule bool   `graphql:"isModule"`
	Priority int    `graphql:"priority"`
}
