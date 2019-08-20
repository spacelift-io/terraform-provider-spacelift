package spacelift

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pkg/errors"
)

// Provider returns an instance of Terraform resource provider for Spacelift.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_token": {
				Type:        schema.TypeString,
				Description: "Spacelift administrative token",
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("SPACELIFT_API_TOKEN", nil),
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"spacelift_context_attachment":   dataContextAttachment(),
			"spacelift_context":              dataContext(),
			"spacelift_environment_variable": dataEnvironmentVariable(),
			"spacelift_mounted_file":         dataMountedFile(),
			"spacelift_stack":                dataStack(),
			"spacelift_stack_aws_role":       dataStackAWSRole(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"spacelift_context_attachment":   resourceContextAttachment(),
			"spacelift_context":              resourceContext(),
			"spacelift_environment_variable": resourceEnvironmentVariable(),
			"spacelift_mounted_file":         resourceMountedFile(),
			"spacelift_stack":                resourceStack(),
			"spacelift_stack_aws_role":       resourceStackAWSRole(),
		},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	token := d.Get("api_token").(string)

	var claims jwt.StandardClaims
	if jwt, err := jwt.ParseWithClaims(token, &claims, nil); jwt == nil && err != nil {
		return nil, errors.Wrap(err, "could not parse the API token")
	}

	return &Client{Endpoint: claims.Audience, Token: token}, nil
}
