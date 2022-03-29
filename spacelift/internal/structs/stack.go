package structs

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

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
	TerraformVersion    *string       `graphql:"terraformVersion"`
	VendorConfig        struct {
		Typename       string `graphql:"__typename"`
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
			Version   *string `graphql:"version"`
			Workspace *string `graphql:"workspace"`
		} `graphql:"... on StackConfigVendorTerraform"`
	} `graphql:"vendorConfig"`
	WorkerPool *struct {
		ID string `graphql:"id"`
	} `graphql:"workerPool"`
}

// ExportVCSSettings exports VCS settings into Terraform schema.
func (s *Stack) ExportVCSSettings(d *schema.ResourceData) error {
	var fieldName string
	vcsSettings := make(map[string]interface{})

	switch s.Provider {
	case VCSProviderAzureDevOps:
		vcsSettings["project"] = s.Namespace
		fieldName = "azure_devops"
	case VCSProviderBitbucketCloud:
		vcsSettings["namespace"] = s.Namespace
		fieldName = "bitbucket_cloud"
	case VCSProviderBitbucketDatacenter:
		vcsSettings["namespace"] = s.Namespace
		fieldName = "bitbucket_datacenter"
	case VCSProviderGitHubEnterprise:
		vcsSettings["namespace"] = s.Namespace
		fieldName = "github_enterprise"
	case VCSProviderGitlab:
		vcsSettings["namespace"] = s.Namespace
		fieldName = "gitlab"
	case VCSProviderShowcases:
		vcsSettings["namespace"] = s.Namespace
		fieldName = "showcase"
	}

	if fieldName != "" {
		if err := d.Set(fieldName, []interface{}{vcsSettings}); err != nil {
			return errors.Wrapf(err, "error setting %s (resource)", fieldName)
		}
	}

	return nil
}
