package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceRun() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_run` allows programmatically triggering runs in response " +
			"to arbitrary changes in the keepers section.",

		CreateContext: resourceRunCreate,
		ReadContext:   schema.NoopContext,
		Delete:        schema.RemoveFromState,

		Schema: map[string]*schema.Schema{
			"stack_id": {
				Type:             schema.TypeString,
				Description:      "ID of the stack on which the run is to be triggered.",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"commit_sha": {
				Description: "The commit SHA for which to trigger a run.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"keepers": {
				Description: "" +
					"Arbitrary map of values that, when changed, will trigger " +
					"recreation of the resource.",
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
			"proposed": {
				Type:        schema.TypeBool,
				Description: "Whether the run is a proposed run. Defaults to `false`.",
				Optional:    true,
				ForceNew:    true,
				Default:     false,
			},
			"id": {
				Description: "The ID of the triggered run.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceRunCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		ID string `graphql:"runResourceCreate(stack: $stack, commitSha: $sha, proposed: $proposed)"`
	}

	stackID := d.Get("stack_id")

	variables := map[string]interface{}{
		"stack":    toID(stackID),
		"sha":      (*graphql.String)(nil),
		"proposed": (*graphql.Boolean)(nil),
	}

	if sha, ok := d.GetOk("commit_sha"); ok {
		variables["sha"] = graphql.NewString(graphql.String(sha.(string)))
	}

	if proposed, ok := d.GetOk("proposed"); ok {
		variables["proposed"] = graphql.NewBoolean(graphql.Boolean(proposed.(bool)))
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ResourceRunCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not trigger run for stack %s: %v", stackID, internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.ID)

	return nil
}
