package structs

import "github.com/shurcooL/graphql"

// WorkerPool represents the WorkerPoool data relevant to the provider.
type WorkerPool struct {
	ID          string         `graphql:"id"`
	Name        graphql.String `graphql:"name"`
	Description graphql.String `graphql:"description"`
}
