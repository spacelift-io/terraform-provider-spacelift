package predicates

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/search"
)

func BuildBoolean(d *schema.ResourceData, schemaName string, optionalPredicateName ...string) (out []search.QueryPredicate) {
	field, ok := d.GetOk(schemaName)
	if !ok {
		return
	}

	for _, element := range field.([]interface{}) {
		predicate := element.(map[string]interface{})
		if predicate["equals"] == nil {
			continue
		}

		out = append(out, search.QueryPredicate{
			Field: getPredicateName(schemaName, optionalPredicateName),
			Constraint: search.QueryFieldConstraint{
				BooleanEquals: &[]graphql.Boolean{graphql.Boolean(predicate["equals"].(bool))},
			},
		})
	}

	return
}

func BuildStringOrEnum(d *schema.ResourceData, isEnum bool, schemaName string, optionalPredicateName ...string) (out []search.QueryPredicate) {
	field, ok := d.GetOk(schemaName)
	if !ok {
		return nil
	}

	for _, element := range field.([]interface{}) {
		predicate := element.(map[string]interface{})
		if predicate["any_of"] == nil {
			continue
		}

		var matches []graphql.String
		for _, element := range predicate["any_of"].([]interface{}) {
			matches = append(matches, graphql.String(element.(string)))
		}

		var constraint search.QueryFieldConstraint
		if isEnum {
			constraint.EnumEquals = &matches
		} else {
			constraint.StringMatches = &matches
		}

		out = append(out, search.QueryPredicate{
			Field:      getPredicateName(schemaName, optionalPredicateName),
			Constraint: constraint,
		})
	}

	return
}

func getPredicateName(schemaName string, optionalPredicateName []string) graphql.String {
	if len(optionalPredicateName) == 0 {
		return graphql.String(schemaName)
	}

	return graphql.String(optionalPredicateName[0])
}
