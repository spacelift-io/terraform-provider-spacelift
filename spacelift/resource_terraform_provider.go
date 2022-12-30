package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceTerraformProvider() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_terraform_provider` represents a Terraform provider in " +
			"Spacelift's own provider registry.",
		Schema: map[string]*schema.Schema{
			"type": {
				Type:             schema.TypeString,
				Description:      "Type of the provider - should be unique in one account",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"space_id": {
				Type:        schema.TypeString,
				Description: "ID (slug) of the space the provider is in",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Free-form description for human users, supports Markdown",
				Optional:    true,
			},
			"labels": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"public": {
				Type:        schema.TypeBool,
				Description: "Whether the provider is public or not, defaults to false (private)",
				Optional:    true,
				Default:     false,
			},
		},
		CreateContext: resourceTerraformProviderCreate,
		ReadContext:   resourceTerraformProviderRead,
		UpdateContext: resourceTerraformProviderUpdate,
		DeleteContext: resourceTerraformProviderDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceTerraformProviderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var createMutation struct {
		CreateTerraformProvider structs.TerraformProvider `graphql:"terraformProviderCreate(type: $type, space: $space, description: $description, labels: $labels)"`
	}

	variables := map[string]any{
		"type":        d.Get("type"),
		"space":       d.Get("space_id"),
		"description": (*graphql.String)(nil),
		"labels":      (*[]graphql.String)(nil),
	}

	if description, ok := d.GetOk("description"); ok {
		variables["description"] = toOptionalString(description)
	}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String

		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}

		variables["labels"] = &labels
	}

	if err := meta.(*internal.Client).Mutate(ctx, "TerraformProviderCreate", &createMutation, variables); err != nil {
		return diag.Errorf("could not create Terraform provider: %v", internal.FromSpaceliftError(err))
	}

	if public := d.Get("public").(bool); public {
		var setVisibilityMutation struct {
			TerraformProviderSetVisibility structs.TerraformProvider `graphql:"terraformProviderSetVisibility(id: $id, public: $public)"`
		}

		variables = map[string]any{
			"id":     toID(createMutation.CreateTerraformProvider.ID),
			"public": toBool(public),
		}

		if err := meta.(*internal.Client).Mutate(ctx, "TerraformProviderSetVisibility", &setVisibilityMutation, variables); err != nil {
			return diag.Errorf("could not set visibility for Terraform provider: %v", internal.FromSpaceliftError(err))
		}
	}

	d.SetId(createMutation.CreateTerraformProvider.ID)

	return resourceTerraformProviderRead(ctx, d, meta)
}

func resourceTerraformProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		TerraformProvider *structs.TerraformProvider `graphql:"terraformProvider(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": toID(d.Id()),
	}

	if err := meta.(*internal.Client).Query(ctx, "TerraformProviderRead", &query, variables); err != nil {
		return diag.Errorf("could not query for Terraform provider: %v", err)
	}

	if query.TerraformProvider == nil {
		d.SetId("")
		return nil
	}

	d.Set("description", query.TerraformProvider.Description)
	d.Set("labels", query.TerraformProvider.Labels)
	d.Set("public", query.TerraformProvider.Public)
	d.Set("space_id", query.TerraformProvider.Space)
	d.Set("type", query.TerraformProvider.ID)

	return nil
}

func resourceTerraformProviderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var updateMutation struct {
		TerraformProviderUpdate structs.TerraformProvider `graphql:"terraformProviderUpdate(id: $id, space: $space, description: $description, labels: $labels)"`
	}

	variables := map[string]any{
		"id":          toID(d.Id()),
		"space":       d.Get("space_id"),
		"description": (*graphql.String)(nil),
		"labels":      (*[]graphql.String)(nil),
	}

	if description, ok := d.GetOk("description"); ok {
		variables["description"] = toOptionalString(description)
	}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String

		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}

		variables["labels"] = &labels
	}

	var ret diag.Diagnostics

	if err := meta.(*internal.Client).Mutate(ctx, "TerraformProviderUpdate", &updateMutation, variables); err != nil {
		ret = append(ret, diag.Errorf("could not update Terraform provider: %v", internal.FromSpaceliftError(err))...)
	}

	if d.HasChange("public") {
		var setVisibilityMutation struct {
			TerraformProviderSetVisibility structs.TerraformProvider `graphql:"terraformProviderSetVisibility(id: $id, public: $public)"`
		}

		variables = map[string]any{
			"id":     toID(d.Id()),
			"public": toBool(d.Get("public")),
		}

		if err := meta.(*internal.Client).Mutate(ctx, "TerraformProviderSetVisibility", &setVisibilityMutation, variables); err != nil {
			ret = append(ret, diag.Errorf("could not set visibility for Terraform provider: %v", internal.FromSpaceliftError(err))...)
		}
	}

	ret = append(ret, resourceTerraformProviderRead(ctx, d, meta)...)

	return ret
}

func resourceTerraformProviderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		TerraformProviderDelete *structs.TerraformProvider `graphql:"terraformProviderDelete(id: $id)"`
	}

	variables := map[string]any{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "TerraformProviderDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete Terraform provider: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	if mutation.TerraformProviderDelete == nil {
		return diag.Diagnostics{{
			Severity: diag.Warning,
			Summary:  "Terraform provider not found",
		}}
	}

	return nil
}
