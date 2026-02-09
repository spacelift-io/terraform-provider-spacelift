package internal

import (
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ExtractWriteOnlyField(value, valueWo, valueWoVersion string, data *schema.ResourceData) (string, diag.Diagnostics) {
	var result string

	if v, ok := data.GetOk(value); ok {
		result = v.(string)
	}

	if _, ok := data.GetOk(valueWoVersion); ok {
		// To get the value of a write-only attribute, we need to access the raw config.
		p := cty.GetAttrPath(valueWo)
		woVal, diags := data.GetRawConfigAt(p)
		if diags.HasError() {
			return "", diag.FromErr(fmt.Errorf("could not get write-only value %s: %v", valueWo, diags))
		}

		if !woVal.IsNull() {
			result = woVal.AsString()
		}
	}

	return result, nil
}
