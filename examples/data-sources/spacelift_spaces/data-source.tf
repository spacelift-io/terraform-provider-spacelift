data "spacelift_spaces" "this" {
}

output "spaces" {
  value = data.spacelift_spaces.this.spaces
}
