---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "spacelift_stack_aws_role Data Source - terraform-provider-spacelift"
subcategory: ""
description: |-
  ~> Note: spacelift_stack_aws_role is deprecated. Please use spacelift_aws_role instead. The functionality is identical.
  spacelift_stack_aws_role represents cross-account IAM role delegation https://docs.aws.amazon.com/IAM/latest/UserGuide/tutorial_cross-account-with-roles.html between the Spacelift worker and an individual stack or module. If this is set, Spacelift will use AWS STS to assume the supplied IAM role and put its temporary credentials in the runtime environment.
  If you use private workers, you can also assume IAM role on the worker side using your own AWS credentials (e.g. from EC2 instance profile).
  Note: when assuming credentials for shared worker, Spacelift will use $accountName@$stackID or $accountName@$moduleID as external ID https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html and $runID@$stackID@$accountName truncated to 64 characters as session ID https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole.
---

# spacelift_stack_aws_role (Data Source)

~> **Note:** `spacelift_stack_aws_role` is deprecated. Please use `spacelift_aws_role` instead. The functionality is identical.

`spacelift_stack_aws_role` represents [cross-account IAM role delegation](https://docs.aws.amazon.com/IAM/latest/UserGuide/tutorial_cross-account-with-roles.html) between the Spacelift worker and an individual stack or module. If this is set, Spacelift will use AWS STS to assume the supplied IAM role and put its temporary credentials in the runtime environment.

If you use private workers, you can also assume IAM role on the worker side using your own AWS credentials (e.g. from EC2 instance profile).

Note: when assuming credentials for **shared worker**, Spacelift will use `$accountName@$stackID` or `$accountName@$moduleID` as [external ID](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html) and `$runID@$stackID@$accountName` truncated to 64 characters as [session ID](https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole).

## Example Usage

```terraform
# For a Module
data "spacelift_stack_aws_role" "k8s-module" {
  module_id = "k8s-module"
}

# For a Stack
data "spacelift_stack_aws_role" "k8s-core" {
  stack_id = "k8s-core"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `module_id` (String) ID of the module which assumes the AWS IAM role
- `stack_id` (String) ID of the stack which assumes the AWS IAM role

### Read-Only

- `duration_seconds` (Number) AWS IAM role session duration in seconds
- `external_id` (String) Custom external ID (works only for private workers).
- `generate_credentials_in_worker` (Boolean) Generate AWS credentials in the private worker
- `id` (String) The ID of this resource.
- `region` (String) AWS region to select a regional AWS STS endpoint.
- `role_arn` (String) ARN of the AWS IAM role to attach
