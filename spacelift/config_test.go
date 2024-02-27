package spacelift

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

var testConfig struct {
	IPs        []string
	SourceCode struct {
		AzureDevOps struct {
			Default struct {
				Name            string
				ID              string
				UserFacingHost  string
				OrganizationURL string
				WebhookSecret   string
				WebhookURL      string
			}
			SpaceLevel struct {
				Name            string
				ID              string
				Space           string
				UserFacingHost  string
				OrganizationURL string
				WebhookSecret   string
				WebhookURL      string
			}
		}
		BitbucketCloud struct {
			Default struct {
				Name       string
				ID         string
				Username   string
				WebhookURL string
			}
			SpaceLevel struct {
				Name       string
				ID         string
				Space      string
				Username   string
				WebhookURL string
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
			}
		}
		GithubEnterprise struct {
			Default struct {
				Name          string
				ID            string
				APIHost       string
				AppID         string
				WebhookSecret string
				WebhookURL    string
			}
			SpaceLevel struct {
				Name          string
				ID            string
				Space         string
				APIHost       string
				AppID         string
				WebhookSecret string
				WebhookURL    string
			}
		}
		Gitlab struct {
			Default struct {
				Name          string
				ID            string
				APIHost       string
				WebhookSecret string
				WebhookURL    string
			}
			SpaceLevel struct {
				Name          string
				ID            string
				Space         string
				APIHost       string
				WebhookSecret string
				WebhookURL    string
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
