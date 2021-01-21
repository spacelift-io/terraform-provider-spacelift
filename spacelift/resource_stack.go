package spacelift

import (
	"net/http"
	"strings"

	"github.com/fluxio/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

const vcsProviderGitlab = "GITLAB"

func resourceStack() *schema.Resource {
	return &schema.Resource{
		Create: resourceStackCreate,
		Read:   resourceStackRead,
		Update: resourceStackUpdate,
		Delete: resourceStackDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"administrative": {
				Type:        schema.TypeBool,
				Description: "Indicates whether this stack can manage others",
				Optional:    true,
				Default:     false,
			},
			"autodeploy": {
				Type:        schema.TypeBool,
				Description: "Indicates whether changes to this stack can be automatically deployed",
				Optional:    true,
				Default:     false,
			},
			"autoretry": {
				Type:        schema.TypeBool,
				Description: "Indicates whether obsolete proposed changes should automatically be retried",
				Optional:    true,
				Default:     false,
			},
			"aws_assume_role_policy_statement": {
				Type:        schema.TypeString,
				Description: "AWS IAM assume role policy statement setting up trust relationship",
				Computed:    true,
			},
			"before_init": {
				Type:        schema.TypeList,
				Description: "List of before-init scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
			"branch": {
				Type:        schema.TypeString,
				Description: "GitHub branch to apply changes to",
				Required:    true,
			},
			"cloudformation": {
				Type:          schema.TypeList,
				ConflictsWith: []string{"pulumi", "terraform"},
				Description:   "CloudFormation-specific configuration. Presence means this Stack is a CloudFormation Stack.",
				Optional:      true,
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"entry_template_file": {
							Type:        schema.TypeString,
							Description: "Template file `cloudformation package` will be called on",
							Required:    true,
						},
						"region": {
							Type:        schema.TypeString,
							Description: "AWS region to use",
							Required:    true,
						},
						"stack_name": {
							Type:        schema.TypeString,
							Description: "CloudFormation stack name",
							Required:    true,
						},
						"template_bucket": {
							Type:        schema.TypeString,
							Description: "S3 bucket to save CloudFormation templates to",
							Required:    true,
						},
					},
				},
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Free-form stack description for users",
				Optional:    true,
			},
			"gitlab": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"import_state": {
				Type:        schema.TypeString,
				Description: "State file to upload when creating a new stack",
				Optional:    true,
				DiffSuppressFunc: func(_, _, _ string, d *schema.ResourceData) bool {
					return d.Id() != ""
				},
			},
			"labels": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"manage_state": {
				Type:        schema.TypeBool,
				Description: "Determines if Spacelift should manage state for this stack",
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the stack - should be unique in one account",
				Required:    true,
			},
			"project_root": {
				Type:        schema.TypeString,
				Description: "Project root is the optional directory relative to the workspace root containing the entrypoint to the Stack.",
				Optional:    true,
			},
			"pulumi": {
				Type:          schema.TypeList,
				ConflictsWith: []string{"cloudformation", "terraform"},
				Description:   "Pulumi-specific configuration. Presence means this Stack is a Pulumi Stack.",
				Optional:      true,
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"login_url": {
							Type:        schema.TypeString,
							Description: "State backend to log into on Run initialize.",
							Required:    true,
						},
						"stack_name": {
							Type:        schema.TypeString,
							Description: "Pulumi stack name to use with the state backend.",
							Required:    true,
						},
					},
				},
			},
			"repository": {
				Type:        schema.TypeString,
				Description: "Name of the repository, without the owner part",
				Required:    true,
			},
			"runner_image": {
				Type:        schema.TypeString,
				Description: "Name of the Docker image used to process Runs",
				Optional:    true,
			},
			"terraform": {
				Type:          schema.TypeList,
				ConflictsWith: []string{"cloudformation", "pulumi", "terraform_version"},
				Description:   "Terraform-specific configuration. Presence means this Stack is a Terraform Stack.",
				Optional:      true,
				Computed:      true,
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:             schema.TypeString,
							Description:      "Terraform version to be used by the Stack",
							Optional:         true,
							DiffSuppressFunc: onceTheVersionIsSetDoNotUnset,
						},
						"workspace": {
							Type:        schema.TypeString,
							Description: "Workspace to select before performing Terraform operations",
							Optional:    true,
						},
					},
				},
			},
			"terraform_version": {
				Type:             schema.TypeString,
				ConflictsWith:    []string{"terraform"},
				Deprecated:       `Please use the "terraform" block instead`,
				Description:      "Terraform version to use",
				Optional:         true,
				DiffSuppressFunc: onceTheVersionIsSetDoNotUnset,
			},
			"worker_pool_id": {
				Type:        schema.TypeString,
				Description: "ID of the worker pool to use",
				Optional:    true,
			},
		},
	}
}

