package structs

import "github.com/shurcooL/graphql"

type ManagedUserInviteInput struct {
	InvitationEmail *graphql.String        `json:"invitationEmail"`
	Username        graphql.String         `json:"username"`
	AccessRules     []SpaceAccessRuleInput `json:"accessRules"`
}

type ManagedUserUpdateInput struct {
	ID          graphql.ID             `json:"id"`
	AccessRules []SpaceAccessRuleInput `json:"accessRules"`
}
