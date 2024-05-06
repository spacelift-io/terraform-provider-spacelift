resource "spacelift_gitlab_integration" "example" {
  name             = "Bitbucket integration"
  space_id         = "root"
  api_host         = "https://mygitlab.myorg.com"
  user_facing_host = "https://mygitlab.myorg.com"
  username         = "bitbucket_user_name"
  access_token     = "ABCD-EFGhiJKlMNoPQrSTuVWxYz0123456789abCDefGhiJkL"
}

resource "spacelift_gitlab_integration" "private-example" {
  name             = "GitLab Default integration"
  is_default       = true
  api_host         = "https://mygitlab.myorg.com"
  user_facing_host = "https://mybitbucket.myorg.com"
  token            = "gitlab-token"
  access_token     = "ABCD-EFGhiJKlMNoPQrSTuVWxYz0123456789abCDefGhiJkL"
}
