---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "spacelift_context Data Source - terraform-provider-spacelift"
subcategory: ""
description: |-
  spacelift_context represents a Spacelift context - a collection of configuration elements (either environment variables or mounted files) that can be administratively attached to multiple stacks (spacelift_stack) or modules (spacelift_module) using a context attachment (spacelift_context_attachment)`
---

# spacelift_context (Data Source)

`spacelift_context` represents a Spacelift **context** - a collection of configuration elements (either environment variables or mounted files) that can be administratively attached to multiple stacks (`spacelift_stack`) or modules (`spacelift_module`) using a context attachment (`spacelift_context_attachment`)`

## Example Usage

```terraform
data "spacelift_context" "prod-k8s-ie" {
  context_id = "prod-k8s-ie"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `context_id` (String) immutable ID (slug) of the context

### Optional

- `after_apply` (List of String) List of after-apply scripts
- `after_destroy` (List of String) List of after-destroy scripts
- `after_init` (List of String) List of after-init scripts
- `after_perform` (List of String) List of after-perform scripts
- `after_plan` (List of String) List of after-plan scripts
- `after_run` (List of String) List of after-run scripts
- `before_apply` (List of String) List of before-apply scripts
- `before_destroy` (List of String) List of before-destroy scripts
- `before_init` (List of String) List of before-init scripts
- `before_perform` (List of String) List of before-perform scripts
- `before_plan` (List of String) List of before-plan scripts

### Read-Only

- `description` (String) free-form context description for users
- `id` (String) The ID of this resource.
- `labels` (Set of String)
- `name` (String) name of the context
- `space_id` (String) ID (slug) of the space the context is in
