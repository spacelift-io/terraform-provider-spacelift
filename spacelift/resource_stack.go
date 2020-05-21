package spacelift

import (
	"net/http"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/structs"
)

func resourceStack() *schema.Resource {
	return &schema.Resource{
		Create: resourceStackCreate,
		Read:   resourceStackRead,
		Update: resourceStackUpdate,
		Delete: resourceStackDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"administrative": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Indicates whether this stack can manage others",
				Optional:    true,
				Default:     false,
			},
			"autodeploy": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Indicates whether changes to this stack can be automatically deployed",
				Optional:    true,
				Default:     false,
			},
			"aws_assume_role_policy_statement": &schema.Schema{
				Type:        schema.TypeString,
				Description: "AWS IAM assume role policy statement setting up trust relationship",
				Computed:    true,
			},
			"branch": &schema.Schema{
				Type:        schema.TypeString,
				Description: "GitHub branch to apply changes to",
				Required:    true,
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Free-form stack description for users",
				Optional:    true,
			},
			"gitlab": &schema.Schema{
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
			"import_state": &schema.Schema{
				Type:        schema.TypeString,
				Description: "State file to upload when creating a new stack",
				Optional:    true,
				DiffSuppressFunc: func(_, _, _ string, d *schema.ResourceData) bool {
					return d.Id() != ""
				},
			},
			"labels": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"manage_state": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Determines if Spacelift should manage state for this stack",
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Name of the stack - should be unique in one account",
				Required:    true,
				ForceNew:    true,
			},
			"repository": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Name of the repository, without the owner part",
				Required:    true,
			},
			"terraform_version": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Terraform version to use",
				Optional:    true,
			},
			"vcs_provider": &schema.Schema{
				Type:        schema.TypeString,
				Description: "VCS provider of the repository",
				Optional:    true,
				Default:     "GITHUB",
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

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
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

	if err := meta.(*Client).Query(&query, variables); err != nil {
		return errors.Wrap(err, "could not query for stack")
	}

	stack := query.Stack
	if stack == nil {
		d.SetId("")
		return nil
	}

	d.Set("administrative", stack.Administrative)
	d.Set("autodeploy", stack.Autodeploy)
	d.Set("aws_assume_role_policy_statement", stack.Integrations.AWS.AssumeRolePolicyStatement)
	d.Set("branch", stack.Branch)
	d.Set("manage_state", stack.ManagesStateFile)
	d.Set("name", stack.Name)
	d.Set("repository", stack.Repository)
	d.Set("vcs_provider", stack.Provider)

	if description := stack.Description; description != nil {
		d.Set("description", *description)
	}

	if stack.Provider == "GITLAB" {
		m := map[string]interface{}{
			"namespace": stack.Namespace,
		}
		err := d.Set("gitlab", []map[string]interface{}{m})
		if err != nil {
			errors.Wrap(err, "error setting gitlab (resource)")
		}
	}

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range stack.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	if terraformVersion := stack.TerraformVersion; terraformVersion != nil {
		d.Set("terraform_version", *terraformVersion)
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

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not update stack")
	}

	return resourceStackRead(d, meta)
}

func resourceStackDelete(d *schema.ResourceData, meta interface{}) error {
	var mutation struct {
		DeleteStack *structs.Stack `graphql:"stackDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*Client).Mutate(&mutation, variables); err != nil {
		return errors.Wrap(err, "could not delete stack")
	}

	d.SetId("")

	return nil
}

func stackInput(d *schema.ResourceData) structs.StackInput {
	ret := structs.StackInput{
		Administrative: graphql.Boolean(d.Get("administrative").(bool)),
		Autodeploy:     graphql.Boolean(d.Get("autodeploy").(bool)),
		Branch:         toString(d.Get("branch")),
		Name:           toString(d.Get("name")),
		Provider:       toString(d.Get("vcs_provider")),
		Repository:     toString(d.Get("repository")),
	}

	description, ok := d.GetOk("description")
	if ok {
		ret.Description = toOptionalString(description)
	}

	if gitlab, ok := d.Get("gitlab").([]interface{}); ok {
		if len(gitlab) > 0 {
			ret.Namespace = toString(gitlab[0].(map[string]interface{})["namespace"])
		}
	}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		var labels []graphql.String
		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}
		ret.Labels = &labels
	}

	terraformVersion, ok := d.GetOk("terraform_version")
	if ok {
		ret.TerraformVersion = toOptionalString(terraformVersion)
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

	if err := meta.(*Client).Mutate(&mutation, nil); err != nil {
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
