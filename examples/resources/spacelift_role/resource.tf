resource "spacelift_role" "readonly" {
  name        = "ReadOnly Role"
  description = "A role that can read Space resources and confirm runs"
  actions     = ["SPACE_READ", "RUN_CONFIRM"]
}
