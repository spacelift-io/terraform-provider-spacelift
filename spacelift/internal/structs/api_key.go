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
	OIDCSettings  *APIKeyOIDCSettings `graphql:"oidcSettings"`
}

type APIKeyOIDCSettings struct {
	Issuer            string       `graphql:"issuer"`
	ClientID          string       `graphql:"clientId"`
	SubjectExpression string       `graphql:"subjectExpression"`
	ClaimMapping      ClaimMapping `graphql:"claimMapping"`
}

type ClaimMapping struct {
	Entries []ClaimMappingEntry `graphql:"entries"`
}

type ClaimMappingEntry struct {
	Name  string `graphql:"name"`
	Value string `graphql:"value"`
}

type ApiKeyInput struct { //nolint:staticcheck // The backend type is spelled that way, so we can't change this.
	Admin     graphql.Boolean  `json:"admin"`
	Name      graphql.String   `json:"name"`
	IDPGroups []graphql.String `json:"teams"`
	OIDC      *APIKeyInputOIDC `json:"oidc,omitempty"`
}

type APIKeyInputOIDC struct {
	Issuer            graphql.String     `json:"issuer"`
	ClientID          graphql.String     `json:"clientId"`
	SubjectExpression graphql.String     `json:"subjectExpression"`
	ClaimMappings     *ClaimMappingInput `json:"claimMappings,omitempty"`
}

type ClaimMappingInput struct {
	Entries []ClaimMappingEntryInput `json:"entries"`
}

type ClaimMappingEntryInput struct {
	Name  graphql.String `json:"name"`
	Value graphql.String `json:"value"`
}

type ApiKeyUpdateInput struct { //nolint:staticcheck // The backend type is spelled that way, so we can't change this.
	Name              *graphql.String    `json:"name,omitempty"`
	IDPGroups         []graphql.String   `json:"teams,omitempty"`
	Description       *graphql.String    `json:"description,omitempty"`
	OIDCClaimMappings *ClaimMappingInput `json:"oidcClaimMappings,omitempty"`
}
