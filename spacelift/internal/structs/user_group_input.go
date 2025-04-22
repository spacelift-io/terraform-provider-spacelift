package structs

import "github.com/shurcooL/graphql"

type SpaceAccessLevel string

type SpaceAccessRuleInput struct {
	Space            graphql.ID       `json:"space"`
	SpaceAccessLevel SpaceAccessLevel `json:"spaceAccessLevel"`
}

type ManagedUserGroupCreateInput struct {
	Name        graphql.String         `json:"groupName"`
	Description graphql.String         `json:"description"`
	AccessRules []SpaceAccessRuleInput `json:"accessRules"`
}

type ManagedUserGroupUpdateInput struct {
	ID          graphql.ID             `json:"id"`
	Description graphql.String         `json:"description"`
	AccessRules []SpaceAccessRuleInput `json:"accessRules"`
}
