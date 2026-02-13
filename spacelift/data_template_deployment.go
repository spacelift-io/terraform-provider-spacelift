package spacelift

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataTemplateDeployment() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_template_deployment` represents an existing deployment of a Spacelift template.",

		ReadContext: dataTemplateDeploymentRead,

		Schema: map[string]*schema.Schema{
			"template_id": {
				Type:             schema.TypeString,
				Description:      "ID of the template this deployment belongs to",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"deployment_id": {
				Type:             schema.TypeString,
				Description:      "ID (slug) of the deployment",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the deployment",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of the deployment",
				Computed:    true,
			},
			"state": {
				Type:        schema.TypeString,
				Description: "Current state of the deployment",
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeInt,
				Description: "Unix timestamp when the deployment was created",
				Computed:    true,
			},
			"space": {
				Type:        schema.TypeString,
				Description: "Space where the stacks created by this deployment are placed",
				Computed:    true,
			},
			"template_version_id": {
				Type:        schema.TypeString,
				Description: "ID of the template version used by this deployment",
				Computed:    true,
			},
			"input": {
				Type:        schema.TypeList,
				Description: "Input values for the template",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "Input identifier",
							Computed:    true,
						},
						"value": {
							Type:        schema.TypeString,
							Description: "Input value, for secrets this will be set to an empty string",
							Computed:    true,
							Sensitive:   true,
						},
						"secret": {
							Type:        schema.TypeBool,
							Description: "True if the input value is a secret",
							Computed:    true,
						},
						"checksum": {
							Type:        schema.TypeString,
							Description: "SHA-256 checksum of the input value",
							Computed:    true,
						},
					},
				},
			},
			"stacks": {
				Type:        schema.TypeList,
				Description: "List of stacks IDs created by the deployment",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "ID of the stack",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataTemplateDeploymentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	templateID := d.Get("template_id").(string)
	deploymentID := d.Get("deployment_id").(string)

	var query struct {
		TemplateDeployment *structs.TemplateDeployment `graphql:"blueprintDeployment(id: $id, blueprintID: $blueprintID)"`
	}

	variables := map[string]interface{}{
		"id":          toID(deploymentID),
		"blueprintID": toID(templateID),
	}

	if err := meta.(*internal.Client).Query(ctx, "BlueprintDeployment", &query, variables); err != nil {
		return diag.Errorf("could not query for template deployment: %v", err)
	}

	if query.TemplateDeployment == nil {
		return diag.Errorf("could not find template deployment %s/%s", templateID, deploymentID)
	}

	deployment := query.TemplateDeployment

	// Set the resource ID in the format template_id/deployment_id
	d.SetId(fmt.Sprintf("%s/%s", templateID, deploymentID))

	d.Set("template_id", deployment.Template.ID)
	d.Set("deployment_id", deployment.ID)
	d.Set("name", deployment.Name)
	d.Set("state", deployment.State)
	d.Set("created_at", deployment.CreatedAt)
	d.Set("space", deployment.Space.ID)
	d.Set("template_version_id", deployment.TemplateVersion.ID)

	if deployment.Description != nil {
		d.Set("description", *deployment.Description)
	} else {
		d.Set("description", "")
	}

	// Set input values
	inputList := make([]map[string]interface{}, len(deployment.Inputs))
	for i, input := range deployment.Inputs {
		in := map[string]interface{}{
			"id":       input.ID,
			"value":    input.Value,
			"secret":   input.Secret,
			"checksum": input.Checksum,
		}
		inputList[i] = in
	}
	if err := d.Set("input", inputList); err != nil {
		return diag.FromErr(err)
	}

	// Set stacks
	stacks := make([]map[string]string, len(deployment.Stacks))
	for i, stack := range deployment.Stacks {
		stacks[i] = map[string]string{
			"id": stack.ID,
		}
	}
	if err := d.Set("stacks", stacks); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
