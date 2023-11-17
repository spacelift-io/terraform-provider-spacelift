package structs

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

// StackConfigVendorAnsible is a graphql union typename.
const StackConfigVendorAnsible = "StackConfigVendorAnsible" // #nosec G101 not a credential

// StackConfigVendorCloudFormation is a graphql union typename.
const StackConfigVendorCloudFormation = "StackConfigVendorCloudFormation"

// StackConfigVendorPulumi is a graphql union typename.
const StackConfigVendorPulumi = "StackConfigVendorPulumi"

// StackConfigVendorTerraform is a graphql union typename.
const StackConfigVendorTerraform = "StackConfigVendorTerraform"

// StackConfigVendorKubernetes is a graphql union typename.
const StackConfigVendorKubernetes = "StackConfigVendorKubernetes"

// Stack represents the Stack data relevant to the provider.
type Stack struct {
	ID                  string        `graphql:"id"`
	Administrative      bool          `graphql:"administrative"`
	AfterApply          []string      `graphql:"afterApply"`
	AfterDestroy        []string      `graphql:"afterDestroy"`
	AfterInit           []string      `graphql:"afterInit"`
	AfterPerform        []string      `graphql:"afterPerform"`
	AfterPlan           []string      `graphql:"afterPlan"`
	AfterRun            []string      `graphql:"afterRun"`
	Autodeploy          bool          `graphql:"autodeploy"`
	Autoretry           bool          `graphql:"autoretry"`
	BeforeApply         []string      `graphql:"beforeApply"`
	BeforeDestroy       []string      `graphql:"beforeDestroy"`
	BeforeInit          []string      `graphql:"beforeInit"`
	BeforePerform       []string      `graphql:"beforePerform"`
	BeforePlan          []string      `graphql:"beforePlan"`
	Branch              string        `graphql:"branch"`
	Deleting            bool          `graphql:"deleting"`
	Description         *string       `graphql:"description"`
	IsDisabled          bool          `graphql:"isDisabled"`
	GitHubActionDeploy  bool          `graphql:"githubActionDeploy"`
	Integrations        *Integrations `graphql:"integrations"`
	Labels              []string      `graphql:"labels"`
	LocalPreviewEnabled bool          `graphql:"localPreviewEnabled"`
	ManagesStateFile    bool          `graphql:"managesStateFile"`
	Name                string        `graphql:"name"`
	Namespace           string        `graphql:"namespace"`
	ProjectRoot         *string       `graphql:"projectRoot"`
	ProtectFromDeletion bool          `graphql:"protectFromDeletion"`
	Provider            VCSProvider   `graphql:"provider"`
	Repository          string        `graphql:"repository"`
	RepositoryURL       *string       `graphql:"repositoryURL"`
	RunnerImage         *string       `graphql:"runnerImage"`
	Space               string        `graphql:"space"`
	TerraformVersion    *string       `graphql:"terraformVersion"`
	VCSIntegration      *struct {
		ID          string   `graphql:"id"`
		Description string   `graphql:"description"`
		IsDefault   bool     `graphql:"isDefault"`
		Labels      []string `graphql:"labels"`
		Name        string   `graphql:"name"`
		Provider    string   `graphql:"provider"`
		Space       struct {
			ID string `graphql:"id"`
		} `graphql:"space"`
	} `graphql:"vcsIntegration"`
	VendorConfig struct {
		Typename string `graphql:"__typename"`
		Ansible  struct {
			Playbook string `graphql:"playbook"`
		} `graphql:"... on StackConfigVendorAnsible"`
		CloudFormation struct {
			EntryTemplateName string `graphql:"entryTemplateFile"`
			Region            string `graphql:"region"`
			StackName         string `graphql:"stackName"`
			TemplateBucket    string `graphql:"templateBucket"`
		} `graphql:"... on StackConfigVendorCloudFormation"`
		Kubernetes struct {
			Namespace      string  `graphql:"namespace"`
			KubectlVersion *string `graphql:"kubectlVersion"`
		} `graphql:"... on StackConfigVendorKubernetes"`
		Pulumi struct {
			LoginURL  string `graphql:"loginURL"`
			StackName string `graphql:"stackName"`
		} `graphql:"... on StackConfigVendorPulumi"`
		Terraform struct {
			UseSmartSanitization       bool    `graphql:"useSmartSanitization"`
			Version                    *string `graphql:"version"`
			WorkflowTool               *string `graphql:"workflowTool"`
			Workspace                  *string `graphql:"workspace"`
			ExternalStateAccessEnabled bool    `graphql:"externalStateAccessEnabled"`
		} `graphql:"... on StackConfigVendorTerraform"`
	} `graphql:"vendorConfig"`
	WorkerPool *struct {
		ID string `graphql:"id"`
	} `graphql:"workerPool"`
}

