package spacelift

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var (
	provider     terraform.ResourceProvider
	providerLock sync.Mutex
)

func testProvider() terraform.ResourceProvider {
	providerLock.Lock()
	defer providerLock.Unlock()
	if provider == nil {
		provider = Provider()
	}

	return provider
}

func testSteps(t *testing.T, steps []resource.TestStep) {
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers: map[string]terraform.ResourceProvider{
			"spacelift": testProvider(),
		},
		Steps: steps,
	})
}
