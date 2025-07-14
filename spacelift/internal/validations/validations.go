package validations

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// DisallowEmptyString ensures that the given value is not an empty string.
func DisallowEmptyString(in interface{}, path cty.Path) diag.Diagnostics {
	if in == "" {
		return diag.Errorf("%s must not be an empty string", path)
	}

	return nil
}

// ValidateAction ensures that the given value is a valid action enum value.
// func ValidateAction(in interface{}, path cty.Path) diag.Diagnostics {
// 	actionStr, ok := in.(string)
// 	if !ok {
// 		return diag.Errorf("%s must be a string", path)
// 	}

// 	if !slices.Contains(structs.ActionList, structs.Action(actionStr)) {
// 		return diag.Errorf("%s must be one of %v", path, structs.ActionList)
// 	}

// 	return nil
// }
