resource "spacelift_user_mapping" "test" {
  email    = "johnk@eample.com"
  username = "johnk"
  policy {
    space_id = "root"
    role = "ADMIN"
  }
  policy {
    space_id = "legacy"
    role = "READ"
  }
}
