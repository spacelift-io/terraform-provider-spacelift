package structs

type SpaceAccessRule struct {
	Space            string `graphql:"space"`
	SpaceAccessLevel string `graphql:"spaceAccessLevel"`
}

type UserGroup struct {
	ID          string            `graphql:"id"`
	Name        string            `graphql:"groupName"`
	AccessRules []SpaceAccessRule `graphql:"accessRules"`
	Description string            `graphql:"description"`
}
