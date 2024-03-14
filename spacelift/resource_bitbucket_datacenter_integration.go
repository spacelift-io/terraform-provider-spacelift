package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

const (
	bitbucketDatacenterAccessToken = "access_token"
)

func resourceBitbucketDatacenterIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_bitbucket_datacenter_integration` represents details of a bitbucket datacenter integration",

		CreateContext: resourceBitbucketDatacenterIntegrationCreate,
		ReadContext:   resourceBitbucketDatacenterIntegrationRead,
		UpdateContext: resourceBitbucketDatacenterIntegrationUpdate,
		DeleteContext: resourceBitbucketDatacenterIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			bitbucketDatacenterID: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration id.",
				Computed:    true,
			},
			bitbucketDatacenterName: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration name",
				Required:    true,
			},
			bitbucketDatacenterIsDefault: {
				Type:        schema.TypeBool,
				Description: "Bitbucket Datacenter integration is default.",
				Required:    true,
			},
			bitbucketDatacenterSpaceID: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration space id. Defaults to `root`.",
				Optional:    true,
				Computed:    true,
			},
			bitbucketDatacenterLabels: {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Description: "Bitbucket Datacenter integration labels",
				Optional:    true,
				Computed:    true,
			},
			bitbucketDatacenterDescription: {
				Type:        schema.TypeString,
				Description: "Bitbucket Datacenter integration description",
				Optional:    true,
				Computed:    true,
			},
			bitbucketDatacenterAPIHost: {
				Type:        schema.TypeString,
				Description: "The API host where requests will be sent",
				Required:    true,
			},
			bitbucketDatacenterUserFacingHost: {
				Type:        schema.TypeString,
				Description: "User Facing Host which will be used for all user-facing URLs displayed in the Spacelift UI",
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
		CreateBitbucketDatacenterIntegration structs.BitbucketDatacenterIntegration `graphql:"bitbucketDatacenterIntegrationCreate(apiHost: $apiHost, userFacingHost: $userFacingHost, username: $username, accessToken: $accessToken, customInput: $customInput)"`
	}

	variables := map[string]interface{}{
		"customInput": &structs.CustomVCSInput{
			Name:        toString(d.Get(bitbucketDatacenterName)),
			IsDefault:   toOptionalBool(d.Get(bitbucketDatacenterIsDefault)),
			SpaceID:     toString(d.Get(bitbucketDatacenterSpaceID)),
			Labels:      toOptionalStringList(d.Get(bitbucketDatacenterLabels)),
			Description: toOptionalString(d.Get(bitbucketDatacenterDescription)),
		},
		"apiHost":        toString(d.Get(bitbucketDatacenterAPIHost)),
		"userFacingHost": toString(d.Get(bitbucketDatacenterUserFacingHost)),
		"username":       toOptionalString(d.Get(bitbucketDatacenterUsername)),
		"accessToken":    toString(d.Get(bitbucketDatacenterAccessToken)),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "BitbucketDatacenterIntegrationCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create the bitbucket datacenter integration: %v", internal.FromSpaceliftError(err))
	}

	fillBitbucketDatacenterIntegrationResults(d, &mutation.CreateBitbucketDatacenterIntegration)

	return nil
}

func resourceBitbucketDatacenterIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		BitbucketDatacenterIntegration *structs.BitbucketDatacenterIntegration `graphql:"bitbucketDatacenterIntegration(id: $id)"`
	}

	variables := map[string]interface{}{"id": d.Id()}
	if err := meta.(*internal.Client).Query(ctx, "BitbucketDatacenterIntegrationRead", &query, variables); err != nil {
		return diag.Errorf("could not query for the bitbucket datacenter integration: %v", err)
	}

	if query.BitbucketDatacenterIntegration == nil {
		d.SetId("")
	} else {
		fillBitbucketDatacenterIntegrationResults(d, query.BitbucketDatacenterIntegration)
	}

	return nil
}

func resourceBitbucketDatacenterIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UpdateBitbucketDatacenterIntegration structs.BitbucketDatacenterIntegration `graphql:"bitbucketDatacenterIntegrationUpdate(apiHost: $apiHost, userFacingHost: $userFacingHost, username: $username, accessToken: $accessToken, customInput: $customInput)"`
	}

	variables := map[string]interface{}{
		"apiHost":        toString(d.Get(bitbucketDatacenterAPIHost)),
		"userFacingHost": toString(d.Get(bitbucketDatacenterUserFacingHost)),
		"username":       toOptionalString(d.Get(bitbucketDatacenterUsername)),
		"accessToken":    toOptionalString(d.Get(bitbucketDatacenterAccessToken)),
		"customInput": &structs.CustomVCSUpdateInput{
			ID:          toID(d.Id()),
			SpaceID:     toString(d.Get(bitbucketDatacenterSpaceID)),
			Description: toOptionalString(d.Get(bitbucketDatacenterDescription)),
			Labels:      toOptionalStringList(d.Get(bitbucketDatacenterLabels)),
		},
	}

	var ret diag.Diagnostics

	if err := meta.(*internal.Client).Mutate(ctx, "BitbucketDatacenterIntegrationUpdate", &mutation, variables); err != nil {
		ret = append(ret, diag.Errorf("could not update the bitbucket datacenter integration: %v", internal.FromSpaceliftError(err))...)
	}

	fillBitbucketDatacenterIntegrationResults(d, &mutation.UpdateBitbucketDatacenterIntegration)

	return ret
}

func resourceBitbucketDatacenterIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteBitbucketDatacenterIntegration *structs.BitbucketDatacenterIntegration `graphql:"bitbucketDatacenterIntegrationDelete(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": toID(d.Id()),
	}

	if err := meta.(*internal.Client).Mutate(ctx, "BitbucketDatacenterIntegrationDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete bitbucket datacenter integration: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}

func fillBitbucketDatacenterIntegrationResults(d *schema.ResourceData, bitbucketDatacenterIntegration *structs.BitbucketDatacenterIntegration) {
	// Note: the access token is not exposed in the API.
	d.SetId(bitbucketDatacenterIntegration.ID)
	d.Set(bitbucketDatacenterAPIHost, bitbucketDatacenterIntegration.APIHost)
	d.Set(bitbucketDatacenterUsername, bitbucketDatacenterIntegration.Username)
	d.Set(bitbucketDatacenterUserFacingHost, bitbucketDatacenterIntegration.UserFacingHost)
	d.Set(bitbucketDatacenterWebhookURL, bitbucketDatacenterIntegration.WebhookURL)
	d.Set(bitbucketDatacenterWebhookSecret, bitbucketDatacenterIntegration.WebhookSecret)
	d.Set(bitbucketDatacenterIsDefault, bitbucketDatacenterIntegration.IsDefault)
	d.Set(bitbucketDatacenterSpaceID, bitbucketDatacenterIntegration.Space.ID)
	d.Set(bitbucketDatacenterName, bitbucketDatacenterIntegration.Name)
	d.Set(bitbucketDatacenterDescription, bitbucketDatacenterIntegration.Description)

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range bitbucketDatacenterIntegration.Labels {
		labels.Add(label)
	}
	d.Set(bitbucketDatacenterLabels, labels)
}
