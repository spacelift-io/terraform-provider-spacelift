# For all saved filter
data "spacelift_saved_filters" "all" {}

# Filters with a matching type
data "spacelift_saved_filters" "stack_filters" {
  type = "stacks"
}

# Filters with a matching name
data "spacelift_saved_filters" "my_filters" {
  name = "My best filter"
}

output "filter_ids" {
  value = data.spacelift_saved_filters.stack_filters.filters[*].id
}
