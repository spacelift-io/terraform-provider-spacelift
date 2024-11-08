data "spacelift_space_by_path" "space" {
  space_path = "root/second space/my space"
}

output "space_description" {
  value = data.spacelift_space_by_path.space.description
}

// The following example shows how to use a relative path. If this is used in a stack in the root space, this is identical to using a path of `root/second space/my space`.
data "spacelift_space_by_relative_path" "space" {
  space_path = "second space/my space"
}
