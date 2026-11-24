package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestWithULID(t *testing.T) {
	r := WithULID("widget", &schema.Resource{
		Schema:   map[string]*schema.Schema{"name": {Type: schema.TypeString, Optional: true}},
		Importer: &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
	})

	if err := r.InternalValidate(nil, true); err != nil {
		t.Fatalf("invalid resource: %v", err)
	}

	for name, state := range map[string]*terraform.InstanceState{
		"by ID":       {ID: "01ULID"},
		"by identity": {Identity: map[string]string{"ulid": "01ULID"}},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := r.Importer.StateContext(t.Context(), r.Data(state), nil)
			if err != nil {
				t.Fatal(err)
			}
			if id := got[0].Id(); id != "01ULID" {
				t.Errorf("got ID %q, want 01ULID", id)
			}
		})
	}

	t.Run("by empty identity", func(t *testing.T) {
		state := &terraform.InstanceState{Identity: map[string]string{"ulid": ""}}
		if _, err := r.Importer.StateContext(t.Context(), r.Data(state), nil); err == nil {
			t.Error("expected an error")
		}
	})
}
