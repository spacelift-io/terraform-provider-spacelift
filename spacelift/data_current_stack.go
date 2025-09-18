package spacelift

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

func dataCurrentStack() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_current_stack` is a data source that provides information " +
			"about the current administrative stack if the run is executed within " +
			"Spacelift by a stack or module. This allows clever tricks like " +
			"attaching contexts or policies to the stack that manages them.",
		ReadContext: dataCurrentStackRead,
	}
}

func dataCurrentStackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	stackID, err := getStackIDFromToken(meta.(*internal.Client).Token)
	if err != nil {
		return diag.Errorf("%v", err)
	}

	d.SetId(strings.TrimRight(stackID, "/"))

	return nil
}
