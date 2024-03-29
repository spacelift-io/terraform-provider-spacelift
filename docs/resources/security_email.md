---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "spacelift_security_email Resource - terraform-provider-spacelift"
subcategory: ""
description: |-
  spacelift_security_email represents an email address that receives notifications about security issues in Spacelift.
---

# spacelift_security_email (Resource)

`spacelift_security_email` represents an email address that receives notifications about security issues in Spacelift.

## Example Usage

```terraform
resource "spacelift_security_email" "example" {
  email = "user@example.com"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `email` (String) Email address to which the security notifications are sent

### Read-Only

- `id` (String) The ID of this resource.
