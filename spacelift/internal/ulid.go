package internal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

// WithULID adds a computed ulid attribute and a {ulid} identity.
func WithULID(objectName string, r *schema.Resource) *schema.Resource {
	r.Schema["ulid"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: fmt.Sprintf("Globally unique identifier (ULID) of the %s", objectName),
		Computed:    true,
	}

	r.Identity = &schema.ResourceIdentity{
		Version: 0,
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"ulid": {Type: schema.TypeString, RequiredForImport: true},
			}
		},
	}

	if r.Importer != nil {
		importByID := r.Importer.StateContext
		r.Importer.StateContext = func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
			if d.Id() == "" {
				return nil, errors.Errorf("importing a %s by identity is not supported, import it by ID instead", objectName)
			}
			return importByID(ctx, d, meta)
		}
	}

	return r
}

func SetULID(d *schema.ResourceData, ulid string) error {
	if err := d.Set("ulid", ulid); err != nil {
		return errors.Wrap(err, "could not set ulid")
	}

	return SetIdentity(d, "ulid", ulid)
}
