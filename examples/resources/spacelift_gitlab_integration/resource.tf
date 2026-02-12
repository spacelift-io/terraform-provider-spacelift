resource "spacelift_gitlab_integration" "example" {
  name             = "GitLab integration (public)"
  space_id         = "root"
  api_host         = "https://mygitlab.myorg.com"
  user_facing_host = "https://mygitlab.myorg.com"
  private_token    = "gitlab-token"
}

resource "spacelift_gitlab_integration" "private-example" {
  name             = "GitLab integration (private)"
  is_default       = true
  api_host         = "private://mygitlab"
  user_facing_host = "https://mygitlab.myorg.com"
  private_token    = "gitlab-token"
}

resource "spacelift_gitlab_integration" "example-write-only" {
  name                     = "GitLab integration (public)"
  space_id                 = "root"
  api_host                 = "https://mygitlab.myorg.com"
  user_facing_host         = "https://mygitlab.myorg.com"
  private_token_wo         = "gitlab-token"
  private_token_wo_version = "gitlab-token"
}
