package structs

type Run struct {
	ID string `graphql:"id"`
}

type RunDiscardAll struct {
	DiscardedRuns    []Run    `graphql:"discardedRuns"`
	FailedDiscarding []string `graphql:"failedDiscarding"`
}
