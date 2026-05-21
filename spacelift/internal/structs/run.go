package structs

type RunType string

const (
	RunTypeTracked  RunType = "TRACKED"
	RunTypeProposed RunType = "PROPOSED"
)

type Run struct {
	ID string `graphql:"id"`
}

type RunDiscardAll struct {
	DiscardedRuns    []Run    `graphql:"discardedRuns"`
	FailedDiscarding []string `graphql:"failedDiscarding"`
}
