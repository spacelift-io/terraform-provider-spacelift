package structs

type Action string

type Role struct {
	ID                string   `graphql:"id"`
	Slug              string   `graphql:"slug"`
	IsSystem          bool     `graphql:"isSystem"`
	Name              string   `graphql:"name"`
	Description       string   `graphql:"description"`
	Actions           []Action `graphql:"actions"`
	RoleBindingsCount int      `graphql:"roleBindingsCount"`
}
