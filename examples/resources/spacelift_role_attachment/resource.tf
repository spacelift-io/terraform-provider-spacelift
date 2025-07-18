
resource "spacelift_role" "devops" {
  name    = "A role for DevOps team"
  actions = ["SPACE_ADMIN"]
}
resource "spacelift_space" "devops" {
  name            = "DevOps"
  parent_space_id = "root"
}

resource "spacelift_idp_group_mapping" "devops" {
  name = "devops-group"
}

# Attach an API key to a role in a specific space
resource "spacelift_role_attachment" "api_key_attachment" {
  api_key_id = "01K09KERE33P95V40YRWWRVAZT"
  role_id    = spacelift_role.devops.id
  space_id   = spacelift_space.devops.id
}

# Attach an IDP group mapping to a role in a specific space
resource "spacelift_role_attachment" "idp_group_attachment" {
  idp_group_mapping_id = spacelift_idp_group_mapping.devops.id
  role_id              = spacelift_role.devops.id
  space_id             = spacelift_space.devops.id
}
