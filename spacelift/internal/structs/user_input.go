package structs

import "github.com/shurcooL/graphql"

type UserInviteInput struct {
	Email       graphql.String         `json:"email"`
	Username    graphql.String         `json:"username"`
	AccessRules []SpaceAccessRuleInput `json:"accessRules"`
}

type UserUpdateInput struct {
	AccessRules []SpaceAccessRuleInput `json:"accessRules"`
}
