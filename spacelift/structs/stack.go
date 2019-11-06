package structs

// Team represents a readers or writers team, though the only thing we really
// care about here is the slug.
type Team struct {
	Slug string `graphql:"slug"`
}

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
	Readers                      *Team   `graphql:"readers"`
	Repository                   string  `graphql:"repository"`
	TerraformVersion             *string `graphql:"terraformVersion"`
	Writers                      *Team   `graphql:"writers"`
}
