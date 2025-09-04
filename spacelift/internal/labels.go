package internal

import (
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func FilterByRequiredLabels[S any](d *schema.ResourceData, ss []S, f func(S) []string) []S {
	labelsRaw, labelsSpecified := d.GetOk("labels")
	requestedLabels := labelsRaw.(*schema.Set).List()

	matchLabels := func(s S) bool {
		if !labelsSpecified {
			return true
		}
		for _, label := range requestedLabels {
			if !slices.Contains(f(s), label.(string)) {
				return false
			}
		}
		return true
	}

	res := make([]S, 0, len(ss))
	for _, s := range ss {
		if !matchLabels(s) {
			continue
		}
		res = append(res, s)
	}
	return res
}
