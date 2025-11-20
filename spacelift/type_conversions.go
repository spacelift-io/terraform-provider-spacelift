package spacelift

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func toBool(input interface{}) graphql.Boolean {
	return graphql.Boolean(input.(bool))
}

func toOptionalBool(input interface{}) *graphql.Boolean {
	return graphql.NewBoolean(toBool(input))
}

func toID(input interface{}) graphql.ID {
	return graphql.ID(input)
}

func toOptionalString(input interface{}) *graphql.String {
	return graphql.NewString(toString(input))
}

func toString(input interface{}) graphql.String {
	return graphql.String(input.(string))
}

func toMap(input interface{}) map[string]interface{} {
	return input.(map[string]interface{})
}

func toOptionalInt(input interface{}) *graphql.Int {
	v := graphql.Int(input.(int)) //nolint:gosec // safe: value known to fit in int32
	return graphql.NewInt(v)
}

func listToStringList(input interface{}) []graphql.String {
	if input == nil {
		return nil
	}

	v := input.([]interface{})
	var arr []graphql.String
	if len(v) > 0 {
		for _, expr := range v {
			arr = append(arr, toString(expr))
		}
	}
	return arr
}

func listToOptionalStringList(input interface{}) *[]graphql.String {
	l := listToStringList(input)
	if l == nil {
		return nil
	}
	return &l
}

func setToOptionalStringList(input interface{}) *[]graphql.String {
	if input == nil {
		return nil
	}

	if labelSet, ok := input.(*schema.Set); ok {
		var labels []graphql.String
		for _, label := range labelSet.List() {
			labels = append(labels, toString(label))
		}
		return &labels
	}

	return nil
}

func toOptionalStringMap(input interface{}) *structs.StringMap {
	var customHeaders structs.StringMap
	for k, v := range toMap(input) {
		customHeaders.Entries = append(customHeaders.Entries, structs.KeyValuePair{
			Key:   k,
			Value: v.(string),
		})
	}
	if len(customHeaders.Entries) == 0 {
		return nil
	}
	return &customHeaders
}

// getOptionalBool checks if user has explicitly set a boolean value in the configuration and returns it as a pointer to graphql.Boolean.
// If the value is not set in the configuration, it returns nil.
func getOptionalBool(d *schema.ResourceData, key string) *graphql.Boolean {
	// Check if the raw config has the key set explicitly
	rawConfig := d.GetRawConfig()

	if rawConfig.IsNull() {
		return nil
	}

	configValue := rawConfig.GetAttr(key)
	if configValue.IsNull() {
		return nil
	}

	// User explicitly set the value in config
	return toOptionalBool(d.Get(key))
}

// convertToString converts various types to their string representation
// This is used for plugin parameters where the API expects strings but users provide typed values
func convertToString(input interface{}) string {
	if input == nil {
		return ""
	}

	switch v := input.(type) {
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", v)
	default:
		// Fallback to fmt.Sprintf for any other type
		return fmt.Sprintf("%v", v)
	}
}
