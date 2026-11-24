package spacelift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs/search"
)

// There is no search query for integrations, so all are fetched and keep
// filters them.
func NewAzureIntegrationListResource() list.ListResource {
	return &sdkListResource[structs.AzureIntegration]{
		typeName: "spacelift_azure_integration", singular: "Azure integration", plural: "Azure integrations",
		resource: resourceAzureIntegration,
		fetchPage: func(ctx context.Context, client *internal.Client, _ search.SearchInput) ([]structs.AzureIntegration, search.PageInfo, error) {
			var query struct {
				AzureIntegrations []structs.AzureIntegration `graphql:"azureIntegrations"`
			}
			err := client.Query(ctx, "AzureIntegrationList", &query, nil)
			return query.AzureIntegrations, search.PageInfo{}, err
		},
		keep: func(filter *listFilter, integration *structs.AzureIntegration) bool {
			return filter.matches(integration.Name, integration.Space, integration.Labels)
		},
		name:     func(integration *structs.AzureIntegration) string { return integration.Name },
		populate: populateAzureIntegration,
	}
}
