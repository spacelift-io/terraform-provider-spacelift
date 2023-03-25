package structs

// Context represents the Context data relevant to the provider.
type Context struct {
	ID          string   `graphql:"id"`
	Description *string  `graphql:"description"`
	Labels      []string `graphql:"labels"`
	Name        string   `graphql:"name"`
	Space       string   `graphql:"space"`
	Hooks       struct {
		AfterApply    []string `graphql:"afterApply"`
		AfterDestroy  []string `graphql:"afterDestroy"`
		AfterInit     []string `graphql:"afterInit"`
		AfterPerform  []string `graphql:"afterPerform"`
		AfterPlan     []string `graphql:"afterPlan"`
		AfterRun      []string `graphql:"afterRun"`
		BeforeApply   []string `graphql:"beforeApply"`
		BeforeDestroy []string `graphql:"beforeDestroy"`
		BeforeInit    []string `graphql:"beforeInit"`
		BeforePerform []string `graphql:"beforePerform"`
		BeforePlan    []string `graphql:"beforePlan"`
	} `graphql:"hooks"`
}
