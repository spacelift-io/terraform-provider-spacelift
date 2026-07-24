data "spacelift_repo" "this" {
  repo_id = "my-repo"
}

# Attach a stack to the repo. Spacelift Repos have no branches, so the branch is
# always "main" and the stack tracks the latest commit.
resource "spacelift_stack" "this" {
  name       = "my-stack"
  repository = data.spacelift_repo.this.repo_id
  branch     = "main"
  space_id   = data.spacelift_repo.this.space_id

  spacelift {
    id = data.spacelift_repo.this.repo_id
  }
}
