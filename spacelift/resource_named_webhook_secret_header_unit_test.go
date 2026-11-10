package spacelift

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Regression test for https://github.com/spacelift-io/terraform-provider-spacelift/issues/880.
func TestNamedWebhookSecretHeaderValueReadsWriteOnly(t *testing.T) {
	t.Parallel()

	d := resourceNamedWebhookSecretHeader().Data(&terraform.InstanceState{
		Attributes: map[string]string{
			"webhook_id":       "webhook",
			"key":              "X-Token",
			"value_wo_version": "1",
		},
		RawConfig: cty.ObjectVal(map[string]cty.Value{
			"id":               cty.NullVal(cty.String),
			"webhook_id":       cty.StringVal("webhook"),
			"key":              cty.StringVal("X-Token"),
			"value":            cty.NullVal(cty.String),
			"value_wo":         cty.StringVal("thisisavalue"),
			"value_wo_version": cty.StringVal("1"),
		}),
	})

	got, diags := namedWebhookSecretHeaderValue(d)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if got != "thisisavalue" {
		t.Fatalf("got %q, want %q", got, "thisisavalue")
	}
}
