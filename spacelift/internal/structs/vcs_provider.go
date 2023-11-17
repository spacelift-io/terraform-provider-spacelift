package structs

type VCSProvider string

const (
	// VCSProviderAzureDevOps represents Azure DevOps VCS provider.
	VCSProviderAzureDevOps VCSProvider = "AZURE_DEVOPS"

	// VCSProviderBitbucketDatacenter represents Bitbucket Datacenter
	// (self-hosted) VCS provider.
	VCSProviderBitbucketDatacenter VCSProvider = "BITBUCKET_DATACENTER"

	// VCSProviderBitbucketCloud represents Bitbucket Cloud (managed) VCS
	// provider.
	VCSProviderBitbucketCloud VCSProvider = "BITBUCKET_CLOUD"

	// VCSProviderGitHub represents GitHub VCS provider.
	VCSProviderGitHub VCSProvider = "GITHUB"

	// VCSProviderGitHubEnterprise represents GitHub Enterprise (self-hosted)
	// VCS provider.
	VCSProviderGitHubEnterprise VCSProvider = "GITHUB_ENTERPRISE"

	// VCSProviderGitlab represents GitLab VCS provider.
	VCSProviderGitlab VCSProvider = "GITLAB"

	// VCSProviderRawGit represents raw Git link VCS provider.
	VCSProviderRawGit VCSProvider = "GIT"

	// VCSProviderShowcases represents the showcases provider.
	VCSProviderShowcases VCSProvider = "SHOWCASE"
)
