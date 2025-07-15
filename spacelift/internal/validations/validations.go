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
