resource "spacelift_repo" "this" {
  name        = "my-repo"
  space_id    = "root"
  description = "Infrastructure code kept inside Spacelift"
  labels      = ["terraform"]
}

resource "spacelift_repo_file" "main" {
  repo_id = spacelift_repo.this.id
  path    = "main.tf"
  content = <<-EOT
    resource "random_pet" "this" {}
  EOT
}

# A stack cannot attach to a repo with no commits, so it has to wait for the
# first file. Spacelift Repos have no branches, so the branch is always "main".
resource "spacelift_stack" "this" {
  name       = "my-stack"
  repository = spacelift_repo.this.id
  branch     = "main"
  space_id   = spacelift_repo.this.space_id

  spacelift {
    id = spacelift_repo.this.id
  }

  depends_on = [spacelift_repo_file.main]
}
