package structs

// Integrations represents external integrations for a Stack and a Module.
type Integrations struct {
	AWS struct {
		AssumedRoleARN              *string `graphql:"assumedRoleArn"`
		AssumeRolePolicyStatement   string  `graphql:"assumeRolePolicyStatement"`
		ExternalID                  *string `graphql:"externalID"`
		GenerateCredentialsInWorker bool    `graphql:"generateCredentialsInWorker"`
		DurationSeconds             *int    `graphql:"durationSeconds"`
	} `graphql:"aws"`
	DriftDetection struct {
		IgnoreState bool     `graphql:"ignoreState"`
		Reconcile   bool     `graphql:"reconcile"`
		Schedule    []string `graphql:"schedule"`
		Timezone    string   `graphql:"timezone"`
	} `graphql:"driftDetection"`
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
}
