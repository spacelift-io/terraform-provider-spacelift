package internal

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func writeOnlyTestResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"value":            {Type: schema.TypeString, Optional: true},
			"value_wo":         {Type: schema.TypeString, Optional: true, WriteOnly: true},
			"value_wo_version": {Type: schema.TypeString, Optional: true},
		},
	}
}

func TestExtractWriteOnlyField(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		attributes map[string]string
		rawConfig  cty.Value
		want       string
	}{
		{
			name:       "write-only value is read from raw config when version is set",
			attributes: map[string]string{"value_wo_version": "1"},
			rawConfig: cty.ObjectVal(map[string]cty.Value{
				"value":            cty.NullVal(cty.String),
				"value_wo":         cty.StringVal("secret"),
				"value_wo_version": cty.StringVal("1"),
			}),
			want: "secret",
		},
		{
			name:       "write-only value is ignored when version is not set",
			attributes: map[string]string{},
			rawConfig: cty.ObjectVal(map[string]cty.Value{
				"value":            cty.NullVal(cty.String),
				"value_wo":         cty.StringVal("secret"),
				"value_wo_version": cty.NullVal(cty.String),
			}),
			want: "",
		},
		{
			name:       "deprecated value is used when present",
			attributes: map[string]string{"value": "plain"},
			rawConfig: cty.ObjectVal(map[string]cty.Value{
				"value":            cty.StringVal("plain"),
				"value_wo":         cty.NullVal(cty.String),
				"value_wo_version": cty.NullVal(cty.String),
			}),
			want: "plain",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d := writeOnlyTestResource().Data(&terraform.InstanceState{
				Attributes: tc.attributes,
				RawConfig:  tc.rawConfig,
			})

			got, diags := ExtractWriteOnlyField("value", "value_wo", "value_wo_version", d)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
