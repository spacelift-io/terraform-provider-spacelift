package structs

type RunType string

const (
	RunTypeProposed RunType = "PROPOSED"
	RunTypeTracked  RunType = "TRACKED"
)

type Run struct {
	ID string `graphql:"id"`
}

type RunDiscardAll struct {
	DiscardedRuns    []Run    `graphql:"discardedRuns"`
	FailedDiscarding []string `graphql:"failedDiscarding"`
}
