package structs

// StackConfigVendorCloudFormation is a graphql union typename.
const StackConfigVendorCloudFormation = "StackConfigVendorCloudFormation"

// StackConfigVendorPulumi is a graphql union typename.
const StackConfigVendorPulumi = "StackConfigVendorPulumi"

// StackConfigVendorTerraform is a graphql union typename.
const StackConfigVendorTerraform = "StackConfigVendorTerraform"

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
