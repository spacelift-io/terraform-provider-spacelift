data "spacelift_space_by_path" "space" {
  space_path = "root/second space/my space"
}

output "space_description" {
  value = data.spacelift_space.space.description
}