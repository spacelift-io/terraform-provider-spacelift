package structs

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

// StackConfigVendorAnsible is a graphql union typename.
const StackConfigVendorAnsible = "StackConfigVendorAnsible"

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
	GitHubActionDeploy  bool          `graphql:"githubActionDeploy"`
	Integrations        *Integrations `graphql:"integrations"`
	Labels              []string      `graphql:"labels"`
	LocalPreviewEnabled bool          `graphql:"localPreviewEnabled"`
	ManagesStateFile    bool          `graphql:"managesStateFile"`
	Name                string        `graphql:"name"`
	Namespace           string        `graphql:"namespace"`
	ProjectRoot         *string       `graphql:"projectRoot"`
	ProtectFromDeletion bool          `graphql:"protectFromDeletion"`
	Provider            string        `graphql:"provider"`
	Repository          string        `graphql:"repository"`
	RunnerImage         *string       `graphql:"runnerImage"`
	Space               string        `graphql:"space"`
	TerraformVersion    *string       `graphql:"terraformVersion"`
	VendorConfig        struct {
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
			Namespace string `graphql:"namespace"`
		} `graphql:"... on StackConfigVendorKubernetes"`
		Pulumi struct {
			LoginURL  string `graphql:"loginURL"`
			StackName string `graphql:"stackName"`
		} `graphql:"... on StackConfigVendorPulumi"`
		Terraform struct {
			UseSmartSanitization bool    `graphql:"useSmartSanitization"`
			Version              *string `graphql:"version"`
			Workspace            *string `graphql:"workspace"`
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
		return "kubernetes", singleKeyMap("namespace", s.VendorConfig.Kubernetes.Namespace)
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
		return "github_enterprise", singleKeyMap("namespace", s.Namespace)
	case VCSProviderGitlab:
		return "gitlab", singleKeyMap("namespace", s.Namespace)
	case VCSProviderShowcases:
		return "showcase", singleKeyMap("namespace", s.Namespace)
	}

	return "", nil
}

func singleKeyMap(key, val string) map[string]interface{} {
	return map[string]interface{}{key: val}
}
