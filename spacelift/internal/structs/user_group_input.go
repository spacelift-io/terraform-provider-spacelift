package structs

import "github.com/shurcooL/graphql"

type SpaceAccessLevel string

type SpaceAccessRuleInput struct {
	Space            graphql.ID       `json:"space"`
	SpaceAccessLevel SpaceAccessLevel `json:"spaceAccessLevel"`
}

type ManagedUserGroupCreateInput struct {
	Name        graphql.String         `json:"groupName"`
	AccessRules []SpaceAccessRuleInput `json:"accessRules"`
}

type ManagedUserGroupUpdateInput struct {
	ID          graphql.ID             `json:"id"`
	AccessRules []SpaceAccessRuleInput `json:"accessRules"`
}
