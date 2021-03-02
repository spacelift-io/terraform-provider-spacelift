package spacelift

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func importIntegration(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	ID := d.Id()

	parts := strings.Split(ID, "/")

	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ID: expected [stack|module]/$id, got %q", ID)
	}

	switch resourceType, resourceID := parts[0], parts[1]; resourceType {
	case "module":
		d.SetId(resourceID)
		d.Set("module_id", resourceID)
	case "stack":
		d.SetId(resourceID)
		d.Set("stack_id", resourceID)
	default:
		return nil, fmt.Errorf("invalid resource type %q, only module and stack are supported", resourceType)
	}

	return []*schema.ResourceData{d}, nil
}
