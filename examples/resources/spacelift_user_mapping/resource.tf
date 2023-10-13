resource "spacelift_user_mapping" "test" {
  invitation_email = "johnk@example.com"
  username         = "johnk"
  policy {
    space_id = "root"
    role     = "ADMIN"
  }
  policy {
    space_id = "legacy"
    role     = "READ"
  }
}
