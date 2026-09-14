package internal

import (
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ExtractWriteOnlyField(value, valueWo, valueWoVersion string, data *schema.ResourceData) (string, diag.Diagnostics) {
	return extractWriteOnlyField(
		value, valueWo, valueWoVersion,
		func(field string) string { return field },
		cty.GetAttrPath,
		data,
	)
}

// ExtractWriteOnlyFieldInBlock is the variant for fields inside a
// `MaxItems: 1` block. Write-only values are only readable from the raw
// config, addressed by cty path, so the block name cannot simply be prefixed
// onto the flat attribute key.
func ExtractWriteOnlyFieldInBlock(block, value, valueWo, valueWoVersion string, data *schema.ResourceData) (string, diag.Diagnostics) {
	return extractWriteOnlyField(
		value, valueWo, valueWoVersion,
		func(field string) string { return fmt.Sprintf("%s.0.%s", block, field) },
		func(field string) cty.Path { return cty.GetAttrPath(block).IndexInt(0).GetAttr(field) },
		data,
	)
}

func extractWriteOnlyField(
	value, valueWo, valueWoVersion string,
	key func(string) string,
	path func(string) cty.Path,
	data *schema.ResourceData,
) (string, diag.Diagnostics) {
	var result string

	if v, ok := data.GetOk(key(value)); ok {
		result = v.(string)
	}

	if _, ok := data.GetOk(key(valueWoVersion)); ok {
		// To get the value of a write-only attribute, we need to access the raw config.
		woVal, diags := data.GetRawConfigAt(path(valueWo))
		if diags.HasError() {
			return "", diag.FromErr(fmt.Errorf("could not get write-only value %s: %v", key(valueWo), diags))
		}

		if !woVal.IsNull() {
			result = woVal.AsString()
		}
	}

	return result, nil
}
