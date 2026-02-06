package validations

import (
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DisallowEmptyString ensures that the given value is not an empty string.
func DisallowEmptyString(in interface{}, path cty.Path) diag.Diagnostics {
	if in == "" {
		return diag.Errorf("%s must not be an empty string", path)
	}

	return nil
}

func ValidateWriteOnlyField(value, valueWo, valueWoVersion string, d *schema.ResourceData) (string, diag.Diagnostics){
	var result string

	if v, ok := d.GetOk(value); ok {
		result = v.(string)
	}

	if _, ok := d.GetOk(valueWoVersion); ok {
		// To get the value of a write-only attribute, we need to access the raw config.
		p := cty.GetAttrPath(valueWo)
		woVal, diags := d.GetRawConfigAt(p)
		if diags.HasError() {
			return "", diag.FromErr(fmt.Errorf("could not get write-only value %s: %v", valueWo, diags))
		}

		if !woVal.IsNull() {
			result = woVal.AsString()
		}
	}

	return result, nil
}
