package spacelift

import "github.com/shurcooL/graphql"

func toID(input interface{}) graphql.ID {
	return graphql.ID(input)
}

func toOptionalString(input interface{}) *graphql.String {
	return graphql.NewString(toString(input))
}

func toString(input interface{}) graphql.String {
	return graphql.String(input.(string))
}

func toOptionalInt(input interface{}) *graphql.Int {
	v := graphql.Int(input.(int))
	return graphql.NewInt(v)
}
