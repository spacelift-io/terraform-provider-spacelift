# Secret API key
resource "spacelift_api_key" "ci" {
  name       = "CI Pipeline Key"
  idp_groups = ["developers"]
}

# OIDC API key
resource "spacelift_api_key" "github_actions" {
  name       = "GitHub Actions"
  idp_groups = ["ci-runners"]

  oidc {
    issuer             = "https://token.actions.githubusercontent.com"
    client_id          = "spacelift"
    subject_expression = "repo:my-org/*:ref:refs/heads/main"
  }
}

# OIDC API key with claim mappings for dynamic team membership
resource "spacelift_api_key" "backstage" {
  name = "Backstage Integration"

  oidc {
    issuer             = "https://accounts.google.com"
    client_id          = "backstage-client-id"
    subject_expression = "service-account:backstage@my-project.iam.gserviceaccount.com"

    claim_mappings = {
      teams = "groups"
    }
  }
}
