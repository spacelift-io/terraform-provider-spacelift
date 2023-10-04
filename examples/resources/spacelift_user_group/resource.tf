resource "spacelift_user_group" "test" {
  name = "test"
  access {
    space_id = "root"
    level    = "ADMIN"
  }
  access {
    space_id = "legacy"
    level    = "ADMIN"
  }
}
