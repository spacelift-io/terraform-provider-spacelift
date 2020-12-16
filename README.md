# Spacelift Terraform provider

The Spacelift Terraform provider is used to programmatically interact with its GraphQL API, allowing Spacelift to declaratively manage itself 🤯

The full list of supported resources is available [here](#resources).

## Example usage

```python
provider "spacelift" {}

resource "spacelift_stack" "core-infra-production" {
  name = "Core Infrastructure (production)"

  administrative    = true
  autodeploy        = true
  branch            = "master"
  description       = "Shared production infrastructure (networking, k8s)"
  repository        = "core-infra"
  terraform_version = "0.12.6"
}
```
## Terraform 0.13.x

With Terraform 0.13.x you also have to add the following:
```python
terraform {
  required_providers {
    spacelift = {
      source  = "spacelift.io/spacelift-io/spacelift"
    }
  }
}
```

## Setup

### Running inside Spacelift

When executed from inside a Spacelift run, this provider is designed to require no setup at all. All Spacelift jobs receive a temporary authentication token in the `SPACELIFT_API_TOKEN` environment variable, which is all the provider needs to run. 

Important thing to note here though is that this API token will only allow operations to be performed on its own account. If you need to set up multiple Spacelift accounts from the same Terraform run, please refer to the next section describing the use of [API keys](https://docs.spacelift.io/integrations/api#api-key-management).

```python
provider "spacelift" {}
```

### Running outside of Spacelift

If you want to run the Spacelift provider outside of Spacelift, or you need to manage resources across multiple Spacelift accounts from the same Terraform project, the preferred method is to generate and use dedicated [API keys](https://docs.spacelift.io/integrations/api#api-key-management). Note that unless you're just accessing whitelisted data resources, the Terraform use case will normally require marking the API key as administrative.

In order to set up the provider to use an API key, you will need the key ID, secret and the API key endpoint:

```python
variable "spacelift_key_id" {}
variable "spacelift_key_secret" {}

provider "spacelift" {
  api_key_endpoint = "https://your-account.app.spacelift.io"
  api_key_id       = var.spacelift_key_id
  api_key_secret   = var.spacelift_key_secret
}
```

These values can also be passed using environment variables, though this will only work to set up the provider for a single Spacelift account:

- `SPACELIFT_API_KEY_ENDPOINT` for `api_key_endpoint`;
- `SPACELIFT_API_KEY_ID` for `api_key_id`;
- `SPACELIFT_API_KEY_SECRET` for `api_key_secret`;

If you want to talk to multiple Spacelift accounts, you just need to set up [provider aliases](https://www.terraform.io/docs/configuration/providers.html#alias-multiple-provider-configurations) like this:

```python
variable "spacelift_first_key_id" {}
variable "spacelift_first_key_secret" {}

variable "spacelift_second_key_id" {}
variable "spacelift_second_key_secret" {}

provider "spacelift" {
  alias = "first"

  api_key_endpoint = "https://first.app.spacelift.io"
  api_key_id       = var.spacelift_first_key_id
  api_key_secret   = var.spacelift_first_key_secret
}

provider "spacelift" {
  alias = "second"

  api_key_endpoint = "https://second.app.spacelift.io"
  api_key_id       = var.spacelift_second_key_id
  api_key_secret   = var.spacelift_second_key_secret
}
```

If you're running from inside Spacelift, you can still use the default, zero-setup provider for the current account with providers for accounts set up through API keys:

```python
variable "spacelift_that_key_id" {}
variable "spacelift_that_key_secret" {}

provider "spacelift" {
  alias = "this"
}

provider "spacelift" {
  alias = "that"

  api_key_endpoint = "https://that.app.spacelift.io"
  api_key_id       = var.spacelift_that_key_id
  api_key_secret   = var.spacelift_that_key_secret
}
```

The alternative approach when running locally is to pass a human user's JWT token, either through the environment (`SPACELIFT_API_TOKEN` variable) or using the provider's `api_token` field. Note though that all Spacelift tokens have a short expiry, so that in practice you will need to generate a new token before each Terraform run. **We stongly discourage this approach** and suggest using an API key instead for all systematic use cases:

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
- `spacelift_current_stack` - [data source](#spacelift_current_stack-data-source);
- `spacelift_environment_variable` - [data source](#spacelift_environment_variable-data-source) and [resource](#spacelift_environment_variable-resource);
- `spacelift_ips` - [data source](#spacelift_ips-data-source);
- `spacelift_module` - [data source](#spacelift_module-data-source) and [resource](#spacelift_module-resource);
- `spacelift_mounted_file` - [data source](#spacelift_mounted_file-data-source) and [resource](#spacelift_mounted_file-resource);
- `spacelift_policy` - [data source](#spacelift_policy-data-source) and [resource](#spacelift_policy-resource);
- `spacelift_policy_attachment` - [resource](#spacelift_policy_attachment-resource);
- `spacelift_stack` - [data source](#spacelift_stack-data-source) and [resource](#spacelift_stack-resource);
- `spacelift_aws_role` - [data source](#spacelift_aws_role-data-source) and [resource](#spacelift_aws_role-resource);
- `spacelift_gcp_service_account` - [data source](#spacelift_gcp_service_account-data-source) and [resource](#spacelift_gcp_service_account-resource);
- `spacelift_webhook` - [data source](#spacelift_webhook-data-source) and [resource](#spacelift_webhook-resource);
- `spacelift_worker_pool` - [data source](#spacelift_worker_pool-data-source) and [resource](#spacelift_worker_pool-resource);

**Note:** `spacelift_stack_aws_role` and `spacelift_stack_gcp_service_account` are **deprecated**. Please use [spacelift_aws_role](#spacelift_aws_role-data-source) and [spacelift_gcp_service_account](#spacelift_gcp_service_account-data-source) instead.

### `spacelift_context` data source

`spacelift_context` represents a Spacelift **context** - a collection of configuration elements (either environment variables or mounted files) that can be administratively attached to multiple [**stacks**](#spacelift_stack-resource) or [**modules**](#spacelift_module-resource) using a [**context attachment**]($spacelift_context_attachment_resource) resource.

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

`spacelift_context` represents a Spacelift **context** - a collection of configuration elements (either environment variables or mounted files) that can be administratively attached to multiple [**stacks**](#spacelift_stack-resource) or [**modules**](#spacelift_module-resource) using a [**context attachment**]($spacelift_context_attachment_resource) resource.

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

`spacelift_context_attachment` represents a Spacelift attachment of a single [context](#spacelift_context-resource) to a single [stack](#spacelift_stack-resource) or [module](#spacelift_module-resource), with a predefined priority.

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

`spacelift_context_attachment` represents a Spacelift attachment of a single [context](#spacelift_context-resource) to a single [stack](#spacelift_stack-resource) or [module](#spacelift_module-resource), with a predefined priority.

#### Example usage

For a stack:

```python
resource "spacelift_context_attachment" "attachment" {
  context_id = "prod-k8s-ie"
  stack_id   = "k8s-core"
  priority   = 0
}
```

For a module:

```python
resource "spacelift_context_attachment" "attachment" {
  context_id = "prod-k8s-ie"
  module_id  = "k8s-module"
  priority   = 0
}
```

#### Argument reference

The following arguments are supported:

- `context_id` - (Required) - ID of the context to attach;
- `module_id` - (Optional) - ID of the module to attach the context to;
- `stack_id` - (Optional) - ID of the stack to attach the context to;
- `priority` - (Optional) - Priority of the context attachment, used in cases where multiple contexts define the same value: the one with the lowest `priority` value will take precedence;

Note that `module_id` and `stack_id` are mutually exclusive, and exactly one of them _must_ be specified.

#### Attributes reference

In addition to all arguments above, the following attributes are exported:

- `id` - The immutable ID of the attachment;

[^ Back to all resources](#resources)

### `spacelift_current_stack` data source

`spacelift_current_stack` is a data source that provides information about the current administrative stack if the run is executed within Spacelift by a stack or module. This allows clever tricks like attaching contexts or policies to the stack that manages them.

#### Example usage

```python
data "spacelift_current_stack" "this" {}

resource "spacelift_environment_variable" "core-kubeconfig" {
  stack_id   = data.spacelift_current_stack.this.id
  name       = "CHUNKY"
  value      = "bacon"
}
```

#### Attributes reference

This resource only exports one attribute:

- `id` - The immutable ID of the current stack or module;

[^ Back to all resources](#resources)

### `spacelift_environment_variable` data source

`spacelift_environment_variable` defines an environment variable on the [context](#spacelift_context-resource), [stack](#spacelift_stack-resource) or a [module](#spacelift_module-resource), thereby allowing to pass and share various secrets and configuration details between Spacelift stacks.

#### Example usage

For a context:

```python
data "spacelift_environment_variable" "ireland-kubeconfig" {
  context_id = "prod-k8s-ie"
  name       = "KUBECONFIG"
}
```

For a module:

```python
data "spacelift_environment_variable" "module-kubeconfig" {
  module_id = "k8s-module"
  name     = "KUBECONFIG"
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
- `module_id` - (Optional) - ID of the module on which the environment variable is defined;
- `stack_id` - (Optional) - ID of the stack on which the environment variable is defined;

Note that `context_id`, `module_id` and `stack_id` are mutually exclusive, and exactly one of them _must_ be specified.

#### Attributes reference

See the [environment variable resource](#spacelift_environment_variable-resource) for details on the returned attributes - they are identical.

[^ Back to all resources](#resources)

### `spacelift_environment_variable` resource

`spacelift_environment_variable` defines an environment variable on the [context](#spacelift_context-resource), [stack](#spacelift_stack-resource) or a [module](#spacelift_module-resource), thereby allowing to pass and share various secrets and configuration details between Spacelift stacks.

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

For a module:

```python
resource "spacelift_environment_variable" "module-kubeconfig" {
  module_id  = "k8s-module"
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
- `module_id` - (Optional) - ID of the module on which the environment variable is defined;
- `stack_id` - (Optional) - ID of the stack on which the environment variable is defined;

Note that `context_id`, `module_id`, `stack_id` are mutually exclusive, and exactly one of them _must_ be specified.

Also note that if `write_only` is set to `true`, the `value` is not stored in the state, and will not be reported back by either the data source or the resource. Instead, its SHA-256 checksum will be used to compare the new value to the one passed to Spacelift.

#### Attributes reference

In addition to all arguments above, the following attributes are exported:

- `id` - ID of the environment variable;
- `checksum` - SHA-256 checksum of the value;

[^ Back to all resources](#resources)

### `spacelift_ips` data source

`spacelift_ips` returns the list of Spacelift's outgoing IP addresses, which you can use to whitelist connections coming from the Spacelift's "mothership".

#### Example usage

```python
data "spacelift_ips" "ips" {}
```

#### Argument reference

This data source does not support arguments.

#### Attributes reference

The following attributes are exported:

- `ids` - set of Spacelift's outgoing IPs;

[^ Back to all resources](#resources)

### `spacelift_module` data source

`spacelift_module` is a special type of [stack](#spacelift_stack-resource) used to test Terraform modules.

#### Example usage

```python
data "spacelift_module" "k8s-module" {
  module_id = "k8s-module"
}
```

#### Argument reference

The following arguments are supported:

- `module_id` - (Required) - The ID (slug) of the module;

#### Attributes reference

See the [module resource](#spacelift_module-resource) for details on the returned attributes - they are identical.

[^ Back to all resources](#resources)

### `spacelift_module` resource

`spacelift_module` is a special type of [stack](#spacelift_stack-resource) used to test Terraform modules.

#### Example usage

```python
resource "spacelift_module" "k8s-module" {
  administrative    = true
  branch            = "master"
  description       = "Infra terraform module"
  repository        = "terraform-super-module"
}
```

For Gitlab-hosted repositories:

```python
resource "spacelift_module" "k8s-module-gitlab" {
  gitlab {
    namespace = "spacelift"
  }

  administrative    = true
  branch            = "master"
  description       = "Infra terraform module"
  repository        = "terraform-super-module"
}
```

With IAM role delegation (only required fields):

```python
resource "spacelift_module" "k8s-module" {
  branch            = "master"
  name              = "Kubernetes terraform module"
  repository        = "terraform-super-module"
}

resource "aws_iam_role" "spacelift" {
  name = "spacelift"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [jsondecode(spacelift_module.k8s-module.aws_assume_role_policy_statement)]
  })
}

resource "aws_iam_role_policy_attachment" "spacelift" {
  role       = aws_iam_role.spacelift.name
  policy_arn = "arn:aws:iam::aws:policy/PowerUserAccess"
}

resource "spacelift_aws_role" "k8s-module" {
  module_id = spacelift_module.k8s-module.id
  role_arn = aws_iam_role.spacelift.arn
}
```

#### Argument reference

The following arguments are supported:

- `branch` - (Required) - VCS branch to apply changes to;
- `gitlab` - (Optional) - namespace block for GitLab integration:
  - `namespace` - (Required) - GitLab namespace;
- `labels` - (Optional) - List of labels to set on the Module;
- `repository` - (Required) - Name of the repository, without the owner part;
- `administrative` - (Optional) - Indicates whether this stack can manage others. Default: `false`;
- `description` - (Optional) - Free-form stack description for GUI users;
- `shared_accounts` - (Optional) - List of the accounts (subdomains) which should have access to the Module;
- `worker_pool_id` - (Optional) - ID of the worker pool to use;

#### Attributes reference

In addition to all arguments above, the following attributes are exported:

- `id` - The immutable ID (slug) of the module;
- `aws_assume_role_policy_statement` - JSON-encoded AWS IAM policy for the AWS IAM role trust relationship;

[^ Back to all resources](#resources)

### `spacelift_mounted_file` data source

`spacelift_mounted_file` represents a file mounted in each Run's workspace that is part of a configuration of a [context](#spacelift_context-resource), [stack](#spacelift_stack-resource) or a [module](#spacelift_module-resource). In principle, it's very similar to an [environment variable](#spacelift_environment_variable-resource) except that the value is written to the filesystem rather than passed to the environment.

#### Example usage

For a context:

```python
data "spacelift_mounted_file" "ireland-kubeconfig" {
  context_id    = "prod-k8s-ie"
  relative_path = "kubeconfig"
}
```

For a module:

```python
data "spacelift_mounted_file" "module-kubeconfig" {
  module_id     = "k8s-module"
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
- `module_id` - (Optional) - ID of the module on which the environment variable is defined;
- `stack_id` - (Optional) - ID of the stack on which the environment variable is defined;

Note that `context_id` `module_id`,`stack_id` are mutually exclusive, and exactly one of them _must_ be specified.

#### Attributes reference

See the [mounted file resource](#spacelift_mounted_file-resource) for details on the returned attributes - they are identical.

[^ Back to all resources](#resources)

### `spacelift_mounted_file` resource

`spacelift_mounted_file` represents a file mounted in each Run's workspace that is part of a configuration of a [context](#spacelift_context-resource), a [stack](#spacelift_stack-resource) or a [module](#spacelift_module-resource). In principle, it's very similar to an [environment variable](#spacelift_environment_variable-resource) except that the value is written to the filesystem rather than passed to the environment.

#### Example usage

For a context:

```python
resource "spacelift_mounted_file" "ireland-kubeconfig" {
  context_id    = "prod-k8s-ie"
  relative_path = "kubeconfig"
  content       = filebase64("${path.module}/kubeconfig.json")
}
```

For a module:

```python
resource "spacelift_mounted_file" "module-kubeconfig" {
  module_id     = "k8s-module"
  relative_path = "kubeconfig"
  content       = filebase64("${path.module}/kubeconfig.json")
}
```

For a stack:

```python
resource "spacelift_mounted_file" "core-kubeconfig" {
  stack_id      = "k8s-core"
  relative_path = "kubeconfig"
  content       = filebase64("${path.module}/kubeconfig.json")
}
```

#### Argument reference

The following arguments are supported:

- `content` - (Required) - Content of the mounted file encoded using Base-64;
- `relative_path` - (Required) - Relative path to the mounted file, without the `/spacelift/project/` prefix;
- `context_id` - (Optional) - ID of the context on which the mounted file is defined;
- `module_id` - (Optional) - ID of the module on which the mounted file is defined;
- `stack_id` - (Optional) - ID of the stack on which the mounted file is defined;
- `write_only` - (Optional) - Indicates whether the content can be read back outside a Run;

Note that `context_id`, `module_id` and `stack_id` are mutually exclusive, and exactly one of them _must_ be specified.

Also note that if `write_only` is set to `true`, the `content` is not stored in the state, and will not be reported back by either the data source or the resource. Instead, its SHA-256 checksum will be used to compare the new value to the one passed to Spacelift.

#### Attributes reference

In addition to all arguments above, the following attributes are exported:

- `id` - ID of the mounted file;
- `checksum` - SHA-256 checksum of the (base-64 encoded) content;

[^ Back to all resources](#resources)

### `spacelift_policy` data source

`spacelift_policy` represents a Spacelift **policy** - a collection of customer-defined rules that are applied by Spacelift at one of the decision points within the application.

#### Example usage

```python
data "spacelift_policy" "policy" {
  policy_id = spacelift_policy.policy.id
}

output "policy_body" {
  value = data.spacelift_policy.policy.body
}
```

#### Argument reference

The following arguments are supported:

- `policy_id` - (Required) The immutable identifier (slug) of the policy;

#### Attributes reference

See the [policy resource](#spacelift_policy-resource) for details on the returned attributes - they are identical.

[^ Back to all resources](#resources)

### `spacelift_policy` resource

`spacelift_policy` represents a Spacelift **policy** - a collection of customer-defined rules that are applied by Spacelift at one of the decision points within the application. Please see tye `type` argument to learn about different supported policy types.

#### Example usage

```python
resource "spacelift_policy" "no-weekend-deploys" {
  name = "Let's not deploy any changes over the weekend
  body = file("policies/no-weekend-deploys.rego")
  type = "PLAN"
}

resource "spacelift_stack" "core-infra-production" {
  name       = "Core Infrastructure (production)"
  branch     = "master"
  repository = "core-infra"
}

resource "spacelift_policy_attachment" "no-weekend-deploys" {
  policy_id = spacelift_policy.no-weekend-deploys.id
  stack_id  = spacelift_stack.core-infra-production.id
}
```

#### Argument reference

The following arguments are supported:

- `name` - (Required) The name of the the policy - should be unique within one account;
- `body` - (Required) The body of the policy - may be provided inline or read from a file;
- `type` - (Required) One of the supported types of policies. Currently the following options are available:
  - `ACCESS` - controls who gets what level of access to a Stack;
  - `GIT_PUSH` - controls how Git push events are interpreted;
  - `LOGIN` - controls who can log in and in what capacity;
  - `INITIALIZATION` - controls whether Spacelift runs can be started;
  - `PLAN` - validates the outcome of Terraform plans;
  - `TASK` - controls whether Spacelift tasks can be started;
  - `TRIGGER` - controls what happens when blocking runs terminate;

#### Attributes reference

In addition to all arguments above, the following attributes are exported:

- `id` - The immutable ID of the policy;

[^ Back to all resources](#resources)

### `spacelift_policy_attachment` resource

`spacelift_policy_attachment` represents a relationship between a Policy and a Stack/Module. Each policy can only be attached to a stack/module once. `LOGIN` policies are the exception because they apply globally and not to individual stacks/modules. An attempt to attach one will fail.

#### Example usage

```python
resource "spacelift_policy" "no-weekend-deploys" {
  name = "Let's not deploy any changes over the weekend"
  body = file("policies/no-weekend-deploys.rego")
  type = "PLAN"
}

resource "spacelift_stack" "core-infra-production" {
  name       = "Core Infrastructure (production)"
  branch     = "master"
  repository = "core-infra"
}

resource "spacelift_policy_attachment" "no-weekend-deploys" {
  policy_id = spacelift_policy.no-weekend-deploys.id
  stack_id  = spacelift_stack.core-infra-production.id
}
```

#### Argument reference

The following arguments are supported:

- `module_id` - (Optional) - ID of the module to attach the policy to;
- `policy_id` - (Required) - ID of the policy to attach;
- `stack_id` - (Optional) - ID of the stack to attach the policy to;
- `custom_input` - (Optional) - JSON-encoded custom input to be passed to the evaluated document at the "attachment" key;

Note that `module_id` and `stack_id` are mutually exclusive, and exactly one of them _must_ be specified.

#### Attributes reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the attachment;

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
  autodeploy        = true
  branch            = "master"
  description       = "Shared cluster services (Datadog, Istio etc.)"
  name              = "Kubernetes core services"
  project_root      = "/project"
  repository        = "core-infra"
  terraform_version = "0.12.6"
}
```

For Gitlab-hosted repositories:

```python
resource "spacelift_stack" "k8s-core-gitlab" {
  gitlab {
    namespace = "spacelift"
  }

  administrative    = true
  autodeploy        = true
  branch            = "master"
  description       = "Shared cluster services (Datadog, Istio etc.)"
  name              = "Kubernetes core services"
  project_root      = "/project"
  repository        = "core-infra"
  terraform_version = "0.12.6"
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

resource "aws_iam_role_policy_attachment" "spacelift" {
  role       = aws_iam_role.spacelift.name
  policy_arn = "arn:aws:iam::aws:policy/PowerUserAccess"
}

resource "spacelift_aws_role" "k8s-core" {
  stack_id = spacelift_stack.k8s-core.id
  role_arn = aws_iam_role.spacelift.arn
}
```

For Pulumi Stacks:
```python
resource "spacelift_stack" "k8s-core-pulumi" {
  pulumi {
    login_url  = "s3://pulumi-state-bucket"
    stack_name = "kubernetes-core-services"
  }

  autodeploy        = true
  branch            = "master"
  description       = "Shared cluster services (Datadog, Istio etc.)"
  name              = "Kubernetes core services"
  project_root      = "/project"
  repository        = "core-infra"
  runner_image      = "public.ecr.aws/t0p9w2l5/runner-pulumi-javascript:latest"
}
```

#### Argument reference

The following arguments are supported:

- `branch` - (Required) - VCS branch to apply changes to;
- `name` - (Required) - Name of the stack - should be unique within one account;
- `repository` - (Required) - Name of the VCS repository, without the owner part;
- `gitlab` - (Optional) - namespace block for GitLab integration:
  - `namespace` - (Required) - GitLab namespace;
- `administrative` - (Optional) - Indicates whether this stack can manage others. Default: `false`;
- `autodeploy` - (Optional) - Indicates whether changes to this stack can be automatically deployed. Default: `false`;
- `before_init` - (Optional) - List of before-init scripts;
- `description` - (Optional) - Free-form stack description for GUI users;
- `import_state` - (Optional) - Content of the state file to import if Spacelift should manage the stack but the state has already been created externally. This only applies during creation and the field can be deleted afterwards without triggering a resource change;
- `labels` - (Optional) - List of labels to set on the Stack;
- `manage_state` - (Optional) - Boolean that determines if Spacelift should manage state for this stack. Default: `true`;
- `project_root` - (Optional) - Directory that is relative to the workspace root containing the entry point to the Stack.;
- `pulumi` - (Optional) - Pulumi-specific configuration block. Sets the Stack vendor to Pulumi if present.
  - `login_url` - (Required) - State backend to log in to.
  - `stack_name` - (Required) - Pulumi stack name to use in the backend (multiple stacks can use one state backend, provided they have different stack_name's).
- `runner_image` - (Optional) - Name of the Docker image used to process Runs;
- `terraform_version` - (Optional) - Terraform version to use;
- `worker_pool_id` - (Optional) - ID of the worker pool to use;

#### Attributes reference

In addition to all arguments above, the following attributes are exported:

- `id` - The immutable ID (slug) of the stack;
- `aws_assume_role_policy_statement` - JSON-encoded AWS IAM policy for the AWS IAM role trust relationship;

[^ Back to all resources](#resources)

### `spacelift_aws_role` data source

`spacelift_aws_role` represents [cross-account IAM role delegation](https://docs.aws.amazon.com/IAM/latest/UserGuide/tutorial_cross-account-with-roles.html) between the Spacelift worker and an individual [stack](#spacelift_stack-resource) or [module](#spacelift_module-resource). If this is set, Spacelift will use AWS STS to assume the supplied IAM role and put its temporary credentials in the runtime environment.

Note: when assuming credentials, Spacelift will use `$accountName/$stackID` or `$accountName/$moduleID` as [external ID](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html) and Run ID as [session ID](https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole).

#### Example usage

For a Module:

```python
data "spacelift_aws_role" "k8s-module" {
  module_id = "k8s-module"
}
```

For a Stack

```python
data "spacelift_aws_role" "k8s-core" {
  stack_id = "k8s-core"
}
```

#### Argument reference

The following arguments are supported:

- `module_id` - (Optional) - The immutable ID (slug) of the module;
- `stack_id` - (Optional) - The immutable ID (slug) of the stack;

Note that `module_id` and `stack_id` are mutually exclusive, and exactly one of them _must_ be specified.

#### Attributes reference

See the [stack AWS role resource](#spacelift_aws_role-resource) for details on the returned attributes - they are identical.

[^ Back to all resources](#resources)

### `spacelift_aws_role` resource

`spacelift_aws_role` represents [cross-account IAM role delegation](https://docs.aws.amazon.com/IAM/latest/UserGuide/tutorial_cross-account-with-roles.html) between the Spacelift worker and an individual [stack](#spacelift_stack-resource) or [module](#spacelift_module-resource). If this is set, Spacelift will use AWS STS to assume the supplied IAM role and put its temporary credentials in the runtime environment.

Note: when assuming credentials, Spacelift will use `$accountName/$stackID` as [external ID](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html) and Run ID as [session ID](https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole).

#### Example usage

```python
resource "aws_iam_role" "spacelift" {
  name = "spacelift"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [jsondecode(spacelift_stack.k8s-core.aws_assume_role_policy_statement)]
  })
}

resource "aws_iam_role_policy_attachment" "spacelift" {
  role       = aws_iam_role.spacelift.name
  policy_arn = "arn:aws:iam::aws:policy/PowerUserAccess"
}

# for a Stack:
resource "spacelift_aws_role" "k8s-core" {
  stack_id = "k8s-core"
  role_arn = aws_iam_role.spacelift.arn
}

# or for a Module:
resource "spacelift_aws_role" "k8s-core" {
  module_id = "k8s-core"
  role_arn = aws_iam_role.spacelift.arn
}
```

#### Argument reference

The following arguments are supported:

- `module_id` - (Optional) - The immutable ID (slug) of the module;
- `role_arn` - (Required) - ARN of the AWS IAM role to attach;
- `stack_id` - (Optional) - The immutable ID (slug) of the stack;

Note that `module_id` and `stack_id` are mutually exclusive, and exactly one of them _must_ be specified.

#### Attributes reference

In addition to all arguments above, the following attributes are exported:

- `id` - The immutable ID (slug) of the role;

[^ Back to all resources](#resources)

### `spacelift_gcp_service_account` data source

`spacelift_gcp_service_account` represents a Google Cloud Platform service account that's linked to a particular Stack or Module. These accounts are created by Spacelift on per-stack basis, and can be added as members to as many organizations and projects as needed. During a Run or a Task, temporary credentials for those service accounts are injected into the environment, which allows credential-less GCP Terraform provider setup.

#### Example usage

For a Module:

```python
data "spacelift_gcp_service_account" "k8s-module" {
  module_id = "k8s-module"
}
```

For a Stack:

```python
data "spacelift_gcp_service_account" "k8s-core" {
  stack_id = "k8s-core"
}
```

#### Argument reference

The following arguments are supported:

- `module_id` - (Optional) - The immutable ID (slug) of the module;
- `stack_id` - (Optional) - The immutable ID (slug) of the stack;

Note that `module_id` and `stack_id` are mutually exclusive, and exactly one of them _must_ be specified.

#### Attributes reference

See the [resource](#spacelift_gcp_service_account-resource) documentation for details on the returned attributes - they are identical.

[^ Back to all resources](#resources)

### `spacelift_gcp_service_account` resource

`spacelift_gcp_service_account` represents a Google Cloud Platform service account that's linked to a particular Stack or Module. These accounts are created by Spacelift on per-stack basis, and can be added as members to as many organizations and projects as needed. During a Run or a Task, temporary credentials for those service accounts are injected into the environment, which allows credential-less GCP Terraform provider setup.

#### Example usage

```python
resource "spacelift_stack" "k8s-core" {
  branch            = "master"
  name              = "Kubernetes core services"
  repository        = "core-infra"
}

resource "spacelift_gcp_service_account" "k8s-core" {
  stack_id = spacelift_stack.k8s-core.id

  token_scopes = [
    "https://www.googleapis.com/auth/compute",
    "https://www.googleapis.com/auth/cloud-platform",
    "https://www.googleapis.com/auth/devstorage.full_control",
  ]
}

resource "google_project" "k8s-core" {
  name       = "Kubernetes code"
  project_id = "unicorn-k8s-core"
  org_id     = var.gcp_organization_id
}

resource "google_project_iam_member" "k8s-core" {
  project = google_project.k8s-core.id
  role    = "roles/owner"
  member  = spacelift_gcp_service_account.k8s-core.service_account_email
}
```

#### Argument reference

The following arguments are supported:

- `module_id` - (Optional) - The immutable ID (slug) of the module;
- `stack_id` - (Optional) - The immutable ID (slug) of the stack;
- `token_scopes` - (Required) - List of scopes to request when generating the temporary OAuth token. At least one scope is required;

Note that `module_id` and `stack_id` are mutually exclusive, and exactly one of them _must_ be specified.

#### Attributes reference

In addition to all arguments above, the following attributes are exported:

- `id` - The immutable ID (slug) of the service account attachment;
- `service_account_email` - The email address associated with the generated GCP service account;

[^ Back to all resources](#resources)


### `spacelift_webhook` data source

`spacelift_webhook` represents a webhook endpoint to which Spacelift sends the POST request about run state changes.

#### Example usage

```python
data "spacelift_webhook" "webhook" {
  webhook_id = spacelift_webhook.webhook.id
}
```

#### Argument reference

The following arguments are supported:

- `webhook_id` - (Required) - The immutable ID (slug) of the webhook.;

#### Attributes reference

See the [webhook](#spacelift_webhook-resource) resource for details on the returned attributes - they are identical.

[^ Back to all resources](#resources)

### `spacelift_webhook` resource

`spacelift_webhook` represents a webhook endpoint to which Spacelift sends the POST request about run state changes.

#### Example usage

```python
resource "spacelift_webhook" "webhook" {
  endpoint = "https://example.com/webhooks"
  stack_id = "k8s-core"
}
```

#### Argument reference

The following arguments are supported:

- `endpoint` - (Required) - The endpoint to send the POST request to;
- `enabled` - (Optional) - The boolean which enables or disables sending webhooks (default: true);
- `module_id` - (Optional) - The ID of the module which triggers the webhooks;
- `secret` - (Optional) - The secret is used to sign each POST request so you're able to verify that the request comes from us;
- `stack_id` - (Optional) - The ID of the stack which triggers the webhooks;

Note that `module_id` and `stack_id` are mutually exclusive, and exactly one of them _must_ be specified.

#### Attributes reference

In addition to all arguments above, the following attributes are exported:

- `id` - The immutable ID (slug) of the webhook;

[^ Back to all resources](#resources)

### `spacelift_worker_pool` data source

`spacelift_worker_pool` represents a worker pool assigned to the Spacelift account.

#### Example usage

```python
data "spacelift_worker_pool" "k8s-core" {
  worker_pool_id        = "k8s-core"
}
```

#### Argument reference

The following arguments are supported:

- `worker_pool_id` - (Required) - The immutable ID (slug) of the worker pool;

#### Attributes reference

The following attributes are exported:

- `id` - The immutable ID (slug) of the worker pool;
- `config` - The credentials necessary to connect WorkerPool's workers to the control plane;
- `name` - The name of the worker pool;
- `description` - The description of the worker pool;

[^ Back to all resources](#resources)


### `spacelift_worker_pool` resource

`spacelift_worker_pool` represents a worker pool assigned to the Spacelift account.

#### Example usage

```python
resource "spacelift_worker_pool" "k8s-core" {
  name        = "Main worker"
  csr         = filebase64("/path/to/csr")
  description = "Used for all type jobs"
}
```

#### Argument reference

The following arguments are supported:

- `name` - (Required) - The name of the worker pool;
- `csr` - (Optional) - Certificate signing request (base64 format). See more [here](https://docs.spacelift.io/concepts/private-worker-pools#setting-up). If you leave it empty, the provider automatically generates `csr` and `private_key`;
- `description` - (Optional) - The description of the worker pool;


#### Attributes reference

In addition to all arguments above, the following attributes are exported:

- `id` - The immutable ID (slug) of the worker pool;
- `config` - The credentials necessary to connect WorkerPool's workers to the control plane;
- `private_key` - Automatically generated private key (base64 format) if `csr` was not provided.;

[^ Back to all resources](#resources)
