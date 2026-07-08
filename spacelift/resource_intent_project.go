package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceIntentProject() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_intent_project` represents a Spacelift **Intent project** - " +
			"a workspace for managing cloud resources through Spacelift Intent. An " +
			"Intent project can be given a time-to-live (TTL) after which its " +
			"resources are automatically cleaned up.",

		CreateContext: resourceIntentProjectCreate,
		ReadContext:   resourceIntentProjectRead,
		UpdateContext: resourceIntentProjectUpdate,
		DeleteContext: resourceIntentProjectDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the intent project",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"space_id": {
				Type:             schema.TypeString,
				Description:      "ID (slug) of the space the intent project is in",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Free-form intent project description for users",
				Optional:    true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "Labels of the intent project, used for policy autoattachment and filtering",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			"worker_pool_id": {
				Type:        schema.TypeString,
				Description: "ID of the worker pool assigned to this intent project for task execution",
				Optional:    true,
			},
			"runner_image": {
				Type:        schema.TypeString,
				Description: "Optional Docker image used to run intent tasks for this project. Leave empty to use the system default intent worker image.",
				Optional:    true,
			},
			"ttl": {
				Type:          schema.TypeString,
				Description:   "Time-to-live for the intent project expressed as a duration (e.g. `72h`, `7d`, `30d`). When the TTL elapses, the project's resources are cleaned up according to `on_expiry_action`. Mutually exclusive with `expires_at`.",
				Optional:      true,
				ConflictsWith: []string{"expires_at"},
			},
			"expires_at": {
				Type:          schema.TypeInt,
				Description:   "Absolute Unix timestamp at which the intent project's TTL elapses and resource cleanup kicks in. Mutually exclusive with `ttl`.",
				Optional:      true,
				ConflictsWith: []string{"ttl"},
			},
			"on_expiry_action": {
				Type:        schema.TypeString,
				Description: "What happens to the project after TTL expiry cleanup. Possible values are `ARCHIVE` (keep the project in a read-only, archived state) and `DELETE` (delete the project entirely). Defaults to `ARCHIVE`.",
				Optional:    true,
				Default:     "ARCHIVE",
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice([]string{"ARCHIVE", "DELETE"}, false),
				),
			},
			"keep_resources_on_destroy": {
				Type:        schema.TypeBool,
				Description: "Keep the intent project's managed cloud resources when the project is destroyed. Defaults to `false`.",
				Optional:    true,
				Default:     false,
			},
			"ttl_seconds": {
				Type:        schema.TypeInt,
				Description: "The originally configured TTL duration in seconds. Zero when the TTL was set as an absolute timestamp (or not at all).",
				Computed:    true,
			},
			"archived_at": {
				Type:        schema.TypeInt,
				Description: "Unix timestamp when the project was archived by TTL expiry cleanup.",
				Computed:    true,
			},
			"state": {
				Type:        schema.TypeString,
				Description: "The state of the intent project.",
				Computed:    true,
			},
		},
	}
}

func resourceIntentProjectCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var mutation struct {
		CreateIntentProject structs.IntentProject `graphql:"intentProjectCreate(name: $name, description: $description, labels: $labels, space: $space, workerPool: $workerPool, runnerImage: $runnerImage, ttl: $ttl, expiresAt: $expiresAt, onExpiryAction: $onExpiryAction)"`
	}

	variables := map[string]any{
		"name":           toString(d.Get("name")),
		"space":          toID(d.Get("space_id")),
		"labels":         intentProjectLabels(d),
		"description":    (*graphql.String)(nil),
		"workerPool":     (*graphql.ID)(nil),
		"runnerImage":    (*graphql.String)(nil),
		"ttl":            (*graphql.String)(nil),
		"expiresAt":      (*graphql.Int)(nil),
		"onExpiryAction": structs.IntentProjectExpiryAction(d.Get("on_expiry_action").(string)),
	}

	if description, ok := d.GetOk("description"); ok {
		variables["description"] = toOptionalString(description)
	}

	if workerPoolID, ok := d.GetOk("worker_pool_id"); ok {
		variables["workerPool"] = graphql.NewID(workerPoolID)
	}

	if runnerImage, ok := d.GetOk("runner_image"); ok {
		variables["runnerImage"] = toOptionalString(runnerImage)
	}

	if ttl, ok := d.GetOk("ttl"); ok {
		variables["ttl"] = toOptionalString(ttl)
	}

	if expiresAt, ok := d.GetOk("expires_at"); ok {
		variables["expiresAt"] = toOptionalInt(expiresAt)
	}

	if err := meta.(*internal.Client).Mutate(ctx, "IntentProjectCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create intent project %v: %v", d.Get("name"), internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.CreateIntentProject.ID)

	return resourceIntentProjectRead(ctx, d, meta)
}

func resourceIntentProjectRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var query struct {
		IntentProject *structs.IntentProject `graphql:"intentProject(id: $id)"`
	}

	variables := map[string]any{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "IntentProjectRead", &query, variables); err != nil {
		return diag.Errorf("could not query for intent project: %v", err)
	}

	project := query.IntentProject
	if project == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", project.Name)
	d.Set("space_id", project.Space.ID)
	d.Set("runner_image", project.RunnerImage)
	d.Set("on_expiry_action", string(project.OnExpiryAction))
	d.Set("state", project.State)

	if project.Description != nil {
		d.Set("description", *project.Description)
	} else {
		d.Set("description", "")
	}

	labels := schema.NewSet(schema.HashString, []any{})
	for _, label := range project.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	if project.WorkerPool != nil {
		d.Set("worker_pool_id", project.WorkerPool.ID)
	} else {
		d.Set("worker_pool_id", "")
	}

	if project.TTLSeconds != nil {
		d.Set("ttl_seconds", *project.TTLSeconds)
	} else {
		d.Set("ttl_seconds", 0)
	}

	if project.ArchivedAt != nil {
		d.Set("archived_at", *project.ArchivedAt)
	} else {
		d.Set("archived_at", 0)
	}

	// expires_at is a user-managed input rather than a computed field. Only
	// refresh it when it is already tracked in state (i.e. explicitly
	// configured): reflecting a backend-derived value — a `ttl` duration or a
	// space default TTL — would create a phantom diff whose apply then clears
	// the TTL via the update path's clearTtl branch.
	if _, ok := d.GetOk("expires_at"); ok && project.ExpiresAt != nil {
		d.Set("expires_at", *project.ExpiresAt)
	}

	return nil
}

func resourceIntentProjectUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var mutation struct {
		UpdateIntentProject structs.IntentProject `graphql:"intentProjectUpdate(id: $id, name: $name, description: $description, labels: $labels, space: $space, workerPool: $workerPool, runnerImage: $runnerImage, ttl: $ttl, expiresAt: $expiresAt, clearTtl: $clearTtl, onExpiryAction: $onExpiryAction)"`
	}

	variables := map[string]any{
		"id":             toID(d.Id()),
		"name":           toString(d.Get("name")),
		"space":          graphql.NewID(d.Get("space_id")),
		"labels":         intentProjectLabels(d),
		"description":    (*graphql.String)(nil),
		"workerPool":     (*graphql.ID)(nil),
		"runnerImage":    (*graphql.String)(nil),
		"ttl":            (*graphql.String)(nil),
		"expiresAt":      (*graphql.Int)(nil),
		"clearTtl":       (*graphql.Boolean)(nil),
		"onExpiryAction": structs.IntentProjectExpiryAction(d.Get("on_expiry_action").(string)),
	}

	if description, ok := d.GetOk("description"); ok {
		variables["description"] = toOptionalString(description)
	}

	if workerPoolID, ok := d.GetOk("worker_pool_id"); ok {
		variables["workerPool"] = graphql.NewID(workerPoolID)
	}

	if runnerImage, ok := d.GetOk("runner_image"); ok {
		variables["runnerImage"] = toOptionalString(runnerImage)
	}

	// Only touch the TTL when its inputs actually changed: re-sending the
	// duration on unrelated updates would re-derive expires_at from "now" and
	// silently extend the project's lifetime on every apply.
	if d.HasChange("ttl") || d.HasChange("expires_at") {
		ttl, ttlSet := d.GetOk("ttl")
		expiresAt, expiresAtSet := d.GetOk("expires_at")

		switch {
		case ttlSet:
			variables["ttl"] = toOptionalString(ttl)
		case expiresAtSet:
			variables["expiresAt"] = toOptionalInt(expiresAt)
		default:
			// Neither TTL input is present anymore. If the project previously had a
			// TTL, explicitly clear it on the backend.
			oldTTL, _ := d.GetChange("ttl")
			oldExpiresAt, _ := d.GetChange("expires_at")
			if oldTTL.(string) != "" || oldExpiresAt.(int) != 0 {
				clearTTL := graphql.Boolean(true)
				variables["clearTtl"] = &clearTTL
			}
		}
	}

	var ret diag.Diagnostics

	if err := meta.(*internal.Client).Mutate(ctx, "IntentProjectUpdate", &mutation, variables); err != nil {
		ret = diag.Errorf("could not update intent project: %v", internal.FromSpaceliftError(err))
	}

	return append(ret, resourceIntentProjectRead(ctx, d, meta)...)
}

func resourceIntentProjectDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var mutation struct {
		DeleteIntentProject *structs.IntentProject `graphql:"intentProjectDelete(id: $id, keepResources: $keepResources)"`
	}

	variables := map[string]any{
		"id":            toID(d.Id()),
		"keepResources": graphql.Boolean(d.Get("keep_resources_on_destroy").(bool)),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "IntentProjectDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete intent project: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func intentProjectLabels(d *schema.ResourceData) []graphql.String {
	labels := []graphql.String{}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}
	}

	return labels
}
