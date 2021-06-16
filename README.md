# Spacelift Terraform provider

The Spacelift Terraform provider is used to programmatically interact with its GraphQL API, allowing Spacelift to declaratively manage itself 🤯

## Documentation

You can browse documentation on the [Terraform provider registry](https://registry.terraform.io/providers/spacelift-io/spacelift/latest/docs).

## Using the Provider

### Terraform 0.13 and above

You can use the provider via the [Terraform provider registry](https://registry.terraform.io/providers/spacelift-io/spacelift/latest).

### Terraform 0.12 or manual installation

You can download a pre-built binary from the [releases](https://github.com/spacelift-io/terraform-provider-spacelift/releases/) page, these are built using [goreleaser](https://goreleaser.com/) (the [configuration](.goreleaser.yml) is in the repo). You can verify the signature using [this key](https://keys.openpgp.org/vks/v1/by-fingerprint/175FD97AD2358EFE02832978E302FB5AA29D88F7).

If you want to build from source, you can simply use `go build` in the root of the repository.
