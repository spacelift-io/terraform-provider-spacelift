terraform {
  required_providers {
    spacelift = {
      source = "spacelift.io/spacelift-io/spacelift"
    }
  }
}

provider "spacelift" {
  api_key_endpoint = ""
  api_key_id = ""
  api_key_secret = ""
}


resource "spacelift_idp_group_mapping" "test" {
  name = "test"
  policy {
    space_id = "root"
    role     = "ADMIN"
  }
  policy {
    space_id = "legacy"
    role     = "ADMIN"
  }
}
