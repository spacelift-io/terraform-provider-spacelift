---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "spacelift_role Resource - terraform-provider-spacelift"
subcategory: ""
description: |-
  spacelift_role represents a Spacelift role - a collection of permissions that can be assigned to IdP groups or API keys to control access to Spacelift resources and operations.
  Note: you must have admin access to the root Space in order to create or mutate roles.
---

# spacelift_role (Resource)

`spacelift_role` represents a Spacelift **role** - a collection of permissions that can be assigned to IdP groups or API keys to control access to Spacelift resources and operations.

**Note:** you must have admin access to the `root` Space in order to create or mutate roles.

## Example Usage

```terraform
resource "spacelift_role" "readonly" {
  name        = "ReadOnly Role"
  description = "A role that can read Space resources and confirm runs"
  actions     = ["SPACE_READ", "RUN_CONFIRM"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `actions` (Set of String) List of actions (permissions) associated with the role. For example: `SPACE_READ`, `SPACE_WRITE`, `SPACE_ADMIN`, `RUN_TRIGGER`. All possible actions can be listed using the `spacelift_role_actions` data source.
- `name` (String) Human-readable, free-form name of the role

### Optional

- `description` (String) Human-readable, free-form description of the role

### Read-Only

- `id` (String) Unique identifier (ULID) of the role. Example: `01K07523Q8B4TBF0YHQRF6J5MW`.
- `slug` (String) URL-friendly unique identifier of the role, generated from the name. Example: `space-admin`.

## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
terraform import spacelift_role.readonly $ROLE_ID
```
