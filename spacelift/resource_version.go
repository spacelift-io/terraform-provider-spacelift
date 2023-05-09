package spacelift

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceVersion() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_version` allows to programmatically trigger a version creation " +
			"in response to arbitrary changes in the keepers section.",

		CreateContext: resourceVersionCreate,
		ReadContext:   schema.NoopContext,
		Delete:        schema.RemoveFromState,

		Schema: map[string]*schema.Schema{
			"module_id": {
				Type:             schema.TypeString,
				Description:      "ID of the module on which the version creation is to be triggered.",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"commit_sha": {
				Description: "The commit SHA for which to trigger a version.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"version_number": {
				Description: "A schematic version number to set for the triggered version, example: 0.11.2",
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
			"id": {
				Description: "The ID of the triggered version.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		Version struct {
			ID string
		} `graphql:"versionCreate(module: $module, commitSha: $sha, version: $version)"`
	}

	moduleID := d.Get("module_id")

	variables := map[string]interface{}{
		"module":  toID(moduleID),
		"sha":     (*graphql.String)(nil),
		"version": (*graphql.String)(nil),
	}

	if sha, ok := d.GetOk("commit_sha"); ok {
		variables["sha"] = toString(sha)
	}

	if version, ok := d.GetOk("version_number"); ok {
		variables["version"] = toString(version)
	}

	if err := meta.(*internal.Client).Mutate(ctx, "ResourceVersionCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not trigger version for module %s: %v", moduleID, internal.FromSpaceliftError(err))
	}

	if mutation.Version.ID != "" {
		diag := waitForVersionCreate(ctx, meta.(*internal.Client), mutation.Version.ID, moduleID.(string))
		if diag.HasError() {
			return diag
		}
	}

	d.SetId(mutation.Version.ID)

	return nil
}

func waitForVersionCreate(ctx context.Context, client *internal.Client, versionID, moduleID string) diag.Diagnostics {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	variables := map[string]interface{}{
		"moduleId":  graphql.ID(moduleID),
		"versionId": graphql.ID(versionID),
	}

	for {
		select {
		case <-ctx.Done():
			return diag.FromErr(ctx.Err())
		case <-ticker.C:
		}

		var query struct {
			Module struct {
				Version struct {
					State string `graphql:"state"`
				} `graphql:"version(id: $versionId)"`
			} `graphql:"module(id: $moduleId)"`
		}

		if err := client.Query(ctx, "GetVersion", &query, variables); err != nil {
			return diag.Errorf("could not query for module %q with version %q: %v", moduleID, versionID, err)
		}

		switch query.Module.Version.State {
		case "ACTIVE":
			return nil
		case "FAILED":
			return diag.Errorf("module %q version %q creation failed, please check the run logs for more information", moduleID, versionID)
		default:
			// We wait if module version is in any other state.
			continue
		}
	}
}
