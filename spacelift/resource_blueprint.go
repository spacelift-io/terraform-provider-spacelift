package spacelift

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

//go:embed blueprint_schema.json
var jsonSchema string
var compiledSchema *jsonschema.Schema

func init() {
	sch, err := jsonschema.CompileString("schema.json", jsonSchema)
	if err != nil {
		panic(err)
	}
	compiledSchema = sch
}

func resourceBlueprint() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_blueprint` represents a Spacelift blueprint, which allows you " +
			"to easily create stacks using a templating engine. " +
			"For Terraform users it's preferable to use `spacelift_stack` instead. " +
			"This resource is mostly useful for those who do not use Terraform " +
			"to create stacks.",

		CreateContext: resourceBlueprintCreate,
		ReadContext:   resourceBlueprintRead,
		UpdateContext: resourceBlueprintUpdate,
		DeleteContext: resourceBlueprintDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the blueprint",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"space": {
				Type:             schema.TypeString,
				Description:      "ID of the space the blueprint is in",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of the blueprint",
				Optional:    true,
			},
			"state": {
				Type:             schema.TypeString,
				Description:      "State of the blueprint. Value can be `DRAFT` or `PUBLISHED`.",
				Required:         true,
				ValidateDiagFunc: validateStateEnum,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "Labels of the blueprint",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"template": {
				Type:             schema.TypeString,
				Description:      "Body of the blueprint. If `state` is set to `PUBLISHED`, this field is required.",
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateTemplateMap,
			},
		},
	}
}

func validateTemplateMap(in interface{}, path cty.Path) diag.Diagnostics {
	blueprint, ok := in.(string)
	if !ok {
		return diag.Errorf("template_map map must be a map")
	}

	var j map[string]interface{}
	err := json.Unmarshal([]byte(blueprint), &j)
	if err != nil {
		return diag.Errorf("Unknown error")
	}

	err = compiledSchema.Validate(j)
	if err == nil {
		return nil
	}

	validationErr, ok := err.(*jsonschema.ValidationError)
	if !ok {
		return diag.Errorf("Unknown error")
	}

	var ignoredFields map[string]any
	ignoredFields = getIgnoredFields([]string{}, j, map[string]any{})

	flatten := flattenCauses(validationErr, ignoredFields)

	messag := fmt.Sprintf("Validation failed for %d fields", len(flatten))
	for _, err := range flatten {
		messag += fmt.Sprintf("\n%s: %s", err.InstanceLocation, err.Message)
	}

	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Error,
			Summary:  messag,
		},
	}
}

func getIgnoredFields(propertyDepth []string, blueprint, ignoredFields map[string]interface{}) map[string]interface{} {
	for key, val := range blueprint {
		switch v := val.(type) {
		case map[string]interface{}:
			ignoredFields = getIgnoredFields(append(propertyDepth, key), v, ignoredFields)
		case []interface{}:
			for i, item := range v {
				switch it := item.(type) {
				case map[string]interface{}:
					ignoredFields = getIgnoredFields(append(propertyDepth, key, strconv.Itoa(i)), it, ignoredFields)
				default:
					ignoredFields = getIgnoredFields(append(propertyDepth, key, strconv.Itoa(i)), map[string]interface{}{"": it}, ignoredFields)
				}
			}
		case string:
			if strings.Contains(v, "{{") && strings.Contains(v, "}}") {
				instanceLocation := "/" + strings.Join(append(propertyDepth, key), "/")
				ignoredFields[instanceLocation] = nil
			}
		}
	}

	return ignoredFields
}

func flattenCauses(cause *jsonschema.ValidationError, ignoredFields map[string]any) []*jsonschema.ValidationError {
	errors := make([]*jsonschema.ValidationError, 0, len(cause.Causes))

	for _, c := range cause.Causes {
		if len(c.Causes) > 0 {
			errors = append(errors, flattenCauses(c, ignoredFields)...)
		} else {
			if c.Message == "expected string, but got null" && len(ignoredFields) == 0 {
				// This typically happens when the user doesn't provide a value. Example:
				// description: ${{ inputs.description }}
				// becomes
				// description:
				// We shouldn't get an error, the emptiness/gap is a valid value.
				continue
			}

			if _, ok := ignoredFields[c.InstanceLocation]; !ok {
				errors = append(errors, c)
			}
		}
	}

	return errors
}

func validateStateEnum(in interface{}, path cty.Path) diag.Diagnostics {
	if in != "DRAFT" && in != "PUBLISHED" {
		return diag.Errorf("%s must be either DRAFT or PUBLISHED", path)
	}

	return nil
}

func resourceBlueprintCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		Blueprint structs.Blueprint `graphql:"blueprintCreate(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": blueprintCreateInput(d),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "BlueprintCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create blueprint: %v", err)
	}

	d.SetId(mutation.Blueprint.ID)

	return resourceBlueprintRead(ctx, d, meta)
}

func resourceBlueprintRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Blueprint *structs.Blueprint `graphql:"blueprint(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": graphql.ID(d.Id()),
	}

	if err := meta.(*internal.Client).Query(ctx, "BlueprintRead", &query, variables); err != nil {
		return diag.Errorf("could not query for blueprint: %v", err)
	}

	if query.Blueprint == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", query.Blueprint.Name)
	d.Set("space", query.Blueprint.Space.ID)
	d.Set("description", query.Blueprint.Description)
	d.Set("state", query.Blueprint.State)

	if query.Blueprint.RawTemplate == nil {
		d.Set("template", "")
	} else {
		d.Set("template", *query.Blueprint.RawTemplate)
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range query.Blueprint.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	return nil
}

func resourceBlueprintUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		Blueprint structs.Blueprint `graphql:"blueprintUpdate(id: $id, input: $input)"`
	}

	variables := map[string]interface{}{
		"id":    graphql.ID(d.Id()),
		"input": blueprintCreateInput(d),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "BlueprintUpdate", &mutation, variables); err != nil {
		return diag.Errorf("could not update blueprint: %v", err)
	}

	return resourceBlueprintRead(ctx, d, meta)
}

func resourceBlueprintDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		Blueprint struct {
			ID string `graphql:"id"`
		} `graphql:"blueprintDelete(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": graphql.ID(d.Id()),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "BlueprintDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete blueprint: %v", err)
	}

	d.SetId("")

	return nil
}

func blueprintCreateInput(d *schema.ResourceData) structs.BlueprintCreateInput {
	var input structs.BlueprintCreateInput

	input.Space = graphql.ID(d.Get("space").(string))
	input.Name = graphql.String(d.Get("name").(string))
	input.State = graphql.String(d.Get("state").(string))

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
