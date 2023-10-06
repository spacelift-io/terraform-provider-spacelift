package structs

type UserManagementPolicy struct {
	Space string `graphql:"space"`
	Role  string `graphql:"spaceAccessLevel"`
}

type IdpGroupMapping struct {
	ID                     string                 `graphql:"id"`
	Name                   string                 `graphql:"groupName"`
	UserManagementPolicies []UserManagementPolicy `graphql:"accessRules"`
}
