package structs

// Module represents the Module data relevant to the provider.
type Module struct {
	ID                  string       `graphql:"id"`
	Administrative      bool         `graphql:"administrative"`
	Branch              string       `graphql:"branch"`
	Description         *string      `graphql:"description"`
	Integrations        Integrations `graphql:"integrations"`
	Labels              []string     `graphql:"labels"`
	Name                string       `graphql:"name"`
	Namespace           string       `graphql:"namespace"`
	ProjectRoot         *string      `graphql:"projectRoot"`
	ProtectFromDeletion bool         `graphql:"protectFromDeletion"`
	Provider            string       `graphql:"provider"`
	Repository          string       `graphql:"repository"`
	SharedAccounts      []string     `graphql:"sharedAccounts"`
	TerraformProvider   string       `graphql:"terraformProvider"`
	WorkerPool          *struct {
		ID string `graphql:"id"`
	} `graphql:"workerPool"`
}
