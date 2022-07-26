package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func resourceAWSIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"**Note:** This resource is experimental. Please continue to use `spacelift_aws_role`." +
			"\n\n" +
			"`spacelift_aws_integration` represents an integration with an AWS " +
			"account. This integration is account-level and needs to be explicitly " +
			"attached to individual stacks in order to take effect." +
			"\n\n" +
			"Note: when assuming credentials for **shared workers**, Spacelift will use `$accountName-$integrationID@$stackID-$suffix` " +
			"or `$accountName-$integrationID@$moduleID-$suffix` as [external ID](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html) " +
			"and `$runID@$stackID@$accountName` truncated to 64 characters as [session ID](https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole)," +
			"$suffix will be `read` or `write`.",

		CreateContext: resourceAWSIntegrationCreate,
		ReadContext:   resourceAWSIntegrationRead,
		UpdateContext: resourceAWSIntegrationUpdate,
		DeleteContext: resourceAWSIntegrationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			// Required.
			"name": {
				Type:        schema.TypeString,
				Description: "The friendly name of the integration",
				Required:    true,
			},
			"role_arn": {
				Type:        schema.TypeString,
				Description: "ARN of the AWS IAM role to attach",
				Required:    true,
			},
			"generate_credentials_in_worker": {
				Type:        schema.TypeBool,
				Description: "Generate AWS credentials in the private worker. Defaults to `false`.",
				Optional:    true,
				Default:     false,
			},
			// Optional.
			"external_id": {
				Type:        schema.TypeString,
				Description: "Custom external ID (works only for private workers).",
				Optional:    true,
			},
			"duration_seconds": {
				Type:        schema.TypeInt,
				Description: "Duration in seconds for which the assumed role credentials should be valid. Defaults to `900`.",
				Default:     900,
				Optional:    true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "Labels to set on the integration",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
			"space_id": {
				Type:        schema.TypeString,
				Description: "ID (slug) of the space the integration is in",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func resourceAWSIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreateAWSIntegration structs.AWSIntegration `graphql:"awsIntegrationCreate(name: $name, roleArn: $roleArn, generateCredentialsInWorker: $generateCredentialsInWorker, externalID: $externalID, durationSeconds: $durationSeconds, labels: $labels, space: $space)"`
	}

	labels := []graphql.String{}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}
	}

	variables := map[string]interface{}{
		"name":                        toString(d.Get("name")),
		"roleArn":                     toString(d.Get("role_arn")),
		"externalID":                  toString(d.Get("external_id")),
		"labels":                      labels,
		"durationSeconds":             graphql.Int(d.Get("duration_seconds").(int)),
		"generateCredentialsInWorker": graphql.Boolean(d.Get("generate_credentials_in_worker").(bool)),
		"space":                       (*graphql.ID)(nil),
	}

	if spaceID, ok := d.GetOk("space_id"); ok {
		variables["space"] = graphql.NewID(spaceID)
	}

	if err := meta.(*internal.Client).Mutate(ctx, "AWSIntegrationCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create AWS integration %v: %v", d.Get("name"), internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.CreateAWSIntegration.ID)

	return resourceAWSIntegrationRead(ctx, d, meta)
}

func resourceAWSIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		AWSIntegration *structs.AWSIntegration `graphql:"awsIntegration(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "AWSIntegrationRead", &query, variables); err != nil {
		return diag.Errorf("could not query for the AWS integration: %v", err)
	}

	if integration := query.AWSIntegration; integration == nil {
		d.SetId("")
	} else {
		integration.PopulateResourceData(d)
	}

	return nil
}

func resourceAWSIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UpdateAWSIntegration structs.AWSIntegration `graphql:"awsIntegrationUpdate(id: $id, name: $name, roleArn: $roleArn, generateCredentialsInWorker: $generateCredentialsInWorker, externalID: $externalID, durationSeconds: $durationSeconds, labels: $labels, space: $space)"`
	}

	labels := []graphql.String{}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}
	}

	variables := map[string]interface{}{
		"id":                          graphql.ID(d.Id()),
		"name":                        toString(d.Get("name")),
		"roleArn":                     toString(d.Get("role_arn")),
		"externalID":                  toString(d.Get("external_id")),
		"labels":                      labels,
		"durationSeconds":             graphql.Int(d.Get("duration_seconds").(int)),
		"generateCredentialsInWorker": graphql.Boolean(d.Get("generate_credentials_in_worker").(bool)),
		"space":                       (*graphql.ID)(nil),
	}

	if spaceID, ok := d.GetOk("space_id"); ok {
		variables["space"] = graphql.NewID(spaceID)
	}

	var ret diag.Diagnostics

	if err := meta.(*internal.Client).Mutate(ctx, "AWSIntegrationUpdate", &mutation, variables); err != nil {
		ret = diag.Errorf("could not update the AWS integration: %v", internal.FromSpaceliftError(err))
	}

	return append(ret, resourceAWSIntegrationRead(ctx, d, meta)...)
}

func resourceAWSIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteAWSIntegration *structs.AWSIntegration `graphql:"awsIntegrationDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "AWSIntegrationDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete the AWS integration: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}
