package spacelift

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

// Provider returns an instance of Terraform resource provider for Spacelift.
func Provider(commit, version string) plugin.ProviderFunc {
	return func() *schema.Provider {
		return &schema.Provider{
			Schema: map[string]*schema.Schema{
				"api_key_endpoint": {
					Type:        schema.TypeString,
					Description: "Endpoint to use when authenticating with an API key outside of Spacelift",
					DefaultFunc: schema.EnvDefaultFunc("SPACELIFT_API_KEY_ENDPOINT", nil),
					Optional:    true,
					Sensitive:   false,
				},
				"api_key_id": {
					Type:        schema.TypeString,
					Description: "ID of the API key to use when executing outside of Spacelift",
					DefaultFunc: schema.EnvDefaultFunc("SPACELIFT_API_KEY_ID", nil),
					Optional:    true,
					Sensitive:   false,
				},
				"api_key_secret": {
					Type:        schema.TypeString,
					Description: "API key secret to use when executing outside of Spacelift",
					DefaultFunc: schema.EnvDefaultFunc("SPACELIFT_API_KEY_SECRET", nil),
					Optional:    true,
					Sensitive:   true,
				},
				"api_token": {
					Type:        schema.TypeString,
					Description: "Spacelift token generated by a run, only useful from within Spacelift",
					DefaultFunc: schema.EnvDefaultFunc("SPACELIFT_API_TOKEN", nil),
					Optional:    true,
					Sensitive:   true,
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"spacelift_aws_role":                         dataAWSRole(),
				"spacelift_azure_devops_integration":         dataAzureDevopsIntegration(),
				"spacelift_bitbucket_cloud_integration":      dataBitbucketCloudIntegration(),
				"spacelift_bitbucket_datacenter_integration": dataBitbucketDatacenterIntegration(),
				"spacelift_context_attachment":               dataContextAttachment(),
				"spacelift_context":                          dataContext(),
				"spacelift_current_stack":                    dataCurrentStack(),
				"spacelift_drift_detection":                  dataDriftDetection(),
				"spacelift_environment_variable":             dataEnvironmentVariable(),
				"spacelift_gcp_service_account":              dataGCPServiceAccount(),
				"spacelift_github_enterprise_integration":    dataGithubEnterpriseIntegration(),
				"spacelift_gitlab_integration":               dataGitlabIntegration(),
				"spacelift_ips":                              dataIPs(),
				"spacelift_module":                           dataModule(),
				"spacelift_mounted_file":                     dataMountedFile(),
				"spacelift_policy":                           dataPolicy(),
				"spacelift_stack":                            dataStack(),
				"spacelift_webhook":                          dataWebhook(),
				"spacelift_stack_aws_role":                   dataStackAWSRole(),           // deprecated
				"spacelift_stack_gcp_service_account":        dataStackGCPServiceAccount(), // deprecated
				"spacelift_vcs_agent_pool":                   dataVCSAgentPool(),
				"spacelift_worker_pool":                      dataWorkerPool(),
				"spacelift_worker_pools":                     dataWorkerPools(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"spacelift_aws_role":                  resourceAWSRole(),
				"spacelift_context_attachment":        resourceContextAttachment(),
				"spacelift_context":                   resourceContext(),
				"spacelift_drift_detection":           resourceDriftDetection(),
				"spacelift_environment_variable":      resourceEnvironmentVariable(),
				"spacelift_gcp_service_account":       resourceGCPServiceAccount(),
				"spacelift_module":                    resourceModule(),
				"spacelift_mounted_file":              resourceMountedFile(),
				"spacelift_policy_attachment":         resourcePolicyAttachment(),
				"spacelift_policy":                    resourcePolicy(),
				"spacelift_run":                       resourceRun(),
				"spacelift_stack":                     resourceStack(),
				"spacelift_stack_destructor":          resourceStackDestructor(),
				"spacelift_stack_aws_role":            resourceStackAWSRole(),           // deprecated
				"spacelift_stack_gcp_service_account": resourceStackGCPServiceAccount(), // deprecated
				"spacelift_vcs_agent_pool":            resourceVCSAgentPool(),
				"spacelift_webhook":                   resourceWebhook(),
				"spacelift_worker_pool":               resourceWorkerPool(),
			},
			ConfigureContextFunc: configureProvider(commit, version),
		}
	}
}

func configureProvider(commit, version string) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		var useAPIKey bool
		var client *internal.Client
		var err error

		if useAPIKey, err = validateProviderConfig(d); err != nil {
			return nil, diag.Errorf("could not validate provider config: %v", err)
		} else if useAPIKey {
			client, err = buildClientFromAPIKeyData(d)
		} else {
			client, err = buildClientFromToken(d.Get("api_token").(string))
		}

		if err != nil {
			return nil, diag.Errorf("could not build API client: %v", err)
		}

		if client == nil {
			return nil, diag.Errorf("client not configured")
		}

		client.Commit = commit
		client.Version = version

		return client, nil
	}
}

