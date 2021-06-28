module github.com/spacelift-io/terraform-provider-spacelift

require (
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/alcortesm/tgz v0.0.0-20161220082320-9c5fe88206d7 // indirect
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.7.0
	github.com/keybase/go-crypto v0.0.0-20161004153544-93f5b35093ba // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/pkg/errors v0.9.1
	github.com/shurcooL/graphql v0.0.0-20181231061246-d48a9a75455f
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
)

go 1.13

replace github.com/shurcooL/graphql => github.com/marcinwyszynski/graphql v0.0.0-20210505073322-ed22d920d37d
