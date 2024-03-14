package spacelift

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestBitbucketDatacenterIntegrationResource(t *testing.T) {
	const resourceName = "spacelift_bitbucket_datacenter_integration.test"

	t.Run("creates and updates a bitbucket datacenter integration without an error", func(t *testing.T) {
		random := func() string { return acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum) }

		var (
			name   = "my-test-bitbucket-datacenter-integration-" + random()
			host   = "https://bitbucket.com/" + random()
			token  = "access-" + random()
			descr  = "description " + random()
			labels = `["label1", "label2"]`
		)

		configBitbucket := func(user, host, token, descr, labels string) string {
			return `
				resource "spacelift_bitbucket_datacenter_integration" "test" {
					name              = "` + name + `"
					is_default        = false
					space_id          = "root"
					api_host          = "` + host + `"
					user_facing_host  = "` + host + `"
					username          = "` + user + `"
					access_token      = "` + token + `"
					description       = "` + descr + `"
					labels            = ` + labels + `
				}
			`
		}

		configStack := func() string {
			return `
				resource "spacelift_worker_pool" "test" {
					name      = "Let's create a dummy worker pool to avoid running the job ` + name + `"
					space_id  = "` + testConfig.SourceCode.BitbucketDatacenter.SpaceLevel.Space + `"
				}

				resource "spacelift_stack" "test" {
					name            = "stack-for-` + name + `"
					repository      = "` + testConfig.SourceCode.BitbucketDatacenter.Repository.Name + `"
					branch          = "` + testConfig.SourceCode.BitbucketDatacenter.Repository.Branch + `"
					space_id        = "` + testConfig.SourceCode.BitbucketDatacenter.SpaceLevel.Space + `"
					administrative  = false
					worker_pool_id  = spacelift_worker_pool.test.id
					bitbucket_datacenter {
						namespace = "` + testConfig.SourceCode.BitbucketDatacenter.Repository.Namespace + `"
						id = spacelift_bitbucket_datacenter_integration.test.id
					}
				}
			`
		}

		configRun := func() string {
			return `
				resource "spacelift_run" "test" {
					stack_id = spacelift_stack.test.id
					keepers = { "branch" = "` + testConfig.SourceCode.BitbucketDatacenter.Repository.Branch + `" }
				}
			`
		}

		spaceLevel := testConfig.SourceCode.BitbucketDatacenter.SpaceLevel

		testSteps(t, []resource.TestStep{
			{
				Config: configBitbucket("username", host, token, descr, "null"),
				Check: Resource(
					resourceName,
					Attribute("id", Equals(name)),
					Attribute("api_host", Equals(host)),
					Attribute("user_facing_host", Equals(host)),
					Attribute("username", Equals("username")),
					Attribute("access_token", Equals(token)),
					Attribute("webhook_url", IsNotEmpty()),
					Attribute("webhook_secret", IsNotEmpty()),
					Attribute("is_default", Equals("false")),
					Attribute("description", Equals(descr)),
					Attribute("labels.#", Equals("0")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"access_token"}, // specified only in the config
			},
			{
				Config: configBitbucket("newUserName", host, token, "new descr", `["new label1"]`),
				Check: Resource(
					resourceName,
					Attribute("api_host", Equals(host)),
					Attribute("user_facing_host", Equals(host)),
					Attribute("is_default", Equals("false")),
					Attribute("username", Equals("newUserName")),
					Attribute("description", Equals("new descr")),
					Attribute("labels.#", Equals("1")),
					Attribute("labels.0", Equals("new label1")),
				),
			},
			{
				Config: configBitbucket(spaceLevel.Username, spaceLevel.APIHost, spaceLevel.AccessToken, descr, labels),
				Check: Resource(
					resourceName,
					Attribute("api_host", Equals(spaceLevel.APIHost)),
					Attribute("user_facing_host", Equals(spaceLevel.APIHost)),
					Attribute("is_default", Equals("false")),
					Attribute("username", Equals(spaceLevel.Username)),
					Attribute("description", Equals(descr)),
					Attribute("labels.#", Equals("2")),
					Attribute("labels.0", Equals("label1")),
					Attribute("labels.1", Equals("label2")),
				),
			},
			{
				Config: configBitbucket(spaceLevel.Username, spaceLevel.APIHost, spaceLevel.AccessToken, descr, labels) + configStack(),
				Check: Resource(
					"spacelift_stack.test",
					Attribute("bitbucket_datacenter.0.id", Equals(name)),
				),
			},
			{
				Config: configBitbucket(spaceLevel.Username, spaceLevel.APIHost, spaceLevel.AccessToken, descr, labels) + configStack() + configRun(),
				Check: Resource(
					"spacelift_run.test",
					Attribute("id", IsNotEmpty()),
					Attribute("stack_id", Equals("stack-for-"+name)),
				),
			},
		})
	})
}