// ExportVCSSettings exports VCS settings into Terraform schema.
func (s *Stack) ExportVCSSettings(d *schema.ResourceData) error {
	if fieldName, vcsSettings := s.VCSSettings(); fieldName != "" {
		if err := d.Set(fieldName, []interface{}{vcsSettings}); err != nil {
			return errors.Wrapf(err, "error setting %s (resource)", fieldName)
		}
	}

	return nil
}

// IaC returns IaC settings of a stack.
func (s *Stack) IaCSettings() (string, map[string]interface{}) {
	switch s.VendorConfig.Typename {
	case StackConfigVendorAnsible:
		return "ansible", singleKeyMap("playbook", s.VendorConfig.Ansible.Playbook)
	case StackConfigVendorCloudFormation:
		return "cloudformation", map[string]interface{}{
			"entry_template_file": s.VendorConfig.CloudFormation.EntryTemplateName,
			"region":              s.VendorConfig.CloudFormation.Region,
			"stack_name":          s.VendorConfig.CloudFormation.StackName,
			"template_bucket":     s.VendorConfig.CloudFormation.TemplateBucket,
		}
	case StackConfigVendorKubernetes:
		return "kubernetes", map[string]interface{}{
			"namespace":       s.VendorConfig.Kubernetes.Namespace,
			"kubectl_version": s.VendorConfig.Kubernetes.KubectlVersion,
		}
	case StackConfigVendorPulumi:
		return "pulumi", map[string]interface{}{
			"login_url":  s.VendorConfig.Pulumi.LoginURL,
			"stack_name": s.VendorConfig.Pulumi.StackName,
		}
	}

	return "", nil
}

// VCSSettings returns VCS settings of a stack.
func (s *Stack) VCSSettings() (string, map[string]interface{}) {
	switch s.Provider {
	case VCSProviderAzureDevOps:
		return "azure_devops", singleKeyMap("project", s.Namespace)
	case VCSProviderBitbucketCloud:
		return "bitbucket_cloud", singleKeyMap("namespace", s.Namespace)
	case VCSProviderBitbucketDatacenter:
		return "bitbucket_datacenter", singleKeyMap("namespace", s.Namespace)
	case VCSProviderGitHubEnterprise:
		return "github_enterprise", map[string]interface{}{
			"id":          s.VCSIntegration.ID,
			"name":        s.VCSIntegration.Name,
			"namespace":   s.Namespace,
			"description": s.VCSIntegration.Description,
			"is_default":  s.VCSIntegration.IsDefault,
			"labels":      s.VCSIntegration.Labels,
			"space_id":    s.VCSIntegration.Space.ID,
		}
	case VCSProviderGitlab:
		return "gitlab", map[string]interface{}{
			"id":          s.VCSIntegration.ID,
			"name":        s.VCSIntegration.Name,
			"namespace":   s.Namespace,
			"description": s.VCSIntegration.Description,
			"is_default":  s.VCSIntegration.IsDefault,
			"labels":      s.VCSIntegration.Labels,
			"space_id":    s.VCSIntegration.Space.ID,
		}
	case VCSProviderRawGit:
		return "raw_git", map[string]interface{}{
			"namespace": s.Namespace,
			"url":       s.RepositoryURL,
		}
	case VCSProviderShowcases:
		return "showcase", singleKeyMap("namespace", s.Namespace)
	}

	return "", nil
}

