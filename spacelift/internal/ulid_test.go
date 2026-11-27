package internal

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestWithULID(t *testing.T) {
	r := WithULID("widget", &schema.Resource{
		Schema:   map[string]*schema.Schema{"name": {Type: schema.TypeString, Optional: true}},
		Importer: &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
	})

	if err := r.InternalValidate(nil, true); err != nil {
		t.Fatalf("invalid resource: %v", err)
	}

	byIdentity := r.Data(nil)
	if _, err := r.Importer.StateContext(t.Context(), byIdentity, nil); err == nil || !strings.Contains(err.Error(), "import it by ID instead") {
		t.Errorf("import by identity: got error %v", err)
	}

	byID := r.Data(nil)
	byID.SetId("my-widget")
	if _, err := r.Importer.StateContext(t.Context(), byID, nil); err != nil {
		t.Errorf("import by ID: %v", err)
	}
}
