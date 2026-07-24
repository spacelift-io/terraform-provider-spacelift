data "spacelift_repos" "this" {
  space_id = "root"
}

output "repos" {
  value = data.spacelift_repos.this.repos
}
