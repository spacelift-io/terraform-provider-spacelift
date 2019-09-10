package structs

// Stack represents the Stack data relevant to the provider.
type Stack struct {
	ID                           string  `graphql:"id"`
	Administrative               bool    `graphql:"administrative"`
	AWSAssumedRoleARN            *string `graphql:"awsAssumedRoleARN"`
	AWSAssumeRolePolicyStatement string  `graphql:"awsAssumeRolePolicyStatement"`
	Branch                       string  `graphql:"branch"`
	Description                  *string `graphql:"description"`
	ManagesStateFile             bool    `graphql:"managesStateFile"`
	Name                         string  `graphql:"name"`
	ReadersSlug                  *string `graphql:"readersSlug"`
	Repo                         string  `graphql:"repo"`
	TerraformVersion             *string `graphql:"terraformVersion"`
	WritersSlug                  *string `graphql:"writersSlug"`
}
