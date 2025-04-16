package spacelift

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceBlueprintVersion() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_blueprint_version` represents a Spacelift BlueprintWithGroup, which allows you " +
			"to easily create stacks using a templating engine." +
			"A version of a blueprint must be assigned to a group `spacelift_blueprint_versioned_group`" +
			"For Terraform users it's preferable to use `spacelift_stack` instead. " +
			"This resource is mostly useful for those who do not use Terraform " +
			"to create stacks.",

		CreateContext: resourceBlueprintVersionCreate,
		ReadContext:   resourceBlueprintVersionRead,
		UpdateContext: resourceBlueprintVersionUpdate,
		DeleteContext: resourceBlueprintVersionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"group": {
				Type:             schema.TypeString,
				Description:      "ID of the Group the Blueprint is in. Group cannot be changed after creation.",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
				ForceNew:         true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of the Blueprint",
				Optional:    true,
			},
			"state": {
				Type:             schema.TypeString,
				Description:      "State of the Blueprint. Value can be `DRAFT` or `PUBLISHED`.",
				Required:         true,
				ValidateDiagFunc: validateStateEnum,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "Labels of the Blueprint",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"template": {
				Type:        schema.TypeString,
				Description: "Body of the Blueprint. If `state` is set to `PUBLISHED`, this field is required.",
				Optional:    true,
				ForceNew:    true,
			},
			"version": {
				Type:             schema.TypeString,
				Description:      "Version of the Blueprint.",
				ValidateDiagFunc: validations.DisallowEmptyString,
				Required:         true,
				ForceNew:         true,
			},
		},
	}
}

func resourceBlueprintVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		BlueprintWithGroup structs.BlueprintWithGroup `graphql:"blueprintWithGroupCreate(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": BlueprintVersionCreateInput(d),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "blueprintWithGroupCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create BlueprintWithGroup: %v", err)
	}

	d.SetId(path.Join(mutation.BlueprintWithGroup.ID, mutation.BlueprintWithGroup.GroupDetails.ID))

	return resourceBlueprintVersionRead(ctx, d, meta)
}

func resourceBlueprintVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Group *struct {
			BlueprintWithGroup *structs.BlueprintWithGroup `graphql:"blueprint(id: $id)"`
		} `graphql:"blueprintVersionedGroup(id: $group)"`
	}

	id, group := versionedBlueprintIDAndGroup(d)
	variables := map[string]interface{}{
		"id":    graphql.ID(id),
		"group": graphql.ID(group),
	}

	if err := meta.(*internal.Client).Query(ctx, "blueprintWithGroupRead", &query, variables); err != nil {
		return diag.Errorf("could not query for BlueprintWithGroup: %v", err)
	}

	if query.Group == nil {
		d.SetId("")
		return nil
	}

	bp := query.Group.BlueprintWithGroup

	d.Set("group", bp.GroupDetails.ID)
	d.Set("description", bp.Description)
	d.Set("state", bp.State)
	d.Set("version", bp.Version)

	if bp.RawTemplate == nil {
		d.Set("template", "")
	} else {
		d.Set("template", *bp.RawTemplate)
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range bp.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	return nil
}

func resourceBlueprintVersionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		BlueprintWithGroup structs.BlueprintWithGroup `graphql:"blueprintWithGroupUpdate(id: $id, input: $input)"`
	}

	id, _ := versionedBlueprintIDAndGroup(d)
	variables := map[string]interface{}{
		"id":    graphql.ID(id),
		"input": BlueprintVersionUpdateInput(d),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "blueprintWithGroupUpdate", &mutation, variables); err != nil {
		return diag.Errorf("could not update BlueprintWithGroup: %v", err)
	}

	return resourceBlueprintVersionRead(ctx, d, meta)
}

func resourceBlueprintVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		BlueprintWithGroup struct {
			ID string `graphql:"id"`
		} `graphql:"blueprintWithGroupDelete(id: $id)"`
	}

	id, _ := versionedBlueprintIDAndGroup(d)
	variables := map[string]interface{}{
		"id": graphql.ID(id),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "blueprintWithGroupDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete BlueprintWithGroup: %v", err)
	}

	d.SetId("")

	return nil
}

func BlueprintVersionUpdateInput(d *schema.ResourceData) structs.BlueprintWithGroupUpdateInput {
	var input structs.BlueprintWithGroupUpdateInput

	input.State = graphql.String(d.Get("state").(string))
	input.Version = graphql.String(d.Get("version").(string))

	if description, ok := d.GetOk("description"); ok {
		input.Description = toOptionalString(description)
	}

	input.Labels = []graphql.String{}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		for _, label := range labelSet.List() {
			input.Labels = append(input.Labels, graphql.String(label.(string)))
		}
	}

	if template, ok := d.GetOk("template"); ok {
		input.Template = toOptionalString(template)
	}

	return input
}

func BlueprintVersionCreateInput(d *schema.ResourceData) structs.BlueprintWithGroupCreateInput {
	input := BlueprintVersionUpdateInput(d)

	return structs.BlueprintWithGroupCreateInput{
		BlueprintWithGroupUpdateInput: input,
		Group:                         graphql.ID(d.Get("group").(string)),
	}
}

func versionedBlueprintIDAndGroup(d *schema.ResourceData) (string, string) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		panic(fmt.Sprintf("Malformed ID: %q, this is a bug in the provider", d.Id()))
	}

	return parts[0], parts[1]
}
