package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

const (
	bitbucketDatacenterUserFacingHost = "user_facing_host"
	bitbucketDatacenterUsername       = "username"
	bitbucketDatacenterAccessToken    = "access_token"
	bitbucketDatacenterAPIHost        = "api_host"
	bitbucketDatacenterWebhookSecret  = "webhook_secret"
	bitbucketDatacenterWebhookURL     = "webhook_url"
)

func resourceBitbucketDatacenterIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_bitbucket_datacenter_integration` represents details of a bitbucket datacenter integration",

		CreateContext: resourceBitbucketDatacenterIntegrationCreate,
		ReadContext:   resourceBitbucketDatacenterIntegrationRead,
		UpdateContext: resourceBitbucketDatacenterIntegrationUpdate,
		DeleteContext: resourceBitbucketDatacenterIntegrationDelete,

		Schema: map[string]*schema.Schema{
			bitbucketDatacenterAPIHost: {
				Type:        schema.TypeString,
				Description: "The API host where requests will be sent",
				Required:    true,
			},
			bitbucketDatacenterUserFacingHost: {
				Type:        schema.TypeString,
				Description: "User Facing Host which will be user for all user-facing URLs displayed in the Spacelift UI",
				Required:    true,
			},
			bitbucketDatacenterUsername: {
				Type:        schema.TypeString,
				Description: "Username which will be used to authenticate requests for cloning repositories",
				Required:    true,
			},
			bitbucketDatacenterAccessToken: {
				Type:        schema.TypeString,
				Description: "User access token from Bitbucket",
				Sensitive:   true,
				Required:    true,
			},
			bitbucketDatacenterWebhookSecret: {
				Type:        schema.TypeString,
				Description: "Secret for webhooks originating from Bitbucket repositories",
				Computed:    true,
				Sensitive:   true,
			},
			bitbucketDatacenterWebhookURL: {
				Type:        schema.TypeString,
				Description: "URL for webhooks originating from Bitbucket repositories",
				Computed:    true,
			},
		},
	}
}

func resourceBitbucketDatacenterIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreateBitbucketDatacenterIntegration structs.BitbucketDatacenterIntegration `graphql:"bitbucketDatacenterIntegrationCreate(apiHost: $apiHost, userFacingHost: $userFacingHost, username: $username, accessToken: $accessToken)"`
	}

	variables := map[string]interface{}{
		"apiHost":        toString(d.Get(bitbucketDatacenterAPIHost)),
		"userFacingHost": toString(d.Get(bitbucketDatacenterUserFacingHost)),
		"username":       toString(d.Get(bitbucketDatacenterUsername)),
		"accessToken":    toString(d.Get(bitbucketDatacenterAccessToken)),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "BitbucketDatacenterIntegrationCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create the bitbucket datacenter integration: %v", internal.FromSpaceliftError(err))
	}

	fillResults(d, &mutation.CreateBitbucketDatacenterIntegration)

	return nil
}

func resourceBitbucketDatacenterIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		BitbucketDatacenterIntegration *structs.BitbucketDatacenterIntegration `graphql:"bitbucketDatacenterIntegration"`
	}

	variables := map[string]interface{}{}
	if err := meta.(*internal.Client).Query(ctx, "BitbucketDatacenterIntegrationRead", &query, variables); err != nil {
		return diag.Errorf("could not query for the bitbucket datacenter integration: %v", err)
	}

	if query.BitbucketDatacenterIntegration == nil {
		d.SetId("")
	} else {
		fillResults(d, query.BitbucketDatacenterIntegration)
	}

	return nil
}

func resourceBitbucketDatacenterIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UpdateBitbucketDatacenterIntegration structs.BitbucketDatacenterIntegration `graphql:"bitbucketDatacenterIntegrationUpdate(apiHost: $apiHost, userFacingHost: $userFacingHost, username: $username, accessToken: $accessToken)"`
	}

	variables := map[string]interface{}{
		"apiHost":        toString(d.Get(bitbucketDatacenterAPIHost)),
		"userFacingHost": toString(d.Get(bitbucketDatacenterUserFacingHost)),
		"username":       toString(d.Get(bitbucketDatacenterUsername)),
		"accessToken":    toString(d.Get(bitbucketDatacenterAccessToken)),
	}

	var ret diag.Diagnostics

	if err := meta.(*internal.Client).Mutate(ctx, "BitbucketDatacenterIntegrationUpdate", &mutation, variables); err != nil {
		ret = append(ret, diag.Errorf("could not update the bitbucket datacenter integration: %v", internal.FromSpaceliftError(err))...)
	}

	fillResults(d, &mutation.UpdateBitbucketDatacenterIntegration)

	return ret
}

func resourceBitbucketDatacenterIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteBitbucketDatacenterIntegration *structs.BitbucketDatacenterIntegration `graphql:"bitbucketDatacenterIntegrationDelete"`
	}

	variables := map[string]interface{}{}

	if err := meta.(*internal.Client).Mutate(ctx, "BitbucketDatacenterIntegrationDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete bitbucket datacenter integration: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func fillResults(d *schema.ResourceData, bitbucketDatacenterIntegration *structs.BitbucketDatacenterIntegration) {
	d.SetId("spacelift_bitbucket_datacenter_integration_id")
	d.Set(bitbucketDatacenterAPIHost, bitbucketDatacenterIntegration.APIHost)
	d.Set(bitbucketDatacenterUsername, bitbucketDatacenterIntegration.Username)
	d.Set(bitbucketDatacenterUserFacingHost, bitbucketDatacenterIntegration.UserFacingHost)
	d.Set(bitbucketDatacenterWebhookURL, bitbucketDatacenterIntegration.WebhookURL)
	d.Set(bitbucketDatacenterWebhookSecret, bitbucketDatacenterIntegration.WebhookSecret)
	// Note: the access token is not exposed in the API.
}
