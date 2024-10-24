data "spacelift_space_by_path" "space" {
  space_path = "root/second space/my space"
}

output "space_description" {
  value = data.spacelift_space_by_path.space.description
}

// Assuming this data source is invoked in a run that belongs to a stack in a space located at "root", then the following data source shall be equal to the one above.
data "spacelift_space_by_relative_path" "space" {
  space_path = "second space/my space"
}
