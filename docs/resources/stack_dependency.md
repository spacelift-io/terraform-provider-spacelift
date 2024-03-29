---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "spacelift_stack_dependency Resource - terraform-provider-spacelift"
subcategory: ""
description: |-
  spacelift_stack_dependency represents a Spacelift stack dependency - a dependency between two stacks. When one stack depends on another, the tracked runs of the stack will not start until the dependent stack is successfully finished. Additionally, changes to the dependency will trigger the dependent.
---

# spacelift_stack_dependency (Resource)

`spacelift_stack_dependency` represents a Spacelift **stack dependency** - a dependency between two stacks. When one stack depends on another, the tracked runs of the stack will not start until the dependent stack is successfully finished. Additionally, changes to the dependency will trigger the dependent.

## Example Usage

```terraform
resource "spacelift_stack" "infra" {
  branch     = "master"
  name       = "Infrastructure stack"
  repository = "core-infra"
}


resource "spacelift_stack" "app" {
  branch     = "master"
  name       = "Application stack"
  repository = "app"
}

resource "spacelift_stack_dependency" "test" {
  stack_id            = spacelift_stack.app.id
  depends_on_stack_id = spacelift_stack.infra.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `depends_on_stack_id` (String) immutable ID (slug) of stack to depend on.
- `stack_id` (String) immutable ID (slug) of stack which has a dependency.

### Read-Only

- `id` (String) The ID of this resource.
