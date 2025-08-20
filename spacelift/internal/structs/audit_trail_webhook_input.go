package structs

import "github.com/shurcooL/graphql"

type AuditTrailWebhookInput struct {
	Enabled        graphql.Boolean  `json:"enabled"`
	Endpoint       graphql.String   `json:"endpoint"`
	IncludeRuns    graphql.Boolean  `json:"includeRuns"`
	Secret         graphql.String   `json:"secret"`
	CustomHeaders  *StringMap       `json:"customHeaders"`
	RetryOnFailure *graphql.Boolean `json:"retryOnFailure"`
}

type StringMap struct {
	Entries []KeyValuePair `json:"entries"`
}

func (m StringMap) ToStdMap() map[string]string {
	mapped := make(map[string]string)
	for _, kv := range m.Entries {
		mapped[kv.Key] = kv.Value
	}
	return mapped
}
