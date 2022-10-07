package spacelift

import (
	"context"
	"path"
	"strings"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

func dataCurrentSpace() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_current_space` is a data source that provides information " +
			"about the space that an administrative stack is in if the run is executed within " +
			"Spacelift by a stack or module. This  makes it easier to create resources " +
			"within the same space.",
		ReadContext: dataCurrentSpaceRead,
	}
}

func dataCurrentSpaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(strings.TrimRight(stackID, "/"))}
	if err := meta.(*internal.Client).Query(ctx, "StackRead", &query, variables); err != nil {
		return diag.Errorf("could not query for stack: %v", err)
	}

	stack := query.Stack
	if stack == nil {
		return diag.Errorf("stack not found")
	}

	d.SetId(stack.Space)
	return nil
}
