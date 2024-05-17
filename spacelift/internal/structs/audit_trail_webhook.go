package structs

type AuditTrailWebhook struct {
	Enabled       bool      `graphql:"enabled"`
	Endpoint      string    `graphql:"endpoint"`
	IncludeRuns   bool      `graphql:"includeRuns"`
	Secret        string    `graphql:"secret"`
	CustomHeaders StringMap `graphql:"customHeaders"`
}
