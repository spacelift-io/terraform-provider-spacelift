package structs

// GCPIntegration represents the GCP service account integration of a Stack or a
// Module. The API marks the field it comes from as deprecated in favour of OIDC,
// so it is selected only where the value is used.
type GCPIntegration struct {
	ServiceAccountEmail *string  `graphql:"serviceAccountEmail"`
	TokenScopes         []string `graphql:"tokenScopes"`
}

// Integrations represents external integrations for a Stack and a Module.
type Integrations struct {
	AWS struct {
		AssumedRoleARN              *string `graphql:"assumedRoleArn"`
		AssumeRolePolicyStatement   string  `graphql:"assumeRolePolicyStatement"`
		ExternalID                  *string `graphql:"externalID"`
		GenerateCredentialsInWorker bool    `graphql:"generateCredentialsInWorker"`
		DurationSeconds             *int    `graphql:"durationSeconds"`
		Region                      *string `graphql:"region"`
	} `graphql:"aws"`
	DriftDetection struct {
		IgnoreState bool     `graphql:"ignoreState"`
		Reconcile   bool     `graphql:"reconcile"`
		Schedule    []string `graphql:"schedule"`
		Timezone    string   `graphql:"timezone"`
	} `graphql:"driftDetection"`
	Webhooks []struct {
		ID             string `graphql:"id"`
		Enabled        bool   `graphql:"enabled"`
		Endpoint       string `graphql:"endpoint"`
		RetryOnFailure *bool  `graphql:"retryOnFailure"`
	} `graphql:"webhooks"`
}
