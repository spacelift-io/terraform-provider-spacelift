package spacelift

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// removeFromState drops a resource that's gone on the Spacelift side from the
// state, and returns a warning so the removal doesn't happen silently.
func removeFromState(d *schema.ResourceData, resource string) diag.Diagnostics {
	id := d.Id()
	d.SetId("")

	return diag.Diagnostics{{
		Severity: diag.Warning,
		Summary:  fmt.Sprintf("%s %s not found or not accessible, removing from state", resource, id),
	}}
}
