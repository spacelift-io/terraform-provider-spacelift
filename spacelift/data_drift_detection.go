package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataDriftDetection() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_drift_detection` represents a Drift Detection configuration " +
			"for a Stack. It will trigger a proposed run on the given schedule, which you can " +
			"listen for using run state webhooks. If reconcile is true, then a tracked run " +
			"will be triggered when drift is detected.",

		ReadContext: dataDriftDetectionRead,

		Schema: map[string]*schema.Schema{
			"reconcile": {
				Type:        schema.TypeBool,
				Description: "Whether a tracked run should be triggered when drift is detected.",
				Computed:    true,
			},
			"ignore_state": {
				Type:        schema.TypeBool,
				Description: "Controls whether drift detection should be performed on a stack in any final state instead of just 'Finished'.",
				Optional:    true,
			},
			"stack_id": {
				Type:             schema.TypeString,
				Description:      "ID of the stack for which to set up drift detection",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"schedule": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of cron schedule expressions based on which drift detection should be triggered.",
				Computed:    true,
			},
			"timezone": {
				Type:        schema.TypeString,
				Description: "Timezone in which the schedule is expressed",
				Computed:    true,
			},
		},
	}
}

func dataDriftDetectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var ret diag.Diagnostics

	stackID := d.Get("stack_id")
	d.SetId(stackID.(string))
	ret = resourceStackDriftDetectionReadWithHooks(ctx, d, meta, func(message string) diag.Diagnostics {
		return diag.Errorf("%s", message)
	})

	if ret.HasError() {
		d.SetId("")
	}

	return ret
}
