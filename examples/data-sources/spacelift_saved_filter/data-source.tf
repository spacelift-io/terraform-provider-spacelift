data "spacelift_saved_filter" "filter" {
  filter_id = spacelift_saved_filter.filter.id
}

output "filter_data" {
  value = data.spacelift_saved_filter.filter.data
}
