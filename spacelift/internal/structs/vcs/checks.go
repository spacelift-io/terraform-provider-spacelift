package vcs

const (
	// CheckTypeIndividual represents the individual VCS check type.
	CheckTypeIndividual = "INDIVIDUAL"
	// CheckTypeAggregated represents the aggregated VCS check type.
	CheckTypeAggregated = "AGGREGATED"
	// CheckTypeAll represents the summary of individual and aggregated VCS checks.
	CheckTypeAll = "ALL"
	// CheckTypeDefault is the default VCS check type.
	CheckTypeDefault = CheckTypeIndividual
)
