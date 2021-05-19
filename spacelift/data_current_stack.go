package spacelift

import (
	"context"
	"path"
	"strings"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

func dataCurrentStack() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_current_stack` is a data source that provides information " +
			"about the current administrative stack if the run is executed within " +
			"Spacelift by a stack or module. This allows clever tricks like " +
			"attaching contexts or policies to the stack that manages them.",
		ReadContext: dataCurrentStackRead,
	}
}

func dataCurrentStackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var claims jwt.StandardClaims

	_, _, err := (&jwt.Parser{}).ParseUnverified(meta.(*internal.Client).Token, &claims)
	if err != nil {
		// Don't care about validation errors, we don't actually validate those
		// tokens, we only parse them.
		var unverifiable *jwt.UnverfiableTokenError
		if !errors.As(err, &unverifiable) {
			return diag.Errorf("could not parse client token: %v", err)
		}
	}

	if issuer := claims.Issuer; issuer != "spacelift" {
		return diag.Errorf("unexpected token issuer %s, is this a Spacelift run?", issuer)
	}

	stackID, _ := path.Split(claims.Subject)

	d.SetId(strings.TrimRight(stackID, "/"))

	return nil
}
