# Lookup an integration by its ID:
data "spacelift_aws_integration" "example" {
  integration_id = "01FPAH5J0JFYSM5953T9KT2VS9"
}

# Lookup an integration by its name:
data "spacelift_aws_integration" "example" {
  name = "Production"
}
