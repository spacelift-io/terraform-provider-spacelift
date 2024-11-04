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
	Provider            VCSProvider  `graphql:"provider"`
	Public              bool         `graphql:"public"`
	Repository          string       `graphql:"repository"`
	RepositoryURL       *string      `graphql:"repositoryURL"`
	SharedAccounts      []string     `graphql:"sharedAccounts"`
	Space               string       `graphql:"space"`
	SpaceDetails        Space        `graphql:"spaceDetails"`
	TerraformProvider   string       `graphql:"terraformProvider"`
	VCSIntegration      *struct {
		ID        string `graphql:"id"`
		IsDefault bool   `graphql:"isDefault"`
	} `graphql:"vcsIntegration"`
	WorkerPool *struct {
		ID string `graphql:"id"`
	} `graphql:"workerPool"`
	WorkflowTool *string `graphql:"workflowTool"`
}

// ExportVCSSettings exports VCS settings into Terraform schema.
func (m *Module) ExportVCSSettings(d *schema.ResourceData) error {
	var fieldName string
	var vcsSettings map[string]interface{}

	switch m.Provider {
	case VCSProviderAzureDevOps:
		if m.VCSIntegration != nil {
			vcsSettings = map[string]interface{}{
				"id":         m.VCSIntegration.ID,
				"project":    m.Namespace,
				"is_default": m.VCSIntegration.IsDefault,
			}
		}
		fieldName = "azure_devops"
	case VCSProviderBitbucketCloud:
		if m.VCSIntegration != nil {
			vcsSettings = map[string]interface{}{
				"id":         m.VCSIntegration.ID,
				"namespace":  m.Namespace,
				"is_default": m.VCSIntegration.IsDefault,
			}
		}
		fieldName = "bitbucket_cloud"
	case VCSProviderBitbucketDatacenter:
		if m.VCSIntegration != nil {
			vcsSettings = map[string]interface{}{
				"id":         m.VCSIntegration.ID,
				"namespace":  m.Namespace,
				"is_default": m.VCSIntegration.IsDefault,
			}
		}
		fieldName = "bitbucket_datacenter"
	case VCSProviderGitHubEnterprise:
		if m.VCSIntegration != nil {
			vcsSettings = map[string]interface{}{
				"id":         m.VCSIntegration.ID,
				"namespace":  m.Namespace,
				"is_default": m.VCSIntegration.IsDefault,
			}
		}
		fieldName = "github_enterprise"
	case VCSProviderGitlab:
		if m.VCSIntegration != nil {
			vcsSettings = map[string]interface{}{
				"id":         m.VCSIntegration.ID,
				"namespace":  m.Namespace,
				"is_default": m.VCSIntegration.IsDefault,
			}
		}
		fieldName = "gitlab"
	case VCSProviderRawGit:
		vcsSettings = map[string]interface{}{
			"namespace": m.Namespace,
			"url":       m.RepositoryURL,
		}
		fieldName = "raw_git"
	}

	if fieldName != "" {
		if err := d.Set(fieldName, []interface{}{vcsSettings}); err != nil {
			return errors.Wrapf(err, "error setting %s (resource)", fieldName)
		}
	}

	return nil
}
