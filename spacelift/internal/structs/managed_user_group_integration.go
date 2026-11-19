package structs

type ManagedUserGroupIntegration struct {
	ID              string            `graphql:"id"`
	IntegrationName string            `graphql:"integrationName"`
	IntegrationType string            `graphql:"integrationType"`
	SlackChannelID  string            `graphql:"slackChannelID"`
	AccessRules     []SpaceAccessRule `graphql:"accessRules"`
}
