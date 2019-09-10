# Spacelift Terraform provider

The Spacelift Terraform provider is used to programmatically interact with its GraphQL API, allowing Spacelift to declaratively manage itself 🤯

The full list of supported resources is available [here](#resources).

## Example usage

```python
provider "spacelift" {}

resource "spacelift_stack" "core-infra-production" {
  name = "Core Infrastructure (production)"

  administrative    = true
  branch            = "master"
  description       = "Shared production infrastructure (networking, k8s)"
  readers_team      = "engineering"
  repository        = "core-infra"
  terraform_version = "0.12.6"
  writers_team      = "devops"
}
```

## Setup

This provider is designed to require no setup. All runs in an [administrative stack](#todo) receive a temporary JWT token in the `SPACELIFT_API_TOKEN` environment variable, which is all the provider needs to run. **We strongly recommend using this approach**:

```python
provider "spacelift" {}
```

The alternative approach when not running in Spacelift is to pass a human user's JWT token, either through the environment (`SPACELIFT_API_TOKEN` variable) or using the provider's `api_token` field. Note though that all Spacelift tokens have a short expiry, so that in practice you will need to generate a new token before each Terraform run. **We discourage this approach**:

```python
variable "spacelift_api_token" {}

provider "spacelift" {
  api_token = var.spacelift_api_token
}
```

## Resources

The Spacelift Terraform provider provides the following building blocks:

- `spacelift_context` - [data source](#spacelift_context-data-source) and [resource](#spacelift_context-resource);
- `spacelift_context_attachment` - [data source](#spacelift_context_attachment_data-source) and [resource](#spacelift_context_attachment_resource);
- `spacelift_environment_variable` - [data source](#spacelift_environment_variable-data-source) and [resource](#spacelift_environment_variable-resource);
- `spacelift_mounted_file` - [data source](#spacelift_mounted_file-data-source) and [resource](#spacelift_mounted_file-resource);
- `spacelift_stack` - [data source](#spacelift_stack-data-source) and [resource](#spacelift_stack-resource);
- `spacelift_stack_aws_role` - [data source](#spacelift_stack_aws_role-data-source) and [resource](#spacelift_stack_aws_role-resource);

### `spacelift_context` data source

`spacelift_context` represents a Spacelift **context** - a collection of configuration elements (either environment variables or mounted files) that can be administratively attached to multiple [**stacks**](#spacelift_stack-resource) using a [**context attachment**]($spacelift_context_attachment_resource) resource.

#### Example usage

```python
data "spacelift_context" "prod-k8s-ie" {
  context_id = "prod-k8s-ie"
}
```

#### Argument reference

The following arguments are supported:

- `context_id` - (Required) The immutable identifier (slug) of the context;

#### Attributes reference

See the [context resource](#spacelift_context-resource) for details on the returned attributes - they are identical.

[^ Back to all resources](#resources)

### `spacelift_context` resource

`spacelift_context` represents a Spacelift **context** - a collection of configuration elements (either environment variables or mounted files) that can be administratively attached to multiple [**stacks**](#spacelift_stack-resource) using a [**context attachment**]($spacelift_context_attachment_resource) resource.

#### Example usage

```python
resource "spacelift_context" "prod-k8s-ie" {
  description = "Configuration details for the compute cluster in 🇮🇪"
  name        = "Production cluster (Ireland)"
}
```

#### Argument reference

The following arguments are supported:

- `name` - (Required) - Name of the context meant to be unique within one account;
- `description` - (Optional) - Free-form context description for GUI users;

#### Attributes reference

In addition to all arguments above, the following attributes are exported:

- `id` - The immutable identifier (slug) of the context;

[^ Back to all resources](#resources)

### `spacelift_context_attachment` data source

`spacelift_context_attachment` represents a Spacelift attachment of a single [context](#spacelift_context-resource) to a single [stack](#spacelift_stack-resource), with a predefined priority.

#### Example usage

```python
data "spacelift_context_attachment" "attachment" {
  attachment_id = "prod-k8s-ie/01DJN6A8MHD9ZKYJ3NHC5QAPTV"
}
```

#### Argument reference

The following arguments are supported:

- `attachment_id` - (Required) - Unique and opaque identifier of the attachment;

#### Attributes reference

See the [context attachment resource](#spacelift_context_attachment-resource) for details on the returned attributes - they are identical.

[^ Back to all resources](#resources)

### `spacelift_context_attachment` resource

`spacelift_context_attachment` represents a Spacelift attachment of a single [context](#spacelift_context-resource) to a single [stack](#spacelift_stack-resource), with a predefined priority.

#### Example usage

```python
resource "spacelift_context_attachment" "attachment" {
  context_id = "prod-k8s-ie"
  stack_id   = "k8s-core"
  priority   = 0
}
```

#### Argument reference

The following arguments are supported:

- `context_id` - (Required) - ID of the context to attach;
- `stack_id` - (Required) - ID of the stack to attach the context to;
- `priority` - (Optional) - Priority of the context attachment, used in cases where multiple contexts define the same value: the one with the lowest `priority` value will take precedence;

#### Attributes reference

In addition to all arguments above, the following attributes are exported:

- `id` - The immutable ID of the attachment;

[^ Back to all resources](#resources)

### `spacelift_environment_variable` data source

`spacelift_environment_variable` defines an environment variable on the [context](#spacelift_context-resource) or a [stack](#spacelift_context-stack), thereby allowing to pass and share various secrets and configuration details between Spacelift stacks.

#### Example usage

For a context:

```python
data "spacelift_environment_variable" "ireland-kubeconfig" {
  context_id = "prod-k8s-ie"
  name       = "KUBECONFIG"
}
```

For a stack:

```python
data "spacelift_environment_variable" "core-kubeconfig" {
  stack_id = "k8s-core"
  name     = "KUBECONFIG"
}
```

#### Argument reference

The following arguments are supported:

- `name` - (Required) - Name of the environment variable;
- `context_id` - (Optional) - ID of the context on which the environment variable is defined;
- `stack_id` - (Optional) - ID of the stack on which the environment variable is defined;

Note that `context_id` and `stack_id` are mutually exclusive, and exactly one of them _must_ be specified.

#### Attributes reference

See the [environment variable resource](#spacelift_environment_variable-resource) for details on the returned attributes - they are identical.

[^ Back to all resources](#resources)

### `spacelift_environment_variable` resource

`spacelift_environment_variable` defines an environment variable on the [context](#spacelift_context-resource) or a [stack](#spacelift_context-stack), thereby allowing to pass and share various secrets and configuration details between Spacelift stacks.

#### Example usage

For a context:

```python
resource "spacelift_environment_variable" "ireland-kubeconfig" {
  context_id = "prod-k8s-ie"
  name       = "KUBECONFIG"
  value      = "/project/spacelift/kubeconfig"
  write_only = false
}
```

For a stack:

```python
resource "spacelift_environment_variable" "core-kubeconfig" {
  stack_id   = "k8s-core"
  name       = "KUBECONFIG"
  value      = "/project/spacelift/kubeconfig"
  write_only = false
}
```

#### Argument reference

The following arguments are supported:

- `name` - (Required) - Name of the environment variable;
- `value` - (Required) - Value of the environment variable;
- `write_only` - (Optional) - Indicates whether the value can be read back outside a Run - for safety, this defaults to **true**;
- `context_id` - (Optional) - ID of the context on which the environment variable is defined;
- `stack_id` - (Optional) - ID of the stack on which the environment variable is defined;

Note that `context_id` and `stack_id` are mutually exclusive, and exactly one of them _must_ be specified.

Also note that if `write_only` is set to `true`, the `value` is not stored in the state, and will not be reported back by either the data source or the resource. Instead, its SHA-256 checksum will be used to compare the new value to the one passed to Spacelift.

#### Attributes reference

In addition to all arguments above, the following attributes are exported:

- `id` - ID of the environment variable;
- `checksum` - SHA-256 checksum of the value;

[^ Back to all resources](#resources)

### `spacelift_mounted_file` data source

`spacelift_mounted_file` represents a file mounted in each Run's workspace that is part of a configuration of a [context](#spacelift_context-resource) or a [stack](#spacelift_context-stack). In principle, it's very similar to an [environment variable](#spacelift_environment_variable-resource) except that the value is written to the filesystem rather than passed to the environment.

#### Example usage

For a context:

```python
data "spacelift_mounted_file" "ireland-kubeconfig" {
  context_id    = "prod-k8s-ie"
  relative_path = "kubeconfig"
}
```

For a stack:

```python
data "spacelift_mounted_file" "core-kubeconfig" {
  stack_id      = "k8s-core"
  relative_path = "kubeconfig"
}
```

#### Argument reference

The following arguments are supported:

- `relative_path` - (Required) - Relative path to the mounted file. The full (absolute) path to the file will be prefixed with `/spacelift/project/`;
- `context_id` - (Optional) - ID of the context on which the environment variable is defined;
- `stack_id` - (Optional) - ID of the stack on which the environment variable is defined;

Note that `context_id` and `stack_id` are mutually exclusive, and exactly one of them _must_ be specified.

#### Attributes reference

See the [mounted file resource](#spacelift_mounted_file-resource) for details on the returned attributes - they are identical.

[^ Back to all resources](#resources)

### `spacelift_mounted_file` resource

`spacelift_mounted_file` represents a file mounted in each Run's workspace that is part of a configuration of a [context](#spacelift_context-resource) or a [stack](#spacelift_context-stack). In principle, it's very similar to an [environment variable](#spacelift_environment_variable-resource) except that the value is written to the filesystem rather than passed to the environment.

#### Example usage

For a context:

```python
resource "spacelift_mounted_file" "ireland-kubeconfig" {
  context_id    = "prod-k8s-ie"
  relative_path = "kubeconfig"
  value         = filebase64("${path.module}/kubeconfig.json")
}
```

For a stack:

```python
resource "spacelift_mounted_file" "core-kubeconfig" {
  stack_id      = "k8s-core"
  relative_path = "kubeconfig"
  value         = filebase64("${path.module}/kubeconfig.json")
}
```

#### Argument reference

The following arguments are supported:

- `content` - (Required) - Content of the mounted file encoded using Base-64;
- `relative_path` - (Required) - Relative path to the mounted file, without the `/spacelift/project/` prefix;
- `context_id` - (Optional) - ID of the context on which the mounted file is defined;
- `stack_id` - (Optional) - ID of the stack on which the mounted file is defined;
- `write_only` - (Optional) - Indicates whether the content can be read back outside a Run;

Note that `context_id` and `stack_id` are mutually exclusive, and exactly one of them _must_ be specified.

Also note that if `write_only` is set to `true`, the `content` is not stored in the state, and will not be reported back by either the data source or the resource. Instead, its SHA-256 checksum will be used to compare the new value to the one passed to Spacelift.

#### Attributes reference

In addition to all arguments above, the following attributes are exported:

- `id` - ID of the mounted file;
- `checksum` - SHA-256 checksum of the (base-64 encoded) content;

[^ Back to all resources](#resources)

### `spacelift_stack` data source

`spacelift_stack` combines source code and configuration to create a runtime environment where resources are managed. In this way it's similar to a stack in AWS CloudFormation, or a project on generic CI/CD platforms.

#### Example usage

```python
data "spacelift_stack" "k8s-core" {
  stack_id = "k8s-core"
}
```

#### Argument reference

The following arguments are supported:

- `stack_id` - (Required) - The ID (slug) of the stack;

#### Attributes reference

See the [stack resource](#spacelift_stack-resource) for details on the returned attributes - they are identical.

[^ Back to all resources](#resources)

### `spacelift_stack` resource

`spacelift_stack` combines source code and configuration to create a runtime environment where resources are managed. In this way it's similar to a stack in AWS CloudFormation, or a project on generic CI/CD platforms.

#### Example usage

```python
resource "spacelift_stack" "k8s-core" {
  administrative    = true
  branch            = "master"
  description       = "Shared cluster services (Datadog, Istio etc.)"
  name              = "Kubernetes core services"
  readers_team      = "engineering"
  repository        = "core-infra"
  terraform_version = "0.12.6"
  writers_team      = "devops"
}
```

With IAM role delegation (only required fields):

```python
resource "spacelift_stack" "k8s-core" {
  branch            = "master"
  name              = "Kubernetes core services"
  repository        = "core-infra"
}

resource "aws_iam_role" "spacelift" {
  name = "spacelift"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [jsondecode(spacelift_stack.k8s-core.aws_assume_role_policy_statement)]
  })
}

resource "aws_iam_role_policy_attachment" "batch_service_role" {
  role       = aws_iam_role.spacelift.name
  policy_arn = "arn:aws:iam::aws:policy/PowerUserAccess"
}

resource "spacelift_stack_aws_role" "k8s-core" {
  stack_id = spacelift_stack.k8s-core.id
  role_arn = aws_iam_role.spacelift.arn
}
```

#### Argument reference

The following arguments are supported:

- `branch` - (Required) - GitHub branch to apply changes to;
- `name` - (Required) - Name of the stack - should be unique within one account;
- `repository` - (Required) - Name of the GitHub repository, without the owner part;
- `administrative` - (Optional) - Indicates whether this stack can manage others;
- `description` - (Optional) - Free-form stack description for GUI users;
- `import_state` - (Optional) - Content of the state file to import if Spacelift should manage the stack but the state has already been created externally. This only applies during creation and the field can be deleted afterwards without triggering a resource change;
- `manage_state` - (Optional) - Boolean that determines if Spacelift should manage state for this stack. Default: `true`;
- `readers_team` - (Optional) - Slug of the GitHub team whose members get read-only access;
- `terraform_version` - (Optional) - Terraform version to use;
- `writers_team` - (Optional) - Slug of the GitHub team whose members get read-write access;

#### Attributes reference

In addition to all arguments above, the following attributes are exported:

- `id` - The immutable ID (slug) of the stack;
- `aws_assume_role_policy_statement` - JSON-encoded AWS IAM policy for the AWS IAM role trust relationship;

[^ Back to all resources](#resources)

### `spacelift_stack_aws_role` data source

`spacelift_stack_aws_role` represents [cross-account IAM role delegation](https://docs.aws.amazon.com/IAM/latest/UserGuide/tutorial_cross-account-with-roles.html) between the Spacelift worker and an individual [stack](#spacelift_stack-resource). If this is set, Spacelift will use AWS STS to assume the supplied IAM role and put its temporary credentials in the runtime environment.

Note: when assuming credentials, Spacelift will use `$accountName/$stackID` as [external ID](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html) and Run ID as [session ID](https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole).

#### Example usage

```python
data "spacelift_stack_aws_role" "k8s-core" {
  stack_id = "k8s-core"
}
```

#### Argument reference

The following arguments are supported:

- `stack_id` - (Required) - The immutable ID (slug) of the stack;

#### Attributes reference

See the [stack AWS role resource](#spacelift_stack_aws_role-resource) for details on the returned attributes - they are identical.

[^ Back to all resources](#resources)

### `spacelift_stack_aws_role` resource

`spacelift_stack_aws_role` represents [cross-account IAM role delegation](https://docs.aws.amazon.com/IAM/latest/UserGuide/tutorial_cross-account-with-roles.html) between the Spacelift worker and an individual [stack](#spacelift_stack-resource). If this is set, Spacelift will use AWS STS to assume the supplied IAM role and put its temporary credentials in the runtime environment.

Note: when assuming credentials, Spacelift will use `$accountName/$stackID` as [external ID](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html) and Run ID as [session ID](https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole).

#### Example usage

```python
resource "spacelift_stack_aws_role" "k8s-core" {
  stack_id = "k8s-core"
  role_arn = "arn:aws:iam::012345678910:role/terraform"
}
```

#### Argument reference

The following arguments are supported:

- `role_arn` - (Required) - ARN of the AWS IAM role to attach;
- `stack_id` - (Required) - ID of the stack which assumes the AWS IAM role;

#### Attributes reference

In addition to all arguments above, the following attributes are exported:

- `id` - The immutable ID (slug) of the role;

[^ Back to all resources](#resources)
