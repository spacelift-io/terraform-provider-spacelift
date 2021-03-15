package spacelift

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

const vcsProviderGitlab = "GITLAB"

func resourceStack() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceStackCreate,
		ReadContext:   resourceStackRead,
		UpdateContext: resourceStackUpdate,
		DeleteContext: resourceStackDelete,

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
			"before_apply": {
				Type:        schema.TypeList,
				Description: "List of before-apply scripts",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
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
				ConflictsWith: []string{"pulumi", "terraform_version", "terraform_workspace"},
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
				ConflictsWith: []string{"cloudformation", "terraform_version", "terraform_workspace"},
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
			"terraform_version": {
				Type:             schema.TypeString,
				Description:      "Terraform version to use",
				Optional:         true,
				DiffSuppressFunc: onceTheVersionIsSetDoNotUnset,
			},
			"terraform_workspace": {
				Type:        schema.TypeString,
				Description: "Terraform workspace to select",
				Optional:    true,
			},
			"wait_for_destroy": {
				Type:        schema.TypeBool,
				Default:     true,
				Description: "If the Stack is marked with destroy_on_delete, wait for the destruction and deletion to finish. This might take a long time, if the stacks you destroy also destroy underlying other stacks and wait for them, etc.",
				Optional:    true,
			},
			"worker_pool_id": {
				Type:        schema.TypeString,
				Description: "ID of the worker pool to use",
				Optional:    true,
			},
		},
	}
}

func resourceStackCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return diag.Errorf(`"import_state" requires "manage_state" to be true`)
	} else if ok {
		objectID, err := uploadStateFile(ctx, content.(string), meta)
		if err != nil {
			return diag.FromErr(err)
		}
		variables["stackObjectID"] = toOptionalString(objectID)
	}

	if err := meta.(*internal.Client).Mutate(ctx, &mutation, variables); err != nil {
		return diag.Errorf("could not create stack: %v", err)
	}

	d.SetId(mutation.CreateStack.ID)

	return resourceStackRead(ctx, d, meta)
}

func resourceStackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}

	if err := meta.(*internal.Client).Query(ctx, &query, variables); err != nil {
		return diag.Errorf("could not query for stack: %v", err)
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
	d.Set("before_apply", stack.BeforeApply)
	d.Set("before_init", stack.BeforeInit)
	d.Set("branch", stack.Branch)
	d.Set("description", stack.Description)
	d.Set("manage_state", stack.ManagesStateFile)
	d.Set("name", stack.Name)
	d.Set("project_root", stack.ProjectRoot)
	d.Set("repository", stack.Repository)
	d.Set("runner_image", stack.RunnerImage)

	if stack.Provider == "GITLAB" {
		m := map[string]interface{}{
			"namespace": stack.Namespace,
		}

		if err := d.Set("gitlab", []interface{}{m}); err != nil {
			return diag.Errorf("error setting gitlab (resource): %v", err)
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
	default:
		d.Set("terraform_version", stack.VendorConfig.Terraform.Version)
		d.Set("terraform_workspace", stack.VendorConfig.Terraform.Workspace)
	}

	if workerPool := stack.WorkerPool; workerPool != nil {
		d.Set("worker_pool_id", workerPool.ID)
	} else {
		d.Set("worker_pool_id", nil)
	}

	return nil
}

func resourceStackUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UpdateStack structs.Stack `graphql:"stackUpdate(id: $id, input: $input)"`
	}

	variables := map[string]interface{}{
		"id":    toID(d.Id()),
		"input": stackInput(d),
	}

	var ret diag.Diagnostics

	if err := meta.(*internal.Client).Mutate(ctx, &mutation, variables); err != nil {
		ret = diag.Errorf("could not update stack: %v", err)
	}

	return append(ret, resourceStackRead(ctx, d, meta)...)
}

func resourceStackDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteStack *structs.Stack `graphql:"stackDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, &mutation, variables); err != nil {
		return diag.Errorf("could not delete stack: %v", err)
	}

	if mutation.DeleteStack.Deleting {
		if wait := d.Get("wait_for_destroy"); wait != nil && wait.(bool) {
			if diagnostics := waitForDestroy(ctx, meta.(*internal.Client), d.Id()); diagnostics.HasError() {
				return diagnostics
			}
		}
	}

	d.SetId("")

	return nil
}

func waitForDestroy(ctx context.Context, client *internal.Client, id string) diag.Diagnostics {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return diag.FromErr(ctx.Err())
		case <-ticker.C:
		}

		var query struct {
			Stack *structs.Stack `graphql:"stack(id: $id)"`
		}

		variables := map[string]interface{}{"id": graphql.ID(id)}

		if err := client.Query(ctx, &query, variables); err != nil {
			return diag.Errorf("could not query for stack: %v", err)
		}

		stack := query.Stack
		if stack == nil {
			return nil
		}

		if !stack.Deleting {
			return diag.Errorf("destruction of Stack unsuccessful, please check the destruction run logs")
		}
	}
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

	beforeApplies := []graphql.String{}
	if commands, ok := d.GetOk("before_apply"); ok {
		for _, cmd := range commands.([]interface{}) {
			beforeApplies = append(beforeApplies, graphql.String(cmd.(string)))
		}
	}
	ret.BeforeApply = &beforeApplies

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
	} else {
		terraformConfig := &structs.TerraformInput{}

		if terraformVersion, ok := d.GetOk("terraform_version"); ok {
			terraformConfig.Version = toOptionalString(terraformVersion)
		}

		if terraformWorkspace, ok := d.GetOk("terraform_workspace"); ok {
			terraformConfig.Workspace = toOptionalString(terraformWorkspace)
		}

		ret.VendorConfig = &structs.VendorConfigInput{Terraform: terraformConfig}
	}

	if workerPoolID, ok := d.GetOk("worker_pool_id"); ok {
		ret.WorkerPool = graphql.NewID(workerPoolID)
	}

	return ret
}

func uploadStateFile(ctx context.Context, content string, meta interface{}) (string, error) {
	var mutation struct {
		StateUploadURL struct {
			ObjectID string `graphql:"objectId"`
			URL      string `graphql:"url"`
		} `graphql:"stateUploadUrl"`
	}

	if err := meta.(*internal.Client).Mutate(ctx, &mutation, nil); err != nil {
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
