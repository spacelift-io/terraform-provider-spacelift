package spacelift

import (
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
)

var testConfig struct {
	IPs        []string
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
				Username      string
				WebhookSecret string
				WebhookURL    string
				VCSChecks     string
			}
			SpaceLevel struct {
				Name          string
				ID            string
				Space         string
				Username      string
				WebhookSecret string
				WebhookURL    string
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

func testUsesAPIKeyAuth() bool {
	return os.Getenv("SPACELIFT_API_KEY_ENDPOINT") != "" &&
		os.Getenv("SPACELIFT_API_KEY_ID") != "" &&
		os.Getenv("SPACELIFT_API_KEY_SECRET") != ""
}

func testIsMachineSession() bool {
	if os.Getenv("SPACELIFT_PROVIDER_TEST_MACHINE_SESSION") == "1" {
		return true
	}

	if os.Getenv("SPACELIFT_PROVIDER_TEST_MACHINE_SESSION") == "0" {
		return false
	}

	if testUsesAPIKeyAuth() {
		return false
	}

	return os.Getenv("SPACELIFT_API_TOKEN") != ""
}
