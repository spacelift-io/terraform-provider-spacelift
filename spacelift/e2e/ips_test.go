package e2e

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type IPsTest struct {
	ResourceTest
}

func (e *IPsTest) TestLifecycle_OK() {
	defer gock.Off()

	e.posts( // Mocking out the IPs query.
		`{"query":"{outgoingIPAddresses}"}`,
		`{"data":{"outgoingIPAddresses":["1.2.3.4","5.6.7.8"]}}`,
		5,
	)

	resource.Test(e.T(), resource.TestCase{
		Providers:  e.providers,
		IsUnitTest: true,
		Steps: []resource.TestStep{{
			Config: `data "spacelift_ips" "this" {}`,
			Check: resource.ComposeTestCheckFunc(
				// Test data.
				resource.TestCheckResourceAttr("data.spacelift_ips.this", "ips.#", "2"),
				resource.TestCheckResourceAttr("data.spacelift_ips.this", "ips.1592319998", "1.2.3.4"),
				resource.TestCheckResourceAttr("data.spacelift_ips.this", "ips.1659128649", "5.6.7.8"),
			),
		}},
	})
}

func TestIPs(t *testing.T) {
	suite.Run(t, new(IPsTest))
}
