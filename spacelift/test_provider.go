package spacelift

import (
	"sync"

	"github.com/hashicorp/terraform/terraform"
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
