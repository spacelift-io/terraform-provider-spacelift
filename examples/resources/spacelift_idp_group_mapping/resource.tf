# Via policy
resource "spacelift_idp_group_mapping" "devops" {
  name = "devops"
  policy {
    space_id = "root"
    role     = "ADMIN"
  }
  description = "Maps the devops IdP group to the root space with admin role"
}

# Via role attachment
resource "spacelift_idp_group_mapping" "devops" {
  name        = "devops"
  description = "Creates a mapping for the devops IdP group"
}

resource "spacelift_role" "devops" {
  name        = "SpaceAdmin"
  description = "A role that provides full admin access to a space"
  actions     = ["SPACE_ADMIN"]
}

resource "spacelift_role_attachment" "devops" {
  idp_group_mapping = spacelift_idp_group_mapping.devops.id
  role_id           = spacelift_role.devops.id
  space_id          = "root"
}
