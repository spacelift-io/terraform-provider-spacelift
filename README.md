# Spacelift Terraform Provider

The Spacelift Terraform provider is used to programmatically interact with its GraphQL API, allowing Spacelift to declaratively manage itself 🤯

## Documentation

You can browse the documentation on the following registries:
- [OpenTofu registry](https://search.opentofu.org/provider/spacelift-io/spacelift/)
- [Terraform registry](https://registry.terraform.io/providers/spacelift-io/spacelift/latest/docs)

## Using the Provider

### Terraform 0.13 and Above

You can use the provider via the [Terraform provider registry](https://registry.terraform.io/providers/spacelift-io/spacelift/latest).

### Terraform 0.12 or Manual Installation

You can download a pre-built binary from the [releases](https://github.com/spacelift-io/terraform-provider-spacelift/releases/) page, these are built using [goreleaser](https://goreleaser.com/) (the [configuration](.goreleaser.yml) is in the repo). You can verify the signature using [this key](https://keys.openpgp.org/vks/v1/by-fingerprint/175FD97AD2358EFE02832978E302FB5AA29D88F7).

If you want to build from source, you can simply use `go build` in the root of the repository.

## Development

### Tools

To develop the provider locally you need the following tools:

- [Go](https://go.dev/doc/install) - see [go.mod](go.mod) for the proper version
- [GoReleaser](https://goreleaser.com/) - minimum v2.0.0
- A Spacelift account to use for testing.

### Generating the Documentation

To generate the documentation, run the following command:

```shell
cd tools
go generate ./...
```

### Using a Local Build of the Provider

Sometimes as well as running unit tests, you want to be able to run a local build of the provider against Spacelift.
This involves the following steps:

1. Building a copy of the provider using GoReleaser.
2. Updating your .terraformrc file to point at your local build.
3. Generating an API key in Spacelift.
4. Running Terraform locally.

#### Building the Provider Using GoReleaser

To build the provider, run the following command:

```shell
goreleaser build --clean --snapshot
```

This will produce a number of binaries in subfolders of the `dist` folder for each supported
architecture and OS:

```text
dist
|-- artifacts.json
|-- config.yaml
|-- metadata.json
|-- terraform-provider-spacelift_darwin_amd64_v1
|   `-- terraform-provider-spacelift_v0.1.11-SNAPSHOT-bb215e9
|-- terraform-provider-spacelift_darwin_arm64
|   `-- terraform-provider-spacelift_v0.1.11-SNAPSHOT-bb215e9
|-- terraform-provider-spacelift_linux_amd64_v1
|   `-- terraform-provider-spacelift_v0.1.11-SNAPSHOT-bb215e9
|-- terraform-provider-spacelift_linux_arm64
|   `-- terraform-provider-spacelift_v0.1.11-SNAPSHOT-bb215e9
|-- terraform-provider-spacelift_windows_amd64_v1
|   `-- terraform-provider-spacelift_v0.1.11-SNAPSHOT-bb215e9.exe
`-- terraform-provider-spacelift_windows_arm64
    `-- terraform-provider-spacelift_v0.1.11-SNAPSHOT-bb215e9.exe
```

#### Updating your .terraformrc file

The next step is telling Terraform to use your local build, rather than a copy from the Terraform
registry. You can do this by specifying [dev_overrides](https://www.terraform.io/cli/config/config-file#development-overrides-for-provider-developers)
in your `.terraformrc` file.

To do this, edit or create a `.terraformrc` in your home folder, and add the following contents:

```hcl
provider_installation {
  dev_overrides {
    "spacelift.io/spacelift-io/spacelift" = "<absolute-path-to-repo>/dist/terraform-provider-spacelift_<OS>_<arch>"
  }

  direct {}
}
```

Make sure to replace `<absolute-path-to-repo>`, `<OS>`, and `<arch>` with the correct values, for example:

```hcl
"spacelift.io/spacelift-io/spacelift" = "/home/my-user/github.com/spacelift-io/terraform-provider-spacelift/dist/terraform-provider-spacelift_linux_amd64_v1"
```

#### Generating an API Key

Follow the information in our [API documentation page](https://docs.spacelift.io/integrations/api) to generate an API key.
Please make sure to generate an admin key since admin permissions are required for most operations
you will be using the provider for.

#### Running Spacelift Terraform Provider Locally

To test your local build, just create the relevant Terraform files needed to test your changes,
and run `terraform plan`, `terraform apply`, etc as normal. The main difference when running
the provider locally rather than within Spacelift is that you need to tell it how to authenticate
with your Spacelift account. Here's a minimal example:

```hcl
terraform {
  required_providers {
    spacelift = {
      source = "spacelift.io/spacelift-io/spacelift"
    }
  }
}

provider "spacelift" {
  api_key_endpoint = "https://<account-name>.app.spacelift.io"
  api_key_id       = "<api-key-id>"
  api_key_secret   = "<api-key-secret>"
}

data "spacelift_account" "this" {
}

output "account_name" {
  value = data.spacelift_account.this.name
}
```

Make sure to replace `<account-name>`, `<api-key-id>` and `<api-key-secret>` with the relevant values.

### Releasing New Versions of the Provider

In order to release a new version of the provider one should follow those simple steps:

- Create a new tag for the latest commit on tha main branch `git tag vX.Y.Z -a -m "Release"`
- Push the tag `git push origin vX.Y.Z`
- Refer to our [internal wiki](https://www.notion.so/spacelift/Spacelift-Terraform-Provider-18cf11e5c8ad4a44bf6395cd69a744b7#1540245e247545bcb38e03b1050d6032) on publishing the release artifacts
