# When a Bitbucket Datacenter server is accessible from the public internet.
resource "spacelift_bitbucket_datacenter_integration" "example" {
  name             = "Bitbucket integration"
  is_default       = false
  space_id         = "root"
  api_host         = "https://mybitbucket.myorg.com"
  user_facing_host = "https://mybitbucket.myorg.com"
  username         = "bitbucket_user_name"
  access_token     = "ABCD-EFGhiJKlMNoPQrSTuVWxYz0123456789abCDefGhiJkL"
}

# When a Bitbucket Datacenter server is not accessible from the public internet.
# We need to use "private://" scheme to reach out our VCS Agent pool.
resource "spacelift_bitbucket_datacenter_integration" "private-example" {
  name             = "Bitbucket integration"
  is_default       = false
  space_id         = "root"
  api_host         = "private://mybitbucket"
  user_facing_host = "https://mybitbucket.myorg.com"
  username         = "bitbucket_user_name"
  access_token     = "ABCD-EFGhiJKlMNoPQrSTuVWxYz0123456789abCDefGhiJkL"
}

resource "spacelift_bitbucket_datacenter_integration" "private-write-only" {
  name                    = "Bitbucket integration"
  is_default              = false
  space_id                = "root"
  api_host                = "https://mybitbucket.myorg.com"
  user_facing_host        = "https://mybitbucket.myorg.com"
  username                = "bitbucket_user_name"
  access_token_wo         = "ABCD-EFGhiJKlMNoPQrSTuVWxYz0123456789abCDefGhiJkL"
  access_token_wo_version = 1
}
