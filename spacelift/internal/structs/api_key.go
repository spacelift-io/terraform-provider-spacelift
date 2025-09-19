package structs

import "github.com/shurcooL/graphql"

type APIKeyType string

type APIKey struct {
	ID            string              `graphql:"id"`
	Admin         bool                `graphql:"admin"`
	Name          string              `graphql:"name"`
	Secret        string              `graphql:"secret"`
	Type          APIKeyType          `graphql:"type"`
	IDPGroups     []string            `graphql:"teams"`
	AccessRules   []SpaceAccessRule   `graphql:"accessRules"`
	IsMachineUser bool                `graphql:"isMachineUser"`
	RoleBindings  []APIKeyRoleBinding `graphql:"apiKeyRoleBindings"`
}

type ApiKeyInput struct { //nolint:staticcheck // The backend type is spelled that way, so we can't change this.
	Admin     graphql.Boolean  `json:"admin"`
	Name      graphql.String   `json:"name"`
	IDPGroups []graphql.String `json:"teams"`
	OIDC      *APIKeyInputOIDC `json:"oidc,omitempty"`
}

type APIKeyInputOIDC struct {
	Issuer            graphql.String `json:"issuer"`
	ClientID          graphql.String `json:"clientId"`
	SubjectExpression graphql.String `json:"subjectExpression"`
}

type ApiKeyUpdateInput struct { //nolint:staticcheck // The backend type is spelled that way, so we can't change this.
	Name        *graphql.String  `json:"name,omitempty"`
	IDPGroups   []graphql.String `json:"teams,omitempty"`
	Description *graphql.String  `json:"description,omitempty"`
}
