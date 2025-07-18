package structs

import "github.com/shurcooL/graphql"

type APIKeyRoleBinding struct {
	ID       string `graphql:"id"`
	Role     Role   `graphql:"role"`
	SpaceID  string `graphql:"spaceID"`
	APIKeyID string `graphql:"apiKeyID"`
}

type UserGroupRoleBinding struct {
	ID        string    `graphql:"id"`
	RoleID    string    `graphql:"roleID"`
	SpaceID   string    `graphql:"spaceID"`
	UserGroup UserGroup `graphql:"userGroup"`
}

type ApiKeyRoleBindingInput struct { //nolint:staticcheck // The backend type is spelled that way, so we can't change this.
	APIKeyID graphql.ID `json:"apiKeyID"`
	RoleID   graphql.ID `json:"roleID"`
	SpaceID  graphql.ID `json:"spaceID"`
}

type UserGroupRoleBindingInput struct {
	UserGroupID graphql.ID `json:"userGroupID"`
	RoleID      graphql.ID `json:"roleID"`
	SpaceID     graphql.ID `json:"spaceID"`
}

type UserRoleBinding struct {
	ID      string `graphql:"id"`
	RoleID  string `graphql:"roleID"`
	Role    Role   `graphql:"role"`
	SpaceID string `graphql:"spaceID"`
	UserID  string `graphql:"userID"`
	User    User   `graphql:"user"`
}

type UserRoleBindingInput struct {
	RoleID  graphql.ID `json:"roleID"`
	UserID  graphql.ID `json:"userID"`
	SpaceID graphql.ID `json:"spaceID"`
}

type UserRoleBindingBatchInput struct {
	Bindings []UserRoleBindingInput `json:"bindings"`
}
