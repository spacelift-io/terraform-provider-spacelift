package spacelift

import "github.com/shurcooL/graphql"

func ToOptionalString(input interface{}) *graphql.String {
	return graphql.NewString(ToString(input))
}

func ToString(input interface{}) graphql.String {
	return graphql.String(input.(string))
}
