# Retrieve a system role by name
data "spacelift_role" "admin" {
  name = "Space admin"
}

# Retrieve a custom (non-system) role by ID
data "spacelift_role" "custom" {
  role_id = spacelift_role.custom_role.id
}
