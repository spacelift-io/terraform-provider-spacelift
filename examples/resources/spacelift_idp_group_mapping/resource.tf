resource "spacelift_idp_group_mapping" "test" {
  name = "test"
  policy {
    space_id = "root"
    role     = "ADMIN"
  }
  description = "test description"
}
