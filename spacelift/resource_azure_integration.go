package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func resourceAzureIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "" +
			"`spacelift_azure_integration` represents an integration with an Azure " +
			"AD tenant. This integration is account-level and needs to be explicitly " +
			"attached to individual stacks in order to take effect. Note that you will " +
			"need to provide admin consent manually for the integration to work",
		CreateContext: resourceAzureIntegrationCreate,
		ReadContext:   resourceAzureIntegrationRead,
		UpdateContext: resourceAzureIntegrationUpdate,
		DeleteContext: resourceAzureIntegrationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			// Required.
			"name": {
				Type:             schema.TypeString,
				Description:      "The friendly name of the integration",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"tenant_id": {
				Type:             schema.TypeString,
				Description:      "The Azure AD tenant ID",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			// Optional.
			"default_subscription_id": {
				Type: schema.TypeString,
				Description: "" +
					"The default subscription ID to use, if one isn't specified " +
					"at the stack/module level",
				Optional: true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Description: "Labels to set on the integration",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validations.DisallowEmptyString,
				},
				Optional: true,
			},
			// Read-only.
			"admin_consent_provided": {
				Type: schema.TypeBool,
				Description: "" +
					"Indicates whether admin consent has been performed for the " +
					"AAD Application.",
				Computed: true,
			},
			"admin_consent_url": {
				Type: schema.TypeString,
				Description: "" +
					"The URL to use to provide admin consent to the application in " +
					"the customer's tenant",
				Computed: true,
			},
			"application_id": {
				Type: schema.TypeString,
				Description: "" +
					"The applicationId of the Azure AD application used by the " +
					"integration.",
				Computed: true,
			},
			"object_id": {
				Type:        schema.TypeString,
				Description: "The objectId of the Azure AD application used by the integration.",
				Computed:    true,
				Deprecated:  "This field will be removed in a future version. Use `service_principal_object_id` instead.",
			},
			"service_principal_object_id": {
				Type: schema.TypeString,
				Description: "This is the unique ID of the service principal object associated with this application. " +
					"This ID can be useful when performing management operations against this application using programmatic interfaces.",
				Computed: true,
			},
			"display_name": {
				Type: schema.TypeString,
				Description: "" +
					"The display name for the application in Azure. This is " +
					"automatically generated when the integration is created, and " +
					"cannot be changed without deleting and recreating the " +
					"integration.",
				Computed: true,
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

func resourceAzureIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		CreateAzureIntegration structs.AzureIntegration `graphql:"azureIntegrationCreate(name: $name, tenantID: $tenantID, labels: $labels, defaultSubscriptionId: $defaultSubscriptionId, space: $space)"`
	}

	labels := []graphql.String{}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}
	}

	variables := map[string]interface{}{
		"name":                  toString(d.Get("name")),
		"tenantID":              toString(d.Get("tenant_id")),
		"labels":                labels,
		"defaultSubscriptionId": (*graphql.String)(nil),
		"space":                 (*graphql.ID)(nil),
	}

	if spaceID, ok := d.GetOk("space_id"); ok {
		variables["space"] = graphql.NewID(spaceID)
	}

	if defaultSubscriptionID, ok := d.GetOk("default_subscription_id"); ok {
		variables["defaultSubscriptionId"] = toOptionalString(defaultSubscriptionID)
	}

	if err := meta.(*internal.Client).Mutate(ctx, "AzureIntegrationCreate", &mutation, variables); err != nil {
		return diag.Errorf("could not create Azure integration %v: %v", d.Get("name"), internal.FromSpaceliftError(err))
	}

	d.SetId(mutation.CreateAzureIntegration.ID)

	return resourceAzureIntegrationRead(ctx, d, meta)
}

func resourceAzureIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var query struct {
		AzureIntegration *structs.AzureIntegration `graphql:"azureIntegration(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(d.Id())}
	if err := meta.(*internal.Client).Query(ctx, "AzureIntegrationRead", &query, variables); err != nil {
		return diag.Errorf("could not query for the Azure integration: %v", err)
	}

	if integration := query.AzureIntegration; integration == nil {
		d.SetId("")
	} else {
		integration.PopulateResourceData(d)
	}

	return nil
}

func resourceAzureIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		UpdateAzureIntegration structs.AzureIntegration `graphql:"azureIntegrationUpdate(id: $id, name: $name, labels: $labels, defaultSubscriptionId: $defaultSubscriptionId, space: $space)"`
	}

	labels := []graphql.String{}

	if labelSet, ok := d.Get("labels").(*schema.Set); ok {
		for _, label := range labelSet.List() {
			labels = append(labels, graphql.String(label.(string)))
		}
	}

	variables := map[string]interface{}{
		"id":                    graphql.ID(d.Id()),
		"name":                  toString(d.Get("name")),
		"labels":                labels,
		"defaultSubscriptionId": (*graphql.String)(nil),
		"space":                 (*graphql.ID)(nil),
	}

	if subID, ok := d.GetOk("default_subscription_id"); ok {
		variables["defaultSubscriptionId"] = toOptionalString(subID)
	}

	if spaceID, ok := d.GetOk("space_id"); ok {
		variables["space"] = graphql.NewID(spaceID)
	}

	var ret diag.Diagnostics

	if err := meta.(*internal.Client).Mutate(ctx, "AzureIntegrationUpdate", &mutation, variables); err != nil {
		ret = diag.Errorf("could not update the Azure integration: %v", internal.FromSpaceliftError(err))
	}

	return append(ret, resourceAzureIntegrationRead(ctx, d, meta)...)
}

func resourceAzureIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var mutation struct {
		DeleteAzureIntegration *structs.AzureIntegration `graphql:"azureIntegrationDelete(id: $id)"`
	}

	variables := map[string]interface{}{"id": toID(d.Id())}

	if err := meta.(*internal.Client).Mutate(ctx, "AzureIntegrationDelete", &mutation, variables); err != nil {
		return diag.Errorf("could not delete the Azure integration: %v", internal.FromSpaceliftError(err))
	}

	d.SetId("")

	return nil
}
