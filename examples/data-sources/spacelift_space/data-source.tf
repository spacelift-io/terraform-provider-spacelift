data "spacelift_space" "space" {
  space_id = spacelift_space.space.id
}

output "space_description" {
  value = data.spacelift_space.space.description
}