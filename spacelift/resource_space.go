package spacelift

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceSpace() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_space` represents a Spacelift **space** - " +
			"a collection of resources such as stacks, modules, policies, etc. Allows for more granular access control. Can have a parent space.\n\n" +
			"**Please note:** the `spacelift_space` resource requires root Admin permissions, and can only be used by administrative stacks " +
			"in the root space, or using an API key or user session that has root space access.",

		CreateContext: resourceSpaceCreate,
		ReadContext:   resourceSpaceRead,
		UpdateContext: resourceSpaceUpdate,
		DeleteContext: resourceSpaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"parent_space_id": {
				Type:        schema.TypeString,
				Description: "immutable ID (slug) of parent space. Defaults to `root`.",
				Optional:    true,
				Default:     "root",
			},
			"description": {
				Type:        schema.TypeString,
				Description: "free-form space description for users",
				Optional:    true,
			},
			"name": {
				Type:             schema.TypeString,
				Description:      "name of the space",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"inherit_entities": {
				Type:        schema.TypeBool,
				Description: "indication whether access to this space inherits read access to entities from the parent space. Defaults to `false`.",
				Optional:    true,
				Default:     false,
			},
			"labels": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "list of labels describing a space",
				Optional:    true,
			},
		},
	}
}

func spaceCreateInput(d *schema.ResourceData) structs.SpaceInput {
	input := structs.SpaceInput{
		Name:            toString(d.Get("name")),
		InheritEntities: graphql.Boolean(d.Get("inherit_entities").(bool)),
		ParentSpace:     toID(""),
	}

	parentSpace, ok := d.GetOk("parent_space_id")
	if ok {
		input.ParentSpace = toID(parentSpace)
	}

	description, ok := d.GetOk("description")
	if ok {
		input.Description = toString(description)
	}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String
		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}
		input.Labels = &labels
	}

	return input
}

func resourceSpaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreateSpace structs.Space `graphql:"spaceCreate(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": spaceCreateInput(d),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "CreateSpace", &mutation, variables); err != nil {
		return diag.Errorf("could not create space %v: %v", toString(d.Get("name")), spaceManagementError(err))
	}

	d.SetId(mutation.CreateSpace.ID)

	return resourceSpaceRead(ctx, d, meta)
}

func resourceSpaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Space *structs.Space `graphql:"space(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "SpaceRead", &query, variables); err != nil {
		return diag.Errorf("could not query for space: %v", err)
	}

	space := query.Space
	if space == nil {
		return diag.Errorf("could not find space %s", d.Id())
	}

	d.SetId(space.ID)
	d.Set("name", space.Name)
	d.Set("description", space.Description)
	d.Set("inherit_entities", space.InheritEntities)
	if space.ParentSpace != nil {
		d.Set("parent_space_id", *space.ParentSpace)
	}
	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range space.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	return nil
}

func resourceSpaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UpdateSpace structs.Space `graphql:"spaceUpdate(space: $space, input: $input)"`
	}

	variables := map[string]interface{}{
		"space": toID(d.Id()),
		"input": spaceCreateInput(d),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "SpaceUpdate", &mutation, variables); err != nil {
		return diag.Errorf("could not update space: %v", spaceManagementError(err))
	}

	return resourceSpaceRead(ctx, d, meta)
}

func resourceSpaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteSpace *structs.Space `graphql:"spaceDelete(space: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "DeleteSpace", &mutation, variables); err != nil {
		return diag.Errorf("could not delete space: %v", spaceManagementError(err))
	}

	d.SetId("")

	return nil
}

func spaceManagementError(err error) error {
	if err == nil || !strings.Contains(err.Error(), "unauthorized") {
		return err
	}

	return fmt.Errorf("%w - is it an administrative stack in the root space?", err)
}
