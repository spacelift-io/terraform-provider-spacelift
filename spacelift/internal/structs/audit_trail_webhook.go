package structs

type AuditTrailWebhook struct {
	AuditTrailWebhookRead
	// CustomHeaders is used only for create/update, API doesn't return them back.
	CustomHeaders StringMap `graphql:"customHeaders"`
}

type AuditTrailWebhookRead struct {
	Enabled     bool   `graphql:"enabled"`
	Endpoint    string `graphql:"endpoint"`
	IncludeRuns bool   `graphql:"includeRuns"`
	Secret      string `graphql:"secret"`
}
