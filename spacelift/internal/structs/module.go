package structs

// Module represents the Module data relevant to the provider.
type Module struct {
	ID             string  `graphql:"id"`
	Administrative bool    `graphql:"administrative"`
	Branch         string  `graphql:"branch"`
	Description    *string `graphql:"description"`
	Integrations   struct {
		AWS struct {
			AssumedRoleARN              *string `graphql:"assumedRoleArn"`
			AssumeRolePolicyStatement   string  `graphql:"assumeRolePolicyStatement"`
			GenerateCredentialsInWorker bool    `graphql:"generateCredentialsInWorker"`
		} `graphql:"aws"`
		GCP struct {
			ServiceAccountEmail *string  `graphql:"serviceAccountEmail"`
			TokenScopes         []string `graphql:"tokenScopes"`
		} `graphql:"gcp"`
		Webhooks []struct {
			ID       string `graphql:"id"`
			Enabled  bool   `graphql:"enabled"`
			Endpoint string `graphql:"endpoint"`
			Secret   string `graphql:"secret"`
		} `graphql:"webhooks"`
	} `graphql:"integrations"`
	Labels         []string `graphql:"labels"`
	Namespace      string   `graphql:"namespace"`
	Provider       string   `graphql:"provider"`
	Repository     string   `graphql:"repository"`
	SharedAccounts []string `graphql:"sharedAccounts"`
	WorkerPool     *struct {
		ID string `graphql:"id"`
	} `graphql:"workerPool"`
}