func PopulateStack(d *schema.ResourceData, stack *Stack) error {
	d.Set("administrative", stack.Administrative)
	d.Set("after_apply", stack.AfterApply)
	d.Set("after_destroy", stack.AfterDestroy)
	d.Set("after_init", stack.AfterInit)
	d.Set("after_perform", stack.AfterPerform)
	d.Set("after_plan", stack.AfterPlan)
	d.Set("after_run", stack.AfterRun)
	d.Set("autodeploy", stack.Autodeploy)
	d.Set("autoretry", stack.Autoretry)
	d.Set("aws_assume_role_policy_statement", stack.Integrations.AWS.AssumeRolePolicyStatement)
	d.Set("before_apply", stack.BeforeApply)
	d.Set("before_destroy", stack.BeforeDestroy)
	d.Set("before_init", stack.BeforeInit)
	d.Set("before_perform", stack.BeforePerform)
	d.Set("before_plan", stack.BeforePlan)
	d.Set("branch", stack.Branch)
	d.Set("description", stack.Description)
	d.Set("enable_local_preview", stack.LocalPreviewEnabled)
	d.Set("github_action_deploy", stack.GitHubActionDeploy)
	d.Set("manage_state", stack.ManagesStateFile)
	d.Set("name", stack.Name)
	d.Set("project_root", stack.ProjectRoot)
	d.Set("protect_from_deletion", stack.ProtectFromDeletion)
	d.Set("repository", stack.Repository)
	d.Set("runner_image", stack.RunnerImage)
	d.Set("space_id", stack.Space)
	d.Set("slug", stack.ID)

	if err := stack.ExportVCSSettings(d); err != nil {
		return err
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range stack.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	switch stack.VendorConfig.Typename {
	case StackConfigVendorAnsible:
		m := map[string]interface{}{
			"playbook": stack.VendorConfig.Ansible.Playbook,
		}

		d.Set("ansible", []interface{}{m})
	case StackConfigVendorCloudFormation:
		m := map[string]interface{}{
			"entry_template_file": stack.VendorConfig.CloudFormation.EntryTemplateName,
			"region":              stack.VendorConfig.CloudFormation.Region,
			"stack_name":          stack.VendorConfig.CloudFormation.StackName,
			"template_bucket":     stack.VendorConfig.CloudFormation.TemplateBucket,
		}

		d.Set("cloudformation", []interface{}{m})
	case StackConfigVendorKubernetes:
		m := map[string]interface{}{
			"namespace":       stack.VendorConfig.Kubernetes.Namespace,
			"kubectl_version": stack.VendorConfig.Kubernetes.KubectlVersion,
		}

		d.Set("kubernetes", []interface{}{m})
	case StackConfigVendorPulumi:
		m := map[string]interface{}{
			"login_url":  stack.VendorConfig.Pulumi.LoginURL,
			"stack_name": stack.VendorConfig.Pulumi.StackName,
		}

		d.Set("pulumi", []interface{}{m})
	default:
		d.Set("terraform_smart_sanitization", stack.VendorConfig.Terraform.UseSmartSanitization)
		d.Set("terraform_version", stack.VendorConfig.Terraform.Version)
		d.Set("terraform_workflow_tool", stack.VendorConfig.Terraform.WorkflowTool)
		d.Set("terraform_workspace", stack.VendorConfig.Terraform.Workspace)
		d.Set("terraform_external_state_access", stack.VendorConfig.Terraform.ExternalStateAccessEnabled)
	}

	if workerPool := stack.WorkerPool; workerPool != nil {
		d.Set("worker_pool_id", workerPool.ID)
	} else {
		d.Set("worker_pool_id", nil)
	}

	return nil
}

func singleKeyMap(key, val string) map[string]interface{} {
	return map[string]interface{}{key: val}
}
