package structs

import "github.com/shurcooL/graphql"

type ManagedUserUpdateInput struct {
	ID          graphql.ID             `json:"id"`
	AccessRules []SpaceAccessRuleInput `json:"accessRules"`
}
