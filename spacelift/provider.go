package spacelift

import (
	"context"
	"fmt"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
)

// Provider returns an instance of Terraform resource provider for Spacelift.
func Provider() *schema.Provider {
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
			"spacelift_aws_role":                  dataAWSRole(),
			"spacelift_context_attachment":        dataContextAttachment(),
			"spacelift_context":                   dataContext(),
			"spacelift_current_stack":             dataCurrentStack(),
			"spacelift_environment_variable":      dataEnvironmentVariable(),
			"spacelift_gcp_service_account":       dataGCPServiceAccount(),
			"spacelift_ips":                       dataIPs(),
			"spacelift_module":                    dataModule(),
			"spacelift_mounted_file":              dataMountedFile(),
			"spacelift_policy":                    dataPolicy(),
			"spacelift_stack":                     dataStack(),
			"spacelift_webhook":                   dataWebhook(),
			"spacelift_stack_aws_role":            dataStackAWSRole(),           // deprecated
			"spacelift_stack_gcp_service_account": dataStackGCPServiceAccount(), // deprecated
			"spacelift_worker_pool":               dataWorkerPool(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"spacelift_aws_role":                  resourceAWSRole(),
			"spacelift_context_attachment":        resourceContextAttachment(),
			"spacelift_context":                   resourceContext(),
			"spacelift_environment_variable":      resourceEnvironmentVariable(),
			"spacelift_gcp_service_account":       resourceGCPServiceAccount(),
			"spacelift_module":                    resourceModule(),
			"spacelift_mounted_file":              resourceMountedFile(),
			"spacelift_policy_attachment":         resourcePolicyAttachment(),
			"spacelift_policy":                    resourcePolicy(),
			"spacelift_stack":                     resourceStack(),
			"spacelift_stack_aws_role":            resourceStackAWSRole(),           // deprecated
			"spacelift_stack_gcp_service_account": resourceStackGCPServiceAccount(), // deprecated
			"spacelift_webhook":                   resourceWebhook(),
			"spacelift_worker_pool":               resourceWorkerPool(),
		},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	if useAPIKey, err := validateProviderConfig(d); err != nil {
		return nil, err
	} else if useAPIKey {
		return buildClientFromAPIKeyData(d)
	}

	return buildClientFromToken(d.Get("api_token").(string))
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

func buildClientFromToken(token string) (interface{}, error) {
	var claims jwt.StandardClaims
	if jwt, err := jwt.ParseWithClaims(token, &claims, nil); jwt == nil && err != nil {
		return nil, errors.Wrap(err, "could not parse the API token")
	}

	return &internal.Client{Endpoint: claims.Audience, Token: token}, nil
}

func buildClientFromAPIKeyData(d *schema.ResourceData) (interface{}, error) {
	// Since validation runs first, we can safely assume that the data is there.
	endpoint := fmt.Sprintf("%s/graphql", d.Get("api_key_endpoint").(string))
	apiKeyID := d.Get("api_key_id").(string)
	apiKeySecret := d.Get("api_key_secret").(string)

	rawClient := graphql.NewClient(endpoint, nil)

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
