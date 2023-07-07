package spacelift

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	provider     *schema.Provider
	providerLock sync.Mutex
)

func testProvider() *schema.Provider {
	providerLock.Lock()
	defer providerLock.Unlock()
	if provider == nil {
		provider = Provider("commit", "version")()
	}

	return provider
}

func testSteps(t *testing.T, steps []resource.TestStep) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers: map[string]*schema.Provider{
			"spacelift": testProvider(),
		},
		Steps: steps,
	})
}
