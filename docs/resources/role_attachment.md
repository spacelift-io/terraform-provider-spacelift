---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "spacelift_role_attachment Resource - terraform-provider-spacelift"
subcategory: ""
description: |-
  spacelift_role_attachment represents a Spacelift role attachment between:
  an API key and a rolean IdP Group Mapping and a roleor a user and a role
  Exactly one of api_key_id, idp_group_mapping_id, or user_id must be set.
---

# spacelift_role_attachment (Resource)

`spacelift_role_attachment` represents a Spacelift role attachment between:
- an API key and a role
- an IdP Group Mapping and a role
- or a user and a role
Exactly one of `api_key_id`, `idp_group_mapping_id`, or `user_id` must be set.

## Example Usage

```terraform
resource "spacelift_role" "devops" {
  name    = "A role for DevOps team"
  actions = ["SPACE_ADMIN"]
}
resource "spacelift_space" "devops" {
  name            = "DevOps"
  parent_space_id = "root"
}

# Attach an API key to a role in a specific space
resource "spacelift_role_attachment" "api_key_attachment" {
  api_key_id = "01K09KERE33P95V40YRWWRVAZT"
  role_id    = spacelift_role.devops.id
  space_id   = spacelift_space.devops.id
}

# Attach an IDP group mapping to a role in a specific space
resource "spacelift_idp_group_mapping" "devops" {
  name = "devops-group"
}

resource "spacelift_role_attachment" "idp_group_attachment" {
  idp_group_mapping_id = spacelift_idp_group_mapping.devops.id
  role_id              = spacelift_role.devops.id
  space_id             = spacelift_space.devops.id
}

# Attach a user to a role in a specific space
resource "spacelift_user" "devops" {
  username         = "devops-user"
  invitation_email = "devops@example.com"
}

resource "spacelift_role_attachment" "user_attachment" {
  user_id  = spacelift_user.devops.id
  role_id  = spacelift_role.devops.id
  space_id = spacelift_space.devops.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `role_id` (String) ID of the role (ULID format) to attach to the API key, IdP Group or to the user. For example: `01F8Z5K4Y3D1G2H3J4K5L6M7N8`.
- `space_id` (String) ID of the space where the role attachment should be created

### Optional

- `api_key_id` (String) ID of the API key (ULID format) to attach to the role. For example: `01F8Z5K4Y3D1G2H3J4K5L6M7N8`.
- `idp_group_mapping_id` (String) ID of the IdP Group Mapping (ULID format) to attach to the role. For example: `01F8Z5K4Y3D1G2H3J4K5L6M7N8`.
- `user_id` (String) ID of the user (ULID format) to attach to the role. For example: `01F8Z5K4Y3D1G2H3J4K5L6M7N8`.

### Read-Only

- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
terraform import spacelift_role_attachment.api_key_attachment API/$ROLE_ATTACHMENT_ID

terraform import spacelift_role_attachment.idp_group_attachment IDP/$ROLE_ATTACHMENT_ID

terraform import spacelift_role_attachment.user_attachment USER/$ROLE_ATTACHMENT_ID
```
