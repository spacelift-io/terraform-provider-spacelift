# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands
- Build project: `go build`
- Build with GoReleaser (local dev): `goreleaser build --clean --snapshot`
- Generate documentation: `cd tools && go generate ./...`

## Test Commands
- The test relies on the `test.env` file for environment variables. This file should be created in the root of the repository and contain the necessary environment variables for testing.
- Run single test: `go test -v ./spacelift -run TestFunctionName`

## Lint Commands
- Format check: `gofmt -s -l -d .`
- Run linter: `golangci-lint run`

## Code Style Guidelines
- **Imports**: Use standard library first, then default packages, then local packages (prefix: `github.com/spacelift-io/terraform-provider-spacelift`)
- **Formatting**: Follow standard Go format with `gofmt -s` (simplified mode)
- **Error Handling**: Use `github.com/pkg/errors` for error wrapping/annotation
- **Types**: Follow standard Go conventions and use Terraform types from `github.com/hashicorp/terraform-plugin-sdk/v2`
