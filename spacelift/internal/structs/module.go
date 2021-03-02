package structs

// Module represents the Module data relevant to the provider.
type Module struct {
	ID             string       `graphql:"id"`
	Administrative bool         `graphql:"administrative"`
	Branch         string       `graphql:"branch"`
	Description    *string      `graphql:"description"`
	Integrations   Integrations `graphql:"integrations"`
	Labels         []string     `graphql:"labels"`
	Namespace      string       `graphql:"namespace"`
	Provider       string       `graphql:"provider"`
	Repository     string       `graphql:"repository"`
	SharedAccounts []string     `graphql:"sharedAccounts"`
	WorkerPool     *struct {
		ID string `graphql:"id"`
	} `graphql:"workerPool"`
}
