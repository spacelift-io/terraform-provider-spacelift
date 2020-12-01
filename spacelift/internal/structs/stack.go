package structs

const StackConfigVendorCloudFormation = "StackConfigVendorCloudFormation"
const StackConfigVendorPulumi = "StackConfigVendorPulumi"
const StackConfigVendorTerraform = "StackConfigVendorTerraform"

// Stack represents the Stack data relevant to the provider.
type Stack struct {
	ID             string   `graphql:"id"`
	Administrative bool     `graphql:"administrative"`
	Autodeploy     bool     `graphql:"autodeploy"`
	Autoretry      bool     `graphql:"autoretry"`
	BeforeInit     []string `graphql:"beforeInit"`
	Branch         string   `graphql:"branch"`
	Description    *string  `graphql:"description"`
	Integrations   *struct {
		AWS struct {
			AssumedRoleARN            *string `graphql:"assumedRoleArn"`
			AssumeRolePolicyStatement string  `graphql:"assumeRolePolicyStatement"`
		} `graphql:"aws"`
		GCP struct {
			ServiceAccountEmail *string  `graphql:"serviceAccountEmail"`
			TokenScopes         []string `graphql:"tokenScopes"`
		} `graphql:"gcp"`
		Webhooks []struct {
			ID       string `graphql:"id"`
			Enabled  bool   `graphql:"enabled"`
			Endpoint string `graphql:"endpoint"`
			Secret   string `graphql:"secret"`
		} `graphql:"webhooks"`
	} `graphql:"integrations"`
	Labels           []string `graphql:"labels"`
	ManagesStateFile bool     `graphql:"managesStateFile"`
	Name             string   `graphql:"name"`
	Namespace        string   `graphql:"namespace"`
	ProjectRoot      *string  `graphql:"projectRoot"`
	Provider         string   `graphql:"provider"`
	Repository       string   `graphql:"repository"`
	RunnerImage      *string  `graphql:"runnerImage"`
	TerraformVersion *string  `graphql:"terraformVersion"`
	VendorConfig     struct {
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
		} `graphql:"... on StackConfigVendorTerraform"`
	} `graphql:"vendorConfig"`
	WorkerPool *struct {
		ID string `graphql:"id"`
	} `graphql:"workerPool"`
}
