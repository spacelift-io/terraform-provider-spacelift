# Via policy
resource "spacelift_user" "devops" {
  invitation_email = "devops@example.com"
  username         = "devops"
  policy {
    space_id = "legacy"
    role     = "READ"
  }
  policy {
    space_id = "root"
    role     = "ADMIN"
  }
}

# Via role attachment
resource "spacelift_user" "devops" {
  invitation_email = "devops@example.com"
  username         = "devops"
}

resource "spacelift_role" "devops" {
  name        = "SpaceAdmin"
  description = "A role that provides full admin access to a space"
  actions     = ["SPACE_ADMIN"]
}

resource "spacelift_role_attachment" "user_attachment" {
  user_id  = spacelift_user.devops.id
  role_id  = spacelift_role.devops.id
  space_id = "root"
}