func resourceStackCreate(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		CreateStack structs.Stack `graphql:"stackCreate(input: $input, manageState: $manageState, stackObjectID: $stackObjectID)"`
	}

	manageState := d.Get("manage_state").(bool)

	variables := map[string]interface{}{
		"input":         stackInput(d),
		"manageState":   graphql.Boolean(manageState),
		"stackObjectID": (*graphql.String)(nil),
	}

	content, ok := d.GetOk("import_state")
	if ok && !manageState {
		return errors.New(`"import_state" requires "manage_state" to be true`)
	} else if ok {
		objectID, err := uploadStateFile(content.(string), meta)
		if err != nil {
			return err
		}
		variables["stackObjectID"] = toOptionalString(objectID)
	}

	if err := meta.(*internal.Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not create stack")
	}

	d.SetId(mutation.CreateStack.ID)

	return resourceStackRead(d, meta)
}

func resourceStackRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}

	if err := meta.(*internal.Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for stack")
	}

	stack := query.Stack
	if stack == nil {
		d.SetId("")
		return nil
	}

	d.Set("administrative", stack.Administrative)
	d.Set("autodeploy", stack.Autodeploy)
	d.Set("autoretry", stack.Autoretry)
	d.Set("aws_assume_role_policy_statement", stack.Integrations.AWS.AssumeRolePolicyStatement)
	d.Set("before_init", stack.BeforeInit)
	d.Set("branch", stack.Branch)
	d.Set("description", stack.Description)
	d.Set("manage_state", stack.ManagesStateFile)
	d.Set("name", stack.Name)
	d.Set("project_root", stack.ProjectRoot)
	d.Set("repository", stack.Repository)
	d.Set("runner_image", stack.RunnerImage)
	d.Set("terraform_version", stack.TerraformVersion)

	if stack.Provider == "GITLAB" {
		m := map[string]interface{}{
			"namespace": stack.Namespace,
		}

		if err := d.Set("gitlab", []interface{}{m}); err != nil {
			return errors.Wrap(err, "error setting gitlab (resource)")
		}
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range stack.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	switch stack.VendorConfig.Typename {
	case structs.StackConfigVendorCloudFormation:
		m := map[string]interface{}{
			"entry_template_file": stack.VendorConfig.CloudFormation.EntryTemplateName,
			"region":              stack.VendorConfig.CloudFormation.Region,
			"stack_name":          stack.VendorConfig.CloudFormation.StackName,
			"template_bucket":     stack.VendorConfig.CloudFormation.TemplateBucket,
		}

		d.Set("cloudformation", []interface{}{m})
	case structs.StackConfigVendorPulumi:
		m := map[string]interface{}{
			"login_url":  stack.VendorConfig.Pulumi.LoginURL,
			"stack_name": stack.VendorConfig.Pulumi.StackName,
		}

		d.Set("pulumi", []interface{}{m})
	case structs.StackConfigVendorTerraform:
		version := stack.VendorConfig.Terraform.Version
		workspace := stack.VendorConfig.Terraform.Workspace

		// Corner case: we don't *require* the "terraform" block to be set, but
		// it is perfectly valid to make it empty and since both of its members
		// are optional and can come back as `nil`, we would unset the key
		// someone has set explicitly.
		if _, ok := d.GetOk("terraform"); ok || version != nil || workspace != nil {
			d.Set("terraform", []interface{}{
				map[string]interface{}{
					"version":   version,
					"workspace": workspace,
				},
			})
		} else {
			d.Set("terraform", nil)
		}
	}

	if workerPool := stack.WorkerPool; workerPool != nil {
		d.Set("worker_pool_id", workerPool.ID)
	} else {
		d.Set("worker_pool_id", nil)
	}

	return nil
}

func resourceStackUpdate(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		UpdateStack structs.Stack `graphql:"stackUpdate(id: $id, input: $input)"`
	}

	variables := map[string]interface{}{
		"id":    toID(d.Id()),
		"input": stackInput(d),
	}

	var acc multierror.Accumulator

	acc.Push(errors.Wrap(meta.(*internal.Client).Mutate(&mutation, variables), "could not update stack"))
	acc.Push(errors.Wrap(resourceStackRead(d, meta), "could not read the current state"))

	return acc.Error()
}

