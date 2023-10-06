package structs

import "github.com/shurcooL/graphql"

type Role string

type UserManagementPolicyInput struct {
	Space graphql.ID `json:"space"`
	Role  Role       `json:"spaceAccessLevel"`
}

type IdpGroupMappingCreateInput struct {
	Name                        graphql.String              `json:"groupName"`
	UserManagementPoliciesInput []UserManagementPolicyInput `json:"accessRules"`
}

type IdpGroupMappingUpdateInput struct {
	ID                          graphql.ID                  `json:"id"`
	UserManagementPoliciesInput []UserManagementPolicyInput `json:"accessRules"`
}
