module github.com/spacelift-io/terraform-provider-spacelift

require (
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/hashicorp/go-retryablehttp v0.7.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.16.0
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/pkg/errors v0.9.1
	github.com/shurcooL/graphql v0.0.0-20200928012149-18c5c3165e3a
	golang.org/x/net v0.0.0-20211206223403-eba003a116a9 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac
)

go 1.13

replace github.com/shurcooL/graphql => github.com/marcinwyszynski/graphql v0.0.0-20210505073322-ed22d920d37d