func resourceStackDelete(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		DeleteStack *structs.Stack `graphql:"stackDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not delete stack")
	}

	d.SetId("")

	return nil
}

func stackInput(d *schema.ResourceData) structs.StackInput {
	ret := structs.StackInput{
		Administrative: graphql.Boolean(d.Get("administrative").(bool)),
		Autodeploy:     graphql.Boolean(d.Get("autodeploy").(bool)),
		Autoretry:      graphql.Boolean(d.Get("autoretry").(bool)),
		Branch:         toString(d.Get("branch")),
		Name:           toString(d.Get("name")),
		Repository:     toString(d.Get("repository")),
	}

	beforeInits := []graphql.String{}
	if commands, ok := d.GetOk("before_init"); ok {
		for _, cmd := range commands.([]interface{}) {
			beforeInits = append(beforeInits, graphql.String(cmd.(string)))
		}
	}
	ret.BeforeInit = &beforeInits

	description, ok := d.GetOk("description")
	if ok {
		ret.Description = toOptionalString(description)
	}

	foundGitlab := false
	if gitlab, ok := d.Get("gitlab").([]interface{}); ok {
		if len(gitlab) > 0 {
			foundGitlab = true
			ret.Namespace = toOptionalString(gitlab[0].(map[string]interface{})["namespace"])
			ret.Provider = graphql.NewString(vcsProviderGitlab)
		}
	}
	if !foundGitlab {
		ret.Provider = graphql.NewString("GITHUB")
	}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String
		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}
		ret.Labels = &labels
	}

	if projectRoot, ok := d.GetOk("project_root"); ok {
		ret.ProjectRoot = toOptionalString(projectRoot)
	}

	if runnerImage, ok := d.GetOk("runner_image"); ok {
		ret.RunnerImage = toOptionalString(runnerImage)
	}

	if terraformVersion, ok := d.GetOk("terraform_version"); ok {
		ret.VendorConfig = &structs.VendorConfigInput{Terraform: &structs.TerraformInput{
			Version: toOptionalString(terraformVersion),
		}}
	}

	if cloudFormation, ok := d.Get("cloudformation").([]interface{}); ok && len(cloudFormation) > 0 {
		ret.VendorConfig = &structs.VendorConfigInput{
			CloudFormationInput: &structs.CloudFormationInput{
				EntryTemplateFile: toString(cloudFormation[0].(map[string]interface{})["entry_template_file"]),
				Region:            toString(cloudFormation[0].(map[string]interface{})["region"]),
				StackName:         toString(cloudFormation[0].(map[string]interface{})["stack_name"]),
				TemplateBucket:    toString(cloudFormation[0].(map[string]interface{})["template_bucket"]),
			},
		}
	} else if pulumi, ok := d.Get("pulumi").([]interface{}); ok && len(pulumi) > 0 {
		ret.VendorConfig = &structs.VendorConfigInput{
			Pulumi: &structs.PulumiInput{
				LoginURL:  toString(pulumi[0].(map[string]interface{})["login_url"]),
				StackName: toString(pulumi[0].(map[string]interface{})["stack_name"]),
			},
		}
	} else if terraform, ok := d.Get("terraform").([]interface{}); ok && len(terraform) > 0 {
		terraformConfig := &structs.TerraformInput{}

		if version := terraform[0].(map[string]interface{})["version"]; version != nil && version.(string) != "" {
			terraformConfig.Version = toOptionalString(version)
		}

		if workspace := terraform[0].(map[string]interface{})["workspace"]; workspace != nil && workspace.(string) != "" {
			terraformConfig.Workspace = toOptionalString(workspace)
		}

		ret.VendorConfig = &structs.VendorConfigInput{Terraform: terraformConfig}
	} else {
		ret.VendorConfig = &structs.VendorConfigInput{Terraform: &structs.TerraformInput{}}
	}

	if workerPoolID, ok := d.GetOk("worker_pool_id"); ok {
		ret.WorkerPool = graphql.NewID(workerPoolID)
	}

	return ret
}

func uploadStateFile(content string, meta interface{}) (string, error) {
	var mutation struct {
		StateUploadURL struct {
			ObjectID string `graphql:"objectId"`
			URL      string `graphql:"url"`
		} `graphql:"stateUploadUrl"`
	}

	if err := meta.(*internal.Client).Mutate(&mutation, nil); err != nil {
		return "", errors.Wrap(err, "could not generate state upload URL")
	}

	response, err := http.Post(mutation.StateUploadURL.URL, "application/json", strings.NewReader(content))
	if err != nil {
		return "", errors.Wrap(err, "could not upload the state to remote URL")
	}

	if (response.StatusCode / 100) != 2 {
		return "", errors.Errorf("unexpected HTTP status code when uploading the state: %d", response.StatusCode)
	}

	return mutation.StateUploadURL.ObjectID, nil
}

func onceTheVersionIsSetDoNotUnset(_, _, new string, _ *schema.ResourceData) bool {
	return new == ""
}
