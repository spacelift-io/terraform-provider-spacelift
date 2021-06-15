package structs

import (
	"github.com/shurcooL/graphql"
)

// DriftDetectionIntegrationInput represents the input required to create or update a drift detection integration.
type DriftDetectionIntegrationInput struct {
	Reconcile graphql.Boolean  `json:"reconcile"`
	Schedule  []graphql.String `json:"schedule"`
}
