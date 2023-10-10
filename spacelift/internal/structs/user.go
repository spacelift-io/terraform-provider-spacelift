package structs

type User struct {
	ID          string            `graphql:"id"`
	Username    string            `graphql:"username"`
	AccessRules []SpaceAccessRule `graphql:"accessRules"`
}
