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

func dataModules() *schema.Resource {
	moduleSchema := dataModule().Schema

	moduleSchema["module_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: "ID (slug) of the module",
		Computed:    true,
	}

	return &schema.Resource{
		Description: "" +
			"`spacelift_modules` represents all the modules in the Spacelift " +
			"account visible to the API user, matching predicates.",

		ReadContext: dataModulesRead,

		Schema: map[string]*schema.Schema{
			"administrative": predicates.BooleanField("Require modules to be administrative or not", 1),
			"branch":         predicates.StringField("Require modules to be on one of the branches", 1),
			"labels":         predicates.StringField("Require modules to have one of the labels", 0),
			"name":           predicates.StringField("Require modules to have one of the names", 1),
			"project_root":   predicates.StringField("Require modules to be in one of the project roots", 1),
			"repository":     predicates.StringField("Require modules to be in one of the repositories", 1),
			"worker_pool":    predicates.StringField("Require modules to use one of the worker pools", 1),
			"commit":         predicates.StringField("Require modules to be on one of the commits", 1),

			"modules": {
				Type:        schema.TypeList,
				Description: "List of modules matching the predicates",
				Elem:        &schema.Resource{Schema: moduleSchema},
				Computed:    true,
			},
		},
	}
}

func dataModulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var conditions []search.SearchQueryPredicate

	conditions = append(conditions, predicates.BuildBoolean(d, "administrative")...)
	conditions = append(conditions, predicates.BuildStringOrEnum(d, false, "branch")...)
	conditions = append(conditions, predicates.BuildStringOrEnum(d, false, "labels", "label")...)
	conditions = append(conditions, predicates.BuildStringOrEnum(d, false, "name")...)
	conditions = append(conditions, predicates.BuildStringOrEnum(d, false, "project_root", "projectRoot")...)
	conditions = append(conditions, predicates.BuildStringOrEnum(d, false, "repository")...)
	conditions = append(conditions, predicates.BuildStringOrEnum(d, false, "worker_pool", "workerPool")...)
	conditions = append(conditions, predicates.BuildStringOrEnum(d, false, "commit")...)

	var query struct {
		SearchModulesOutput struct {
			Edges []struct {
				Node structs.Module `graphql:"node"`
			} `graphql:"edges"`
			PageInfo search.PageInfo `graphql:"pageInfo"`
		} `graphql:"searchModules(input: $input)"`
	}

	input := search.SearchInput{
		First:      graphql.NewInt(50),
		Predicates: &conditions,
	}

	var modules []interface{}

	for {
		variables := map[string]interface{}{"input": input}

		if err := meta.(*internal.Client).Query(ctx, "ModulesPage", &query, variables); err != nil {
			return diag.Errorf("could not query for modules: %v", err)
		}

		for _, edge := range query.SearchModulesOutput.Edges {
			node := edge.Node

			module := map[string]interface{}{
				"administrative":                   node.Administrative,
				"aws_assume_role_policy_statement": node.Integrations.AWS.AssumeRolePolicyStatement,
				"branch":                           node.Branch,
				"description":                      node.Description,
				"enable_local_preview":             node.LocalPreviewEnabled,
				"labels":                           node.Labels,
				"module_id":                        node.ID,
				"name":                             node.Name,
				"project_root":                     node.ProjectRoot,
				"protect_from_deletion":            node.ProtectFromDeletion,
				"repository":                       node.Repository,
				"runner_image":                     node.RunnerImage,
				"space_id":                         node.Space,
				"terraform_provider":               node.TerraformProvider,
				"workflow_tool":                    node.WorkflowTool,
				"git_sparse_checkout_paths":        node.GitSparseCheckoutPaths,
			}

			module["worker_pool_id"] = nil
			if workerPool := node.WorkerPool; workerPool != nil {
				module["worker_pool_id"] = workerPool.ID
			}

			sharedAccountsList := []interface{}{}
			spaceSharesList := []interface{}{}
			for _, share := range node.ModuleShares {
				if share.To.Space != nil {
					spaceSharesList = append(spaceSharesList, share.To.Space.ID)
					continue
				}
				sharedAccountsList = append(sharedAccountsList, share.To.Account.Subdomain)
			}
			module["shared_accounts"] = sharedAccountsList
			module["space_shares"] = spaceSharesList

			switch node.Provider {
			case structs.VCSProviderAzureDevOps:
				if node.VCSIntegration != nil {
					module["azure_devops"] = []interface{}{
						map[string]interface{}{
							"id":         node.VCSIntegration.ID,
							"project":    node.Namespace,
							"is_default": node.VCSIntegration.IsDefault,
						},
					}
				}
			case structs.VCSProviderBitbucketCloud:
				if node.VCSIntegration != nil {
					module["bitbucket_cloud"] = []interface{}{
						map[string]interface{}{
							"id":         node.VCSIntegration.ID,
							"namespace":  node.Namespace,
							"is_default": node.VCSIntegration.IsDefault,
						},
					}
				}
			case structs.VCSProviderBitbucketDatacenter:
				if node.VCSIntegration != nil {
					module["bitbucket_datacenter"] = []interface{}{
						map[string]interface{}{
							"id":         node.VCSIntegration.ID,
							"namespace":  node.Namespace,
							"is_default": node.VCSIntegration.IsDefault,
						},
					}
				}
			case structs.VCSProviderGitHubEnterprise:
				if node.VCSIntegration != nil {
					module["github_enterprise"] = []interface{}{
						map[string]interface{}{
							"id":         node.VCSIntegration.ID,
							"namespace":  node.Namespace,
							"is_default": node.VCSIntegration.IsDefault,
						},
					}
				}
			case structs.VCSProviderGitlab:
				if node.VCSIntegration != nil {
					module["gitlab"] = []interface{}{
						map[string]interface{}{
							"id":         node.VCSIntegration.ID,
							"namespace":  node.Namespace,
							"is_default": node.VCSIntegration.IsDefault,
						},
					}
				}
			case structs.VCSProviderRawGit:
				module["raw_git"] = []interface{}{
					map[string]interface{}{
						"namespace": node.Namespace,
						"url":       node.RepositoryURL,
					},
				}
			}

			modules = append(modules, module)
		}

		if !query.SearchModulesOutput.PageInfo.HasNextPage {
			break
		}

		after := graphql.String(query.SearchModulesOutput.PageInfo.EndCursor)
		input.After = &after
	}

	d.SetId(fmt.Sprintf("modules-%d", time.Now().UnixNano()))

	if err := d.Set("modules", modules); err != nil {
		return diag.Errorf("could not set modules: %v", err)
	}

	return nil
}
