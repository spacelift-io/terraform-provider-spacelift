package framework

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

func importIntegration(_ context.Context, id string, d *AWSRoleResourceModel) error {
	parts := strings.Split(id, "/")

	if len(parts) != 2 {
		return fmt.Errorf("invalid ID: expected [stack|module]/$id, got %q", id)
	}

	switch resourceType, resourceID := parts[0], parts[1]; resourceType {
	case "module":
		d.ID = types.StringValue(resourceID)
		d.ModuleID = types.StringValue(resourceID)
	case "stack":
		d.ID = types.StringValue(resourceID)
		d.StackID = types.StringValue(resourceID)
	default:
		return fmt.Errorf("invalid resource type %q, only module and stack are supported", resourceType)
	}

	return nil
}
