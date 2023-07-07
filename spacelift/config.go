package spacelift

import (
	"os"
	"strings"

	"github.com/pkg/errors"
)

type Config struct {
	APIKeyEndpoint string
	APIKeyID       string
	APIKeySecret   string

	APIToken string

	UseAPIKey bool
}

func (c *Config) OverwriteWithEnvironmentWhenNotSet() {
	if c.APIKeyEndpoint == "" {
		c.APIKeyEndpoint = os.Getenv("SPACELIFT_API_KEY_ENDPOINT")
	}
	if c.APIKeyID == "" {
		c.APIKeyID = os.Getenv("SPACELIFT_API_KEY_ID")
	}
	if c.APIKeySecret == "" {
		c.APIKeySecret = os.Getenv("SPACELIFT_API_KEY_SECRET")
	}
	if c.APIToken == "" {
		c.APIToken = os.Getenv("SPACELIFT_API_TOKEN")
	}
}

func (c *Config) Validate() error {
	// Scenario 1: full API key config has been provided, so it takes precedence.
	if c.APIKeyEndpoint != "" && c.APIKeyID != "" && c.APIKeySecret != "" {
		c.UseAPIKey = true
		return nil
	}

	// Scenario 2: the API token is provided.
	if c.APIToken != "" {
		c.UseAPIKey = false
		return nil
	}

	var missing []string
	if c.APIKeyEndpoint == "" {
		missing = append(missing, "api_key_endpoint (`SPACELIFT_API_KEY_ENDPOINT`)")
	}
	if c.APIKeyID == "" {
		missing = append(missing, "api_key_id (`SPACELIFT_API_KEY_ID`)")
	}
	if c.APIKeySecret == "" {
		missing = append(missing, "api_key_secret (`SPACELIFT_API_KEY_SECRET`)")
	}

	// Failure: the API key is not provided, and not all the API key config settings have been provided.
	// This is an error.
	return errors.Errorf(
		"either the api_token (`SPACELIFT_API_TOKEN`) must be set or the following settings must be provided: %s",
		strings.Join(missing, ", "),
	)
}
