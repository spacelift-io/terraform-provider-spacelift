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
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func dataCurrentSpace() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_current_space` is a data source that provides information " +
			"about the space that an administrative stack is in if the run is executed within " +
			"Spacelift by a stack or module. This  makes it easier to create resources " +
			"within the same space.",
		ReadContext: dataCurrentSpaceRead,

		Schema: map[string]*schema.Schema{
			"parent_space_id": {
				Type:        schema.TypeString,
				Description: "immutable ID (slug) of parent space",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "free-form space description for users",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "name of the space",
				Computed:    true,
			},
			"inherit_entities": {
				Type:        schema.TypeBool,
				Description: "indication whether access to this space inherits read access to entities from the parent space",
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "list of labels describing a space",
				Computed:    true,
			},
		},
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
		Stack  *structs.Stack  `graphql:"stack(id: $id)"`
		Module *structs.Module `graphql:"module(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(strings.TrimRight(stackID, "/"))}
	if err := meta.(*internal.Client).Query(ctx, "StackRead", &query, variables); err != nil {
		if strings.Contains(err.Error(), "denied") {
			return diag.Errorf("could not query for stack: %v, is this stack administrative?", err)
		}
		return diag.Errorf("could not query for stack: %v", err)
	}

	var space structs.Space

	switch {
	case query.Stack != nil:
		space = query.Stack.SpaceDetails
	case query.Module != nil:
		space = query.Module.SpaceDetails
	default:
		return diag.Errorf("could not find stack or module with ID %s", stackID)
	}

	d.SetId(space.ID)
	d.Set("name", space.Name)
	d.Set("description", space.Description)
	d.Set("inherit_entities", space.InheritEntities)

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range space.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	if space.ParentSpace != nil {
		d.Set("parent_space_id", *space.ParentSpace)
	}

	return nil
}
