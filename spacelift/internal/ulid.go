package internal

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func WithULID(objectName string, r *schema.Resource) *schema.Resource {
	r.Schema["ulid"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: fmt.Sprintf("Globally unique identifier (ULID) of the %s", objectName),
		Computed:    true,
	}

	return r
}

func SetULID(d *schema.ResourceData, ulid string) error {
	return errors.Wrap(d.Set("ulid", ulid), "could not set ulid")
}
