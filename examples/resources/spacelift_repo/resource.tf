resource "spacelift_repo" "this" {
  name        = "my-repo"
  space_id    = "root"
  description = "Infrastructure code kept inside Spacelift"
  labels      = ["terraform"]
}
