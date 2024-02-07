package spacelift

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/search"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/search/predicates"
)

func dataStacks() *schema.Resource {
	stackSchema := dataStack().Schema

	stackSchema["stack_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: "ID (slug) of the stack",
		Computed:    true,
	}

	return &schema.Resource{
		Description: "" +
			"`spacelift_stacks` represents all the stacks in the Spacelift " +
			"account visible to the API user, matching predicates.",

		ReadContext: dataStacksRead,

		Schema: map[string]*schema.Schema{
			// Search predicates.
			"administrative": predicates.BooleanField("Require stacks to be administrative or not", 1),
			"branch":         predicates.StringField("Require stacks to be on one of the branches", 1),
			"commit":         predicates.StringField("Require stacks to be on one of the commits", 1),
			"labels":         predicates.StringField("Require stacks to have one of the labels", 0),
			"locked":         predicates.BooleanField("Require stacks to be locked", 1),
			"name":           predicates.StringField("Require stacks to have one of the names", 1),
			"project_root":   predicates.StringField("Require stacks to be in one of the project roots", 1),
			"repository":     predicates.StringField("Require stacks to be in one of the repositories", 1),
			"state":          predicates.StringField("Require stacks to have one of the states", 1),
			"vendor":         predicates.StringField("Require stacks to use one of the IaC vendors", 1),
			"worker_pool":    predicates.StringField("Require stacks to use one of the worker pools", 1),

			// Results.
			"stacks": {
				Type:        schema.TypeList,
				Description: "List of stacks matching the predicates",
				Elem:        &schema.Resource{Schema: stackSchema},
				Computed:    true,
			},
		},
	}
}

func dataStacksRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Build the conditions.
	var conditions []search.SearchQueryPredicate

	conditions = append(conditions, predicates.BuildBoolean(d, "administrative")...)
	conditions = append(conditions, predicates.BuildStringOrEnum(d, false, "branch")...)
	conditions = append(conditions, predicates.BuildStringOrEnum(d, false, "commit")...)
	conditions = append(conditions, predicates.BuildStringOrEnum(d, false, "labels", "label")...)
	conditions = append(conditions, predicates.BuildBoolean(d, "locked")...)
	conditions = append(conditions, predicates.BuildStringOrEnum(d, false, "name")...)
	conditions = append(conditions, predicates.BuildStringOrEnum(d, false, "project_root", "projectRoot")...)
	conditions = append(conditions, predicates.BuildStringOrEnum(d, false, "repository")...)
	conditions = append(conditions, predicates.BuildStringOrEnum(d, true, "state")...)
	conditions = append(conditions, predicates.BuildStringOrEnum(d, true, "vendor")...)
	conditions = append(conditions, predicates.BuildStringOrEnum(d, false, "worker_pool", "workerPool")...)

	var query struct {
		SearchStacksOutput struct {
			Edges []struct {
				Node structs.Stack `graphql:"node"`
			} `graphql:"edges"`
			PageInfo search.PageInfo `graphql:"pageInfo"`
		} `graphql:"searchStacks(input: $input)"`
	}

	input := search.SearchInput{
		First:      graphql.NewInt(50),
		Predicates: &conditions,
	}

	var stacks []interface{}

	for {
		variables := map[string]interface{}{"input": input}

		if err := meta.(*internal.Client).Query(ctx, "StacksPage", &query, variables); err != nil {
			return diag.Errorf("could not query for stacks: %v", err)
		}

		for _, edge := range query.SearchStacksOutput.Edges {
			node := edge.Node

			stack := map[string]interface{}{
				"administrative":                   node.Administrative,
				"after_apply":                      node.AfterApply,
				"after_destroy":                    node.AfterDestroy,
				"after_init":                       node.AfterInit,
				"after_perform":                    node.AfterPerform,
				"after_plan":                       node.AfterPlan,
				"autodeploy":                       node.Autodeploy,
				"autoretry":                        node.Autoretry,
				"aws_assume_role_policy_statement": node.Integrations.AWS.AssumeRolePolicyStatement,
				"before_apply":                     node.BeforeApply,
				"before_destroy":                   node.BeforeDestroy,
				"before_init":                      node.BeforeInit,
				"before_perform":                   node.BeforePerform,
				"before_plan":                      node.BeforePlan,
				"branch":                           node.Branch,
				"description":                      node.Description,
				"enable_local_preview":             node.LocalPreviewEnabled,
				"labels":                           node.Labels,
				"manage_state":                     node.ManagesStateFile,
				"name":                             node.Name,
				"project_root":                     node.ProjectRoot,
				"protect_from_deletion":            node.ProtectFromDeletion,
				"repository":                       node.Repository,
				"runner_image":                     node.RunnerImage,
				"space_id":                         node.Space,
				"stack_id":                         node.ID,
				"terraform_version":                node.TerraformVersion,
			}

			if workerPool := node.WorkerPool; workerPool != nil {
				stack["worker_pool_id"] = workerPool.ID
			} else {
				stack["worker_pool_id"] = nil
			}

			if vcsKey, vcsSettings := node.VCSSettings(); vcsKey != "" {
				vcsValue := []interface{}{vcsSettings}
				if vcsSettings == nil {
					vcsValue = nil
				}
				stack[vcsKey] = vcsValue
			}

			if iacKey, iacSettings := node.IaCSettings(); iacKey != "" {
				stack[iacKey] = []interface{}{iacSettings}
			} else { // this is a Terraform stack
				stack["terraform_version"] = node.VendorConfig.Terraform.Version
				stack["terraform_workspace"] = node.VendorConfig.Terraform.Workspace
				stack["terraform_smart_sanitization"] = node.VendorConfig.Terraform.UseSmartSanitization
			}

			stacks = append(stacks, stack)
		}

		if !query.SearchStacksOutput.PageInfo.HasNextPage {
			break
		}

		after := graphql.String(query.SearchStacksOutput.PageInfo.EndCursor)
		input.After = &after
	}

	d.SetId(fmt.Sprintf("stacks-%d", time.Now().UnixNano()))

	if err := d.Set("stacks", stacks); err != nil {
		return diag.Errorf("could not set stacks: %v", err)
	}

	return nil
}
