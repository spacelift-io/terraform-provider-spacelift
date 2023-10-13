package structs

import "github.com/shurcooL/graphql"

type ManagedUserInviteInput struct {
	Email       graphql.String         `json:"email"`
	Username    graphql.String         `json:"username"`
	AccessRules []SpaceAccessRuleInput `json:"accessRules"`
}

type ManagedUserUpdateInput struct {
	AccessRules []SpaceAccessRuleInput `json:"accessRules"`
}
