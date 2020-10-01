package e2e

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift"
)

type ResourceTest struct {
	suite.Suite

	endpoint  string
	token     string
	providers map[string]terraform.ResourceProvider
}

func (r *ResourceTest) SetupTest() {
	r.endpoint = "https://bacon.org"

	claims := jwt.StandardClaims{Audience: r.endpoint}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(nil))
	r.NoError(err)

	r.token = signed
	r.NoError(os.Setenv("SPACELIFT_API_TOKEN", r.token))

	r.providers = map[string]terraform.ResourceProvider{
		"spacelift": spacelift.Provider(),
	}
}

func (r *ResourceTest) peekBody() {
	gock.
		New(r.endpoint).
		Post("/graphql").
		AddMatcher(func(req *http.Request, _ *gock.Request) (bool, error) {
			if req.Body != nil {
				data, err := ioutil.ReadAll(req.Body)
				r.NoError(err)
				r.Empty(string(data))
			}

			return true, nil
		})
}

func (r *ResourceTest) posts(request, response string, times int) {
	gock.
		New(r.endpoint).
		Post("/graphql").
		Times(times).
		BodyString(request).
		Reply(http.StatusOK).
		BodyString(response)
}

func (r *ResourceTest) testsResource(steps []resource.TestStep) {
	resource.Test(r.T(), resource.TestCase{
		IsUnitTest: true,
		Providers:  r.providers,
		Steps:      steps,
	})
}
