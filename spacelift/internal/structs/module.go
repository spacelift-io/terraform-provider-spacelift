package structs

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

// Module represents the Module data relevant to the provider.
type Module struct {
	ID                  string       `graphql:"id"`
	Administrative      bool         `graphql:"administrative"`
	Branch              string       `graphql:"branch"`
	Description         *string      `graphql:"description"`
	Integrations        Integrations `graphql:"integrations"`
	Labels              []string     `graphql:"labels"`
	LocalPreviewEnabled bool         `graphql:"localPreviewEnabled"`
	Name                string       `graphql:"name"`
	Namespace           string       `graphql:"namespace"`
	ProjectRoot         *string      `graphql:"projectRoot"`
	ProtectFromDeletion bool         `graphql:"protectFromDeletion"`
	Provider            string       `graphql:"provider"`
	Repository          string       `graphql:"repository"`
	SharedAccounts      []string     `graphql:"sharedAccounts"`
	Space               string       `graphql:"space"`
	TerraformProvider   string       `graphql:"terraformProvider"`
	WorkerPool          *struct {
		ID string `graphql:"id"`
	} `graphql:"workerPool"`
	WorkflowTool *string `graphql:"workflowTool"`
}

// ExportVCSSettings exports VCS settings into Terraform schema.
func (m *Module) ExportVCSSettings(d *schema.ResourceData) error {
	var fieldName string
	vcsSettings := make(map[string]interface{})

	switch m.Provider {
	case VCSProviderAzureDevOps:
		vcsSettings["project"] = m.Namespace
		fieldName = "azure_devops"
	case VCSProviderBitbucketCloud:
		vcsSettings["namespace"] = m.Namespace
		fieldName = "bitbucket_cloud"
	case VCSProviderBitbucketDatacenter:
		vcsSettings["namespace"] = m.Namespace
		fieldName = "bitbucket_datacenter"
	case VCSProviderGitHubEnterprise:
		vcsSettings["namespace"] = m.Namespace
		fieldName = "github_enterprise"
	case VCSProviderGitlab:
		vcsSettings["namespace"] = m.Namespace
		fieldName = "gitlab"
	}

	if fieldName != "" {
		if err := d.Set(fieldName, []interface{}{vcsSettings}); err != nil {
			return errors.Wrapf(err, "error setting %s (resource)", fieldName)
		}
	}

	return nil
}
