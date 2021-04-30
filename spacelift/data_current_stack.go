package spacelift

import (
	"path"
	"strings"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

func dataCurrentStack() *schema.Resource {
	return &schema.Resource{Read: dataCurrentStackRead}
}

func dataCurrentStackRead(d *schema.ResourceData, meta interface{}) error {
	var claims jwt.StandardClaims

	_, _, err := (&jwt.Parser{}).ParseUnverified(meta.(*internal.Client).Token, &claims)
	if err != nil {
		// Don't care about validation errors, we don't actually validate those
		// tokens, we only parse them.
		var unverifiable *jwt.UnverfiableTokenError
		if !errors.As(err, &unverifiable) {
			return errors.Wrap(err, "could not parse client token")
		}
	}

	if issuer := claims.Issuer; issuer != "spacelift" {
		return errors.Errorf("unexpected token issuer %s, is this a Spacelift run?", issuer)
	}

	stackID, _ := path.Split(claims.Subject)

	d.SetId(strings.TrimRight(stackID, "/"))

	return nil
}
