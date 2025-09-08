package structs

import "github.com/shurcooL/graphql"

// PolicyType represents a policy type.
type PolicyType string

type PolicyEngineType string

// Policy represents Policy data relevant to the provider.
type Policy struct {
	ID          string   `graphql:"id"`
	Labels      []string `graphql:"labels"`
	Name        string   `graphql:"name"`
	Body        string   `graphql:"body"`
	Type        string   `graphql:"type"`
	Space       string   `graphql:"space"`
	Description string   `graphql:"description"`
	EngineType  string   `graphql:"engineType"`
}

type PolicyCreateInput struct {
	PolicyUpdateInput
	Type PolicyType `json:"type"`
}

func NewPolicyCreateInput(name, body graphql.String, policyType PolicyType) PolicyCreateInput {
	return PolicyCreateInput{
		PolicyUpdateInput: PolicyUpdateInput{
			Name: name,
			Body: body,
		},
		Type: policyType,
	}
}

type PolicyUpdateInput struct {
	Name        graphql.String    `json:"name"`
	Body        graphql.String    `json:"body"`
	Labels      *[]graphql.String `json:"labels"`
	Space       *graphql.ID       `json:"space"`
	Description *graphql.String   `json:"description"`
	EngineType  *PolicyEngineType `json:"engineType"`
}

func NewPolicyUpdateInput(name, body graphql.String) PolicyUpdateInput {
	return PolicyUpdateInput{
		Name: name,
		Body: body,
	}
}
