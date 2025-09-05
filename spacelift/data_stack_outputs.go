package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataStackOutputs() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_stack_outputs` represents the outputs of a Spacelift stack. " +
			"This data source can be used to retrieve output metadata " +
			"from other stacks for dynamic reference creation.",

		ReadContext: dataStackOutputsRead,

		Schema: map[string]*schema.Schema{
			"stack_id": {
				Type:             schema.TypeString,
				Description:      "ID (slug) of the stack",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"outputs": {
				Type:        schema.TypeSet,
				Description: "Map of stack outputs with their metadata",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "ID (name) of the output",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "Brief explanation of output's purpose or value",
							Computed:    true,
						},
						"sensitive": {
							Type:        schema.TypeBool,
							Description: "Indicates whether the output is sensitive",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataStackOutputsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Stack *struct {
			Outputs []structs.StackOutput `graphql:"outputs"`
		} `graphql:"stack(id: $id)"`
	}

	stackID := d.Get("stack_id")
	variables := map[string]interface{}{"id": toID(stackID)}
	if err := meta.(*internal.Client).Query(ctx, "StackOutputsRead", &query, variables); err != nil {
		return diag.Errorf("could not query for stack outputs: %v", err)
	}

	stack := query.Stack
	if stack == nil {
		return diag.Errorf("stack not found")
	}

	d.SetId(stackID.(string))

	outputs := make([]any, 0, len(stack.Outputs))
	for _, output := range stack.Outputs {
		outputs = append(outputs, map[string]any{
			"id":          output.ID,
			"description": output.Description,
			"sensitive":   output.Sensitive,
		})
	}
	d.Set("outputs", outputs)

	return nil
}