func validateProviderConfig(d *schema.ResourceData) (bool, error) {
	var missingConfigSettings []string

	for _, config := range []string{"api_key_endpoint", "api_key_id", "api_key_secret"} {
		if _, ok := d.GetOk(config); !ok {
			missingConfigSettings = append(missingConfigSettings, config)
		}
	}

	// Scenario 1: full API key config has been provided, so it takes precedence
	// and we will use it.
	if len(missingConfigSettings) == 0 {
		return true, nil
	}

	// Scenario 2: the API token is provided, so we will use it.
	if _, ok := d.GetOk("api_token"); ok {
		return false, nil
	}

	// Failure: the API key is not provided, and not all of the API key config
	// settings have been provided. This is an error.
	return false, errors.Errorf(
		"either the API key must be set or the following settings must be provided: %s",
		strings.Join(missingConfigSettings, ", "),
	)
}

func buildClientFromToken(token string) (*internal.Client, error) {
	var claims jwt.StandardClaims

	_, _, err := (&jwt.Parser{}).ParseUnverified(token, &claims)
	if unverifiable := new(jwt.UnverfiableTokenError); err != nil && !errors.As(err, &unverifiable) {
		return nil, errors.Wrap(err, "could not parse client token")
	}

	if len(claims.Audience) != 1 {
		return nil, fmt.Errorf("invalid audience in token: %v", claims.Audience)
	}

	requestsPerSecond, maxBurst, err := getRateLimit()
	if err != nil {
		return nil, errors.Wrap(err, "could not create rate limiter for client")
	}

	return internal.NewClient(claims.Audience[0], token, requestsPerSecond, maxBurst), nil
}

func buildClientFromAPIKeyData(d *schema.ResourceData) (*internal.Client, error) {
	// Since validation runs first, we can safely assume that the data is there.
	endpoint := fmt.Sprintf("%s/graphql", d.Get("api_key_endpoint").(string))
	apiKeyID := d.Get("api_key_id").(string)
	apiKeySecret := d.Get("api_key_secret").(string)

	retryableClient := retryablehttp.NewClient()
	retryableClient.Logger = nil

	rawClient := graphql.NewClient(endpoint, retryableClient.StandardClient())

	var mutation struct {
		User *struct {
			Token string `graphql:"jwt"`
		} `graphql:"apiKeyUser(id: $id, secret: $secret)"`
	}

	err := rawClient.Mutate(context.Background(), &mutation, map[string]interface{}{
		"id":     graphql.ID(apiKeyID),
		"secret": graphql.String(apiKeySecret),
	})

	if err != nil {
		return nil, errors.Wrap(err, "could not get API user data")
	}

	if mutation.User == nil {
		return nil, errors.New("no such API user, your key ID may be incorrect")
	}

	return buildClientFromToken(mutation.User.Token)
}

func getRateLimit() (*int, *int, error) {
	maxRequestsPerSecondString := os.Getenv("SPACELIFT_MAX_REQUESTS_PER_SECOND")
	maxRequestBurstString := os.Getenv("SPACELIFT_MAX_REQUESTS_BURST")

	// If the env vars aren't supplied, just default to no limit being applied
	if maxRequestsPerSecondString == "" || maxRequestBurstString == "" {
		return nil, nil, nil
	}

	parsedRate, err := strconv.Atoi(maxRequestsPerSecondString)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse 'SPACELIFT_MAX_REQUESTS_PER_SECOND'")
	}

	parsedBurst, err := strconv.Atoi(maxRequestBurstString)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse 'SPACELIFT_MAX_REQUESTS_BURST'")
	}

	return &parsedRate, &parsedBurst, nil
}
