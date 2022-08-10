package structs

import (
	"github.com/shurcooL/graphql"
)

// ConfigType is a type of configuration element.
type ConfigType string

// ConfigInput represents the input required to create or update a config
// element.
type ConfigInput struct {
	ID        graphql.ID      `json:"id"`
	Type      ConfigType      `json:"type"`
	FileMode  *graphql.String `json:"fileMode"`
	Value     graphql.String  `json:"value"`
	WriteOnly graphql.Boolean `json:"writeOnly"`
}
