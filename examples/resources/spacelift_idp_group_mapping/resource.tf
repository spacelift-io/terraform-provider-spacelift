resource "spacelift_idp_group_mapping" "test" {
  name = "test"
  policy {
    space_id = "root"
    role     = "ADMIN"
  }
  policy {
    space_id = "legacy"
    role     = "ADMIN"
  }
}
