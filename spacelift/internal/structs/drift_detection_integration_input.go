package structs

import (
	"github.com/shurcooL/graphql"
)

// DriftDetectionIntegrationInput represents the input required to create or update a drift detection integration.
type DriftDetectionIntegrationInput struct {
	IgnoreState graphql.Boolean  `json:"ignoreState"`
	Reconcile   graphql.Boolean  `json:"reconcile"`
	Schedule    []graphql.String `json:"schedule"`
	Timezone    *graphql.String  `json:"timezone"`
}
