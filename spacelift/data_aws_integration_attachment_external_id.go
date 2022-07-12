package spacelift

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

// TODO(adamc): add examples usages
func dataAWSIntegrationAttachmentExternalID() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_aws_integration_attachment_external_id` is used to generate the external ID " +
			"that would be used for role assumption when an AWS integration is attached to a stack or module.",

		ReadContext: dataAWSIntegrationAttachmentExternalIDRead,

		Schema: map[string]*schema.Schema{
			"integration_id": {
				Type:        schema.TypeString,
				Description: "immutable ID (slug) of the AWS integration",
				Required:    true,
			},
			"stack_id": {
				Type:         schema.TypeString,
				Description:  "immutable ID (slug) of the stack",
				ExactlyOneOf: []string{"module_id", "stack_id"},
				Optional:     true,
			},
			"module_id": {
				Type:         schema.TypeString,
				Description:  "immutable ID (slug) of the module",
				ExactlyOneOf: []string{"module_id", "stack_id"},
				Optional:     true,
			},
			"read": {
				Type:         schema.TypeBool,
				Description:  "whether the integration will be used for read operations",
				AtLeastOneOf: []string{"read", "write"},
				Optional:     true,
			},
			"write": {
				Type:         schema.TypeBool,
				Description:  "whether the integration will be used for write operations",
				AtLeastOneOf: []string{"read", "write"},
				Optional:     true,
			},
			"external_id": {
				Type:        schema.TypeString,
				Description: "The external ID that will be used during role assumption",
				Computed:    true,
			},
			"assume_role_policy_statement": {
				Type:        schema.TypeString,
				Description: "An assume role policy statement that can be attached to your role to allow Spacelift to assume it",
				Computed:    true,
			},
		},
	}
}

func dataAWSIntegrationAttachmentExternalIDRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		Integration *struct {
			ID         string `graphql:"id"`
			IDForStack struct {
				ExternalID                string `graphql:"externalId"`
				AssumeRolePolicyStatement string `graphql:"assumeRolePolicyStatement"`
			} `graphql:"externalIdForStack(stackId: $stackId, read: $read, write: $write)"`
		} `graphql:"awsIntegration(id: $id)"`
	}

	id := toID(d.Get("integration_id"))
	projectID := projectID(d)
	read := d.Get("read").(bool)
	write := d.Get("write").(bool)

	if !read && !write {
		return diag.Errorf("at least one of either 'read' or 'write' must be true")
	}

	variables := map[string]interface{}{
		"id":      id,
		"stackId": projectID,
		"read":    graphql.Boolean(read),
		"write":   graphql.Boolean(write),
	}

	if err := meta.(*internal.Client).Query(ctx, "AWSIntegrationAttachmentExternalIDRead", &query, variables); err != nil {
		return diag.Errorf("could not query external ID for AWS integration attachment: %v", err)
	}

	integration := query.Integration
	if integration == nil {
		return diag.Errorf("AWS integration %q not found", id)
	}

	accessLevel := "write"
	if !write {
		accessLevel = "read"
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", integration.ID, projectID, accessLevel))
	d.Set("external_id", integration.IDForStack.ExternalID)
	d.Set("assume_role_policy_statement", integration.IDForStack.AssumeRolePolicyStatement)

	return nil
}
