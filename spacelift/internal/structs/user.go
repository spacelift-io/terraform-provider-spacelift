package structs

type User struct {
	ID              string            `graphql:"id"`
	InvitationEmail string            `graphql:"invitationEmail"`
	Username        string            `graphql:"username"`
	AccessRules     []SpaceAccessRule `graphql:"accessRules"`
}
