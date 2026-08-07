package structs

// IntentProjectExpiryAction represents what happens to an intent project once
// its TTL expiry cleanup finishes.
type IntentProjectExpiryAction string

// IntentProject represents the IntentProject data relevant to the provider.
type IntentProject struct {
	ID          string   `graphql:"id"`
	Name        string   `graphql:"name"`
	Description *string  `graphql:"description"`
	Labels      []string `graphql:"labels"`
	Space       struct {
		ID string `graphql:"id"`
	} `graphql:"space"`
	WorkerPool *struct {
		ID string `graphql:"id"`
	} `graphql:"workerPool"`
	RunnerImage    string                    `graphql:"runnerImage"`
	ExpiresAt      *int                      `graphql:"expiresAt"`
	TTLSeconds     *int                      `graphql:"ttlSeconds"`
	OnExpiryAction IntentProjectExpiryAction `graphql:"onExpiryAction"`
	ArchivedAt     *int                      `graphql:"archivedAt"`
	State          string                    `graphql:"state"`
	CreatedAt      int                       `graphql:"createdAt"`
}
