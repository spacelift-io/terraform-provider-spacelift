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
func NewAWSIntegrationListResource() list.ListResource {
	return &sdkListResource[structs.AWSIntegration]{
		typeName: "spacelift_aws_integration", singular: "AWS integration", plural: "AWS integrations",
		resource: resourceAWSIntegration,
		fetchPage: func(ctx context.Context, client *internal.Client, _ search.SearchInput) ([]structs.AWSIntegration, search.PageInfo, error) {
			var query struct {
				AWSIntegrations []structs.AWSIntegration `graphql:"awsIntegrations"`
			}
			err := client.Query(ctx, "AWSIntegrationList", &query, nil)
			return query.AWSIntegrations, search.PageInfo{}, err
		},
		keep: func(filter *listFilter, integration *structs.AWSIntegration) bool {
			return filter.matches(integration.Name, integration.Space, integration.Labels)
		},
		name:     func(integration *structs.AWSIntegration) string { return integration.Name },
		populate: populateAWSIntegration,
	}
}
