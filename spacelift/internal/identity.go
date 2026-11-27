package internal

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func SetIdentity(d *schema.ResourceData, attribute, value string) error {
	identity, err := d.Identity()
	if err != nil {
		return errors.Wrap(err, "could not get resource identity")
	}

	return errors.Wrapf(identity.Set(attribute, value), "could not set identity attribute %q", attribute)
}
