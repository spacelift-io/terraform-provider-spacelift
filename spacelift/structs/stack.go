package structs

// Stack represents the Stack data relevant to the provider.
type Stack struct {
	ID             string  `graphql:"id"`
	Administrative bool    `graphql:"administrative"`
	Autodeploy     bool    `graphql:"autodeploy"`
	Branch         string  `graphql:"branch"`
	Description    *string `graphql:"description"`
	Integrations   struct {
		AWS struct {
			AssumedRoleARN            *string `graphql:"assumedRoleArn"`
			AssumeRolePolicyStatement string  `graphql:"assumeRolePolicyStatement"`
		} `graphql:"aws"`
		GCP struct {
			ServiceAccountEmail *string  `graphql:"serviceAccountEmail"`
			TokenScopes         []string `graphql:"tokenScopes"`
		} `graphql:"gcp"`
	} `graphql:"integrations"`
	Labels           []string `graphql:"labels"`
	ManagesStateFile bool     `graphql:"managesStateFile"`
	Name             string   `graphql:"name"`
	Namespace        string   `graphql:"namespace"`
	Provider         string   `graphql:"provider"`
	Repository       string   `graphql:"repository"`
	TerraformVersion *string  `graphql:"terraformVersion"`
}
