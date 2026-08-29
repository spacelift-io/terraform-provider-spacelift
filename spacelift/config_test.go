package spacelift

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

var testConfig struct {
	IPs []string
	AI  struct {
		// Creating an AI integration needs the space-admin role, so allow
		// pointing the tests at a space other than root.
		Space string `default:"root"`
		// Spacelift does not validate the key against the provider at create
		// time, so one placeholder covers every provider.
		APIKey  string `default:"placeholder-api-key"`
		Bedrock struct {
			// Bedrock cannot be faked: Spacelift assumes the AWS integration's
			// role and validates the inference profiles against AWS. Without
			// these the Bedrock tests skip.
			IntegrationID string
			Region        string `default:"us-east-1"`
			Profile       string
		}
	}
	SourceCode struct {
		AzureDevOps struct {
			Default struct {
				Name                string
				ID                  string
				PersonalAccessToken string
				UserFacingHost      string
				OrganizationURL     string
				WebhookSecret       string
				WebhookURL          string
				VCSChecks           string
				UseGitCheckout      bool
			}
			SpaceLevel struct {
				Name                string
				ID                  string
				Space               string
				PersonalAccessToken string
				UserFacingHost      string
				OrganizationURL     string
				WebhookSecret       string
				WebhookURL          string
				VCSChecks           string
				UseGitCheckout      bool
			}
			Repository struct {
				Name      string
				Namespace string
				Branch    string
			}
		}
		BitbucketCloud struct {
			Default struct {
				Name          string
				ID            string
				Email         string
				WebhookSecret string
				VCSChecks     string
			}
			SpaceLevel struct {
				Name          string
				ID            string
				Space         string
				Email         string
				WebhookSecret string
				VCSChecks     string
			}
			Repository struct {
				Name      string
				Namespace string
				Branch    string
			}
		}
		BitbucketDatacenter struct {
			Default struct {
				Name           string
				ID             string
				Username       string
				UserFacingHost string
				APIHost        string
				WebhookSecret  string
				WebhookURL     string
				VCSChecks      string
				UseGitCheckout bool
			}
			SpaceLevel struct {
				Name           string
				ID             string
				Space          string
				Username       string
				UserFacingHost string
				APIHost        string
				WebhookSecret  string
				WebhookURL     string
				AccessToken    string
				VCSChecks      string
				UseGitCheckout bool
			}
			Repository struct {
				Name      string
				Namespace string
				Branch    string
			}
		}
		GithubEnterprise struct {
			Default struct {
				Name           string
				ID             string
				APIHost        string
				AppID          string
				WebhookSecret  string
				WebhookURL     string
				VCSChecks      string
				UseGitCheckout bool
			}
			SpaceLevel struct {
				Name           string
				ID             string
				Space          string
				APIHost        string
				AppID          string
				WebhookSecret  string
				WebhookURL     string
				VCSChecks      string
				UseGitCheckout bool
			}
			Repository struct {
				Name      string
				Namespace string
				Branch    string
			}
		}
		Gitlab struct {
			Default struct {
				Name           string
				ID             string
				Token          string
				APIHost        string
				WebhookSecret  string
				WebhookURL     string
				VCSChecks      string
				UseGitCheckout bool
			}
			SpaceLevel struct {
				Name           string
				ID             string
				Space          string
				APIHost        string
				Token          string
				WebhookSecret  string
				WebhookURL     string
				VCSChecks      string
				UseGitCheckout bool
			}
			Repository struct {
				Name      string
				Namespace string
				Branch    string
			}
		}
	}
}

func init() {
	err := envconfig.Process("SPACELIFT_PROVIDER_TEST", &testConfig)
	if err != nil {
		log.Fatalln("couldn't process environment variables:", err)
	}
}
