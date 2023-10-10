package structs

type User struct {
	ID          string            `graphql:"id"`
	Email       string            `graphql:"email"`
	Username    string            `graphql:"username"`
	AccessRules []SpaceAccessRule `graphql:"accessRules"`
}
