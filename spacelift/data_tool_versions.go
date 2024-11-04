package spacelift

import (
	"context"
	"slices"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

func dataToolVersions() *schema.Resource {
	return &schema.Resource{
		Description: "Lists supported versions for a given tool.",
		ReadContext: dataToolVersionsRead,
		Schema: map[string]*schema.Schema{
			"tool": {
				Type:             schema.TypeString,
				Description:      "The tool to get a list of supported versions for. This can be one of `KUBECTL`, `OPEN_TOFU`, `TERRAFORM_FOSS`, or `TERRAGRUNT`.",
				ValidateDiagFunc: dataToolVersionsValidateInput,
				Required:         true,
			},
			"versions": {
				Type:        schema.TypeList,
				Description: "Supported versions of the given tool.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
		},
	}
}

func dataToolVersionsValidateInput(i interface{}, p cty.Path) diag.Diagnostics {
	tool := i.(string)

	validTools := []string{
		"KUBECTL",
		"OPEN_TOFU",
		"TERRAFORM_FOSS",
		"TERRAGRUNT",
	}

	if !slices.Contains(validTools, tool) {
		return diag.Errorf("tool must be one of %v", validTools)
	}

	return nil
}

func dataToolVersionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tool := d.Get("tool").(string)
	switch tool {
	case "KUBECTL":
		var query struct {
			KubectlVersions []string `graphql:"kubectlVersions"`
		}
		if err := meta.(*internal.Client).Query(ctx, "ReadKubectlVersions", &query, nil); err != nil {
			d.SetId("")
			return diag.Errorf("could not query for tool: %v", err)
		}
		d.SetId("kubectl-versions")
		d.Set("versions", query.KubectlVersions)
	case "OPEN_TOFU":
		var query struct {
			OpenTofuVersions []string `graphql:"openTofuVersions"`
		}
		if err := meta.(*internal.Client).Query(ctx, "ReadOpenTofuVersions", &query, nil); err != nil {
			d.SetId("")
			return diag.Errorf("could not query for tool: %v", err)
		}
		d.SetId("open-tofu-versions")
		d.Set("versions", query.OpenTofuVersions)
	case "TERRAFORM_FOSS":
		var query struct {
			TerraformVersions []string `graphql:"terraformVersions"`
		}
		if err := meta.(*internal.Client).Query(ctx, "ReadTerraformVersions", &query, nil); err != nil {
			d.SetId("")
			return diag.Errorf("could not query for tool: %v", err)
		}
		d.SetId("terraform-foss-versions")
		d.Set("versions", query.TerraformVersions)
	case "TERRAGRUNT":
		var query struct {
			TerragruntVersions []string `graphql:"terragruntVersions"`
		}
		if err := meta.(*internal.Client).Query(ctx, "ReadTerragruntVersions", &query, nil); err != nil {
			d.SetId("")
			return diag.Errorf("could not query for tool: %v", err)
		}
		d.SetId("terragrunt-versions")
		d.Set("versions", query.TerragruntVersions)
	default:
		return diag.Errorf("unsupported tool: %s", tool)
	}

	return nil
}
