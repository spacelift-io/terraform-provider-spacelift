package structs

import "github.com/shurcooL/graphql"

type ManagedUserGroupIntegrationCreateInput struct {
	IntegrationType graphql.String         `json:"integrationType"`
	IntegrationName graphql.String         `json:"integrationName"`
	SlackChannelID  graphql.String         `json:"slackChannelID"`
	AccessRules     []SpaceAccessRuleInput `json:"accessRules"`
}

type ManagedUserGroupIntegrationUpdateInput struct {
	ID              graphql.ID             `json:"id"`
	IntegrationName graphql.String         `json:"integrationName"`
	SlackChannelID  graphql.String         `json:"slackChannelID"`
	AccessRules     []SpaceAccessRuleInput `json:"accessRules"`
}
