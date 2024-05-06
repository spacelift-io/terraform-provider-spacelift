package spacelift

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestGitLabIntegrationResource(t *testing.T) {
	const resourceName = "spacelift_gitlab_integration.test"

	t.Run("creates and updates a bitbucket datacenter integration without an error", func(t *testing.T) {
		random := func() string { return acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum) }

		var (
			name   = "my-test-gitlab-integration-" + random()
			host   = "https://gitlab.com/" + random()
			token  = "access-" + random()
			descr  = "description " + random()
			labels = `["label1", "label2"]`
		)

		configGitLab := func(host, token, descr, labels string) string {
			return `
				resource "spacelift_gitlab_integration" "test" {
					name              = "` + name + `"
					api_host          = "` + host + `"
					user_facing_host  = "` + host + `"
					token             = "` + token + `"
					description       = "` + descr + `"
					labels            = ` + labels + `
				}
			`
		}

		configStack := func() string {
			return `
				resource "spacelift_worker_pool" "test" {
					name      = "Let's create a dummy worker pool to avoid running the job ` + name + `"
					space_id  = "` + testConfig.SourceCode.Gitlab.SpaceLevel.Space + `"
				}

				resource "spacelift_stack" "test" {
					name            = "stack-for-` + name + `"
					repository      = "` + testConfig.SourceCode.Gitlab.Repository.Name + `"
					branch          = "` + testConfig.SourceCode.Gitlab.Repository.Branch + `"
					space_id        = "` + testConfig.SourceCode.Gitlab.SpaceLevel.Space + `"
					administrative  = false
					worker_pool_id  = spacelift_worker_pool.test.id
					gitlab {
						namespace = "` + testConfig.SourceCode.Gitlab.Repository.Namespace + `"
						id = spacelift_gitlab_integration.test.id
					}
				}
			`
		}

		configRun := func() string {
			return `
				resource "spacelift_run" "test" {
					stack_id = spacelift_stack.test.id
					keepers = { "branch" = "` + testConfig.SourceCode.Gitlab.Repository.Branch + `" }
				}
			`
		}

		spaceLevel := testConfig.SourceCode.Gitlab.SpaceLevel

		testSteps(t, []resource.TestStep{
			{
				Config: configGitLab(host, token, descr, "null"),
				Check: Resource(
					resourceName,
					Attribute(gitLabID, Equals(name)),
					Attribute(gitLabID, Equals(host)),
					Attribute(gitLabUserFacingHost, Equals(host)),
					Attribute(gitLabToken, Equals(token)),
					Attribute(gitLabWebhookURL, IsNotEmpty()),
					Attribute(gitLabWebhookSecret, IsNotEmpty()),
					Attribute(gitLabIsDefault, Equals("false")),
					Attribute(gitLabDescription, Equals(descr)),
					Attribute("labels.#", Equals("0")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{gitLabToken}, // specified only in the config
			},
			{
				Config: configGitLab(host, token, "new descr", `["new label1"]`),
				Check: Resource(
					resourceName,
					Attribute(gitLabAPIHost, Equals(host)),
					Attribute(gitLabUserFacingHost, Equals(host)),
					Attribute(gitLabIsDefault, Equals("false")),
					Attribute(gitLabDescription, Equals("new descr")),
					Attribute(gitLabLabels+".#", Equals("1")),
					Attribute(gitLabLabels+".0", Equals("new label1")),
				),
			},
			{
				Config: configGitLab(spaceLevel.APIHost, spaceLevel.Token, descr, labels),
				Check: Resource(
					resourceName,
					Attribute(gitLabAPIHost, Equals(spaceLevel.APIHost)),
					Attribute(gitLabUserFacingHost, Equals(spaceLevel.APIHost)),
					Attribute(gitLabIsDefault, Equals("false")),
					Attribute(gitLabToken, Equals(spaceLevel.Token)),
					Attribute(gitLabDescription, Equals(descr)),
					Attribute(gitLabLabels+".#", Equals("2")),
					Attribute(gitLabLabels+".0", Equals("label1")),
					Attribute(gitLabLabels+".1", Equals("label2")),
				),
			},
			{
				Config: configGitLab(spaceLevel.APIHost, spaceLevel.Token, descr, labels) + configStack(),
				Check: Resource(
					"spacelift_stack.test",
					Attribute("gitlab.0.id", Equals(name)),
				),
			},
			{
				Config: configGitLab(spaceLevel.APIHost, spaceLevel.Token, descr, labels) + configStack() + configRun(),
				Check: Resource(
					"spacelift_run.test",
					Attribute(gitLabID, IsNotEmpty()),
					Attribute("stack_id", Equals("stack-for-"+name)),
				),
			},
		})
	})
}
