package structs

import "github.com/shurcooL/graphql"

type AuditTrailWebhookInput struct {
	Enabled     graphql.Boolean `json:"enabled"`
	Endpoint    graphql.String  `json:"endpoint"`
	IncludeRuns graphql.Boolean `json:"includeRuns"`
	Secret      graphql.String  `json:"secret"`
}
