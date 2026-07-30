package spacelift

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// TestMuxedProviderSchema resolves the schema of the muxed provider exactly as Terraform
// does on startup. It is the guard for the main hazard of a gradual migration: a resource
// left registered in both provider.go and provider_framework.go makes the mux fail with a
// duplicate-resource error. That is a runtime failure, so `go build` sails past it, and
// every acceptance test that would catch it needs TF_ACC plus credentials. This one needs
// neither, so it runs on every CI job.
func TestMuxedProviderSchema(t *testing.T) {
	t.Parallel()

	server, err := testAccProtoV6MuxProviderFactories()["spacelift"]()
	if err != nil {
		t.Fatalf("could not build the muxed provider: %v", err)
	}

	resp, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("could not resolve the muxed provider schema: %v", err)
	}

	for _, diagnostic := range resp.Diagnostics {
		if diagnostic.Severity == tfprotov6.DiagnosticSeverityError {
			t.Errorf("schema error: %s: %s", diagnostic.Summary, diagnostic.Detail)
		}
	}

	// One resource from each provider, proving the mux serves both halves. Update the
	// Framework entry if spacelift_stack_dependency is ever the SDKv2 one again.
	for _, name := range []string{
		"spacelift_stack_dependency", // Plugin Framework
		"spacelift_stack",            // SDKv2
	} {
		if _, ok := resp.ResourceSchemas[name]; !ok {
			t.Errorf("%s is not served by the muxed provider", name)
		}
	}
}
