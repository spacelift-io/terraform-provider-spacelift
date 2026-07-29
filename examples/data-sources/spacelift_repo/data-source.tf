data "spacelift_repo" "this" {
  repo_id = "my-repo"
}

output "repo_space" {
  value = data.spacelift_repo.this.space_id
}
