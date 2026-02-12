package spacelift

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceTemplateDeployment() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_template_deployment` deploys a Spacelift template, " +
			"creating and managing the stacks defined in the template.",

		CreateContext: resourceTemplateDeploymentCreate,
		ReadContext:   resourceTemplateDeploymentRead,
		UpdateContext: resourceTemplateDeploymentUpdate,
		DeleteContext: resourceTemplateDeploymentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"template_version_id": {
				Type:             schema.TypeString,
				Description:      "ID of the template version to deploy. Changing this will upgrade the deployment to the new version.",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"space": {
				Type:             schema.TypeString,
				Description:      "Space where the stacks created by this deployment will be placed",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the deployment",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of the deployment",
				Optional:    true,
			},
			"input": {
				Type:        schema.TypeList,
				Description: "Input values for the template",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:             schema.TypeString,
							Description:      "Input identifier",
							Required:         true,
							ValidateDiagFunc: validations.DisallowEmptyString,
						},
						"value": {
							Type:             schema.TypeString,
							Description:      "Input value",
							Required:         true,
							DiffSuppressFunc: suppressInputValueChange,
							Sensitive:        true,
						},
						"secret": {
							Type:        schema.TypeBool,
							Description: "True if the input value is a secret",
							Computed:    true,
						},
						"checksum": {
							Type:        schema.TypeString,
							Description: "SHA-256 checksum of the input value",
							Computed:    true,
						},
					},
				},
			},
			"template_id": {
				Type:        schema.TypeString,
				Description: "ID of the template this deployment belongs to",
				Computed:    true,
			},
			"deployment_id": {
				Type:        schema.TypeString,
				Description: "ID (slug) of the deployment",
				Computed:    true,
			},
			"state": {
				Type:        schema.TypeString,
				Description: "Current state of the deployment",
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeInt,
				Description: "Unix timestamp when the deployment was created",
				Computed:    true,
			},
			"stacks": {
				Type:        schema.TypeList,
				Description: "List of stacks IDs created by the deployment",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "ID of the stack",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func resourceTemplateDeploymentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		TemplateDeployment structs.TemplateDeployment `graphql:"blueprintDeploymentCreate(id: $id, input: $input)"`
	}

	variables := map[string]interface{}{
		"id":    toID(d.Get("template_version_id")),
		"input": templateDeploymentCreateInput(d),
	}

	client := meta.(*internal.Client)

	if err := client.Mutate(ctx, "BlueprintDeploymentCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create template deployment: %v", err)
	}

	deployment := mutation.TemplateDeployment
	d.SetId(fmt.Sprintf("%s/%s", deployment.Template.ID, deployment.ID))
	d.Set("deployment_id", deployment.ID)
	d.Set("template_id", deployment.Template.ID)

	if diags := waitForDeployment(ctx, client, d, func(conf *retry.StateChangeConf) {
		conf.Timeout = d.Timeout(schema.TimeoutCreate)
	}); diags.HasError() {
		return diags
	}

	return resourceTemplateDeploymentRead(ctx, d, meta)
}

func resourceTemplateDeploymentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	deploymentID, templateID, err := parseDeploymentID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var query struct {
		TemplateDeployment *structs.TemplateDeployment `graphql:"blueprintDeployment(id: $id, blueprintID: $blueprintID)"`
	}

	variables := map[string]interface{}{
		"id":          toID(deploymentID),
		"blueprintID": toID(templateID),
	}

	if err := meta.(*internal.Client).Query(ctx, "BlueprintDeployment", &query, variables); err != nil {
		return diag.Errorf("could not query for template deployment: %v", err)
	}

	if query.TemplateDeployment == nil {
		d.SetId("")
		return nil
	}

	deployment := query.TemplateDeployment

	d.Set("space", deployment.Space.ID)
	d.Set("template_id", deployment.Template.ID)
	d.Set("deployment_id", deployment.ID)
	d.Set("template_version_id", deployment.TemplateVersion.ID)
	d.Set("name", deployment.Name)
	d.Set("state", deployment.State)
	d.Set("created_at", deployment.CreatedAt)

	if deployment.Description != nil {
		d.Set("description", *deployment.Description)
	} else {
		d.Set("description", "")
	}

	inputList := make([]map[string]any, len(deployment.Inputs))
	for i, input := range deployment.Inputs {
		value := input.Value
		if input.Secret {
			value = d.Get(fmt.Sprintf("input.%d.value", i)).(string)
		}
		in := map[string]any{
			"id":       input.ID,
			"value":    value,
			"secret":   input.Secret,
			"checksum": input.Checksum,
		}
		inputList[i] = in
	}
	if err := d.Set("input", inputList); err != nil {
		return diag.FromErr(err)
	}

	stacks := make([]map[string]string, len(deployment.Stacks))
	for i, stack := range deployment.Stacks {
		stacks[i] = map[string]string{
			"id": stack.ID,
		}
	}
	d.Set("stacks", stacks)

	return nil
}

func resourceTemplateDeploymentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	deploymentID, templateID, err := parseDeploymentID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var mutation struct {
		TemplateDeployment structs.TemplateDeployment `graphql:"blueprintDeploymentUpdate(id: $id, blueprintID: $blueprintID, input: $input)"`
	}

	inputs := structs.BlueprintDeploymentUpdateInput{}

	if d.HasChange("template_version_id") {
		inputs.VersionID = graphql.NewID(graphql.ID(d.Get("template_version_id").(string)))
	}

	if d.HasChange("description") {
		inputs.Description = graphql.NewString(graphql.String(d.Get("description").(string)))
	}

	if d.HasChange("name") {
		inputs.Name = graphql.NewString(graphql.String(d.Get("name").(string)))
	}

	if d.HasChange("input") {
		inputs.Inputs = templateDeploymentSchemaToInputs(d)
	}

	variables := map[string]any{
		"id":          toID(deploymentID),
		"blueprintID": toID(templateID),
		"input":       inputs,
	}

	client := meta.(*internal.Client)

	if err := client.Mutate(ctx, "BlueprintDeploymentUpdate", &mutation, variables); err != nil {
		return diag.Errorf("could not update template deployment: %v", err)
	}

	// When the version has been changed, a new run is triggered, so we need to wait
	// for the status update.
	if d.HasChange("template_version_id") || d.HasChange("input") {
		if diags := waitForDeployment(ctx, client, d, func(conf *retry.StateChangeConf) {
			conf.Timeout = d.Timeout(schema.TimeoutUpdate)
			// Let's give Spacelift some time to update the deployment state
			// When the state has been updated, it might take time for it's state to change.
			// The following option makes sure we are getting 3 times in a row the finished state.
			// This catches the case where the update does not immediately trigger a state change and so the deployment
			// will still be in the FINISHED state.
			// This is slightly better and more dynamic than setting conf.Delay.
			conf.ContinuousTargetOccurence = 3
		}); diags.HasError() {
			return diags
		}
	}

	return resourceTemplateDeploymentRead(ctx, d, meta)
}

func resourceTemplateDeploymentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		TemplateDeployment *structs.TemplateDeployment `graphql:"blueprintDeploymentDelete(id: $id, blueprintID: $blueprintID)"`
	}

	variables := map[string]interface{}{
		"id":          toID(d.Get("deployment_id").(string)),
		"blueprintID": toID(d.Get("template_id").(string)),
	}

	client := meta.(*internal.Client)

	if err := client.Mutate(ctx, "TemplateDeploymentDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete template deployment: %v", err)
	}

	if diags := waitForDeployment(ctx, client, d, func(conf *retry.StateChangeConf) {
		conf.Timeout = d.Timeout(schema.TimeoutDelete)
		conf.Target = []string{"NOT_FOUND", "UNCONFIRMED"}
	}); diags.HasError() {
		return diags
	}

	d.SetId("")

	return nil
}

func parseDeploymentID(id string) (string, string, error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid template deployment ID format, expected: template_id/deployment_id, got: %s", id)
	}
	return parts[1], parts[0], nil
}

func templateDeploymentCreateInput(d *schema.ResourceData) structs.BlueprintDeploymentCreateInput {
	input := structs.BlueprintDeploymentCreateInput{
		Space:  graphql.ID(d.Get("space").(string)),
		Name:   graphql.String(d.Get("name").(string)),
		Inputs: []structs.BlueprintDeploymentCreateInputPair{},
	}

	if description, ok := d.GetOk("description"); ok {
		input.Description = toOptionalString(description)
	}

	input.Inputs = templateDeploymentSchemaToInputs(d)

	return input
}

func templateDeploymentSchemaToInputs(d *schema.ResourceData) []structs.BlueprintDeploymentCreateInputPair {
	inputs := make([]structs.BlueprintDeploymentCreateInputPair, 0)
	if inputList, ok := d.GetOk("input"); ok {
		stateInputs := inputList.([]any)
		inputs = make([]structs.BlueprintDeploymentCreateInputPair, len(stateInputs))
		for i, v := range stateInputs {
			item := v.(map[string]any)
			inputs[i] = structs.BlueprintDeploymentCreateInputPair{
				ID:    graphql.String(item["id"].(string)),
				Value: graphql.String(item["value"].(string)),
			}
		}
		return inputs
	}
	return inputs
}

type waitInput struct {
	Options []func(conf *retry.StateChangeConf)
}

type deploymentStateQuery struct {
	Deployment *struct {
		State string `graphql:"state"`
	} `graphql:"blueprintDeployment(id: $id, blueprintID: $blueprintID)"`
}

func waitForDeploymentState(ctx context.Context, waitInput waitInput) diag.Diagnostics {
	waitConf := retry.StateChangeConf{
		PollInterval: 3 * time.Second,
		Pending:      []string{"NONE", "IN_PROGRESS", "DESTROYING"},
		Target:       []string{"FINISHED", "UNCONFIRMED"},
	}

	for _, option := range waitInput.Options {
		option(&waitConf)
	}

	if _, err := waitConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("could not wait for template deployment state: %v", err)
	}

	return nil
}

func waitForDeployment(ctx context.Context, client *internal.Client, d *schema.ResourceData, options ...func(conf *retry.StateChangeConf)) diag.Diagnostics {
	deploymentID := d.Get("deployment_id").(string)
	options = append(options, func(conf *retry.StateChangeConf) {
		conf.Refresh = func() (r any, state string, err error) {
			state, err = queryDeploymentState(ctx, client, toID(deploymentID), toID(d.Get("template_id").(string)))
			if err != nil {
				err = errors.Join(err, errors.New("error querying for template deployment state"))
			}
			r = state
			return
		}
	})
	return waitForDeploymentState(ctx, waitInput{
		Options: options,
	})
}

func queryDeploymentState(ctx context.Context, client *internal.Client, id, blueprintID graphql.ID) (string, error) {
	variables := map[string]any{
		"id":          id,
		"blueprintID": blueprintID,
	}
	var query deploymentStateQuery
	if err := client.Query(ctx, "BlueprintDeployment", &query, variables); err != nil {
		return "", errors.Join(err, errors.New("error querying for template deployment state"))
	}
	if query.Deployment == nil {
		return "NOT_FOUND", nil
	}
	return query.Deployment.State, nil
}

// suppressInputValueChange suppresses diff on input value if the config value's checksum
// matches the remote checksum. This allows detecting drift when a secret value is changed
// outside of Terraform. The key format is "input.<index>.value", so we extract the index
// to find the corresponding checksum.
func suppressInputValueChange(k, old, new string, d *schema.ResourceData) bool {
	// Extract the index from the key (e.g., "input.0.value" -> "0")
	parts := strings.Split(k, ".")
	if len(parts) != 3 {
		return false
	}
	index := parts[1]

	isSecret, ok := d.GetOk(fmt.Sprintf("input.%s.secret", index))
	if !ok {
		return false
	}
	if !isSecret.(bool) {
		return new == old
	}

	oldChecksum := d.Get(fmt.Sprintf("input.%s.checksum", index)).(string)

	checksum := sha256.Sum256([]byte(new))
	newChecksum := hex.EncodeToString(checksum[:])

	return oldChecksum == newChecksum
}
