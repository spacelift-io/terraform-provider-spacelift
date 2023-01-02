package structs

// TerraformProvider represents the data about a Terraform provider.
type TerraformProvider struct {
	ID          string   `graphql:"id"`
	Description *string  `graphql:"description"`
	Labels      []string `graphql:"labels"`
	Public      bool     `graphql:"public"`
	Space       string   `graphql:"space"`
}
