package spacelift

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// The SDK's unknown-value sentinel; the constant naming it is in an internal package.
const unknownValue = "74D93920-ED26-11E3-AC10-0800200C9A66"

func TestValidateSpaceliftRepoVCS(t *testing.T) {
	config := func(mutate func(map[string]any)) map[string]any {
		raw := map[string]any{
			"name":               "test",
			"repository":         "my-repo",
			"branch":             "main",
			"space_id":           "root",
			"terraform_provider": "default",
			"spacelift":          []any{map[string]any{"id": "my-repo"}},
		}
		mutate(raw)
		return raw
	}

	for _, tc := range []struct {
		name    string
		mutate  func(map[string]any)
		wantErr string
	}{
		{
			name:   "matching repository and repo id",
			mutate: func(map[string]any) {},
		},
		{
			name:    "mismatched repository",
			mutate:  func(raw map[string]any) { raw["repository"] = "not-the-repo" },
			wantErr: `repository must be the Spacelift repo ID (slug) "my-repo"`,
		},
		{
			name:   "unknown repo id",
			mutate: func(raw map[string]any) { raw["spacelift"] = []any{map[string]any{"id": unknownValue}} },
		},
		{
			name:   "unknown repository",
			mutate: func(raw map[string]any) { raw["repository"] = unknownValue },
		},
		{
			name:   "unknown branch",
			mutate: func(raw map[string]any) { raw["branch"] = unknownValue },
		},
		{
			name:    "branch other than main",
			mutate:  func(raw map[string]any) { raw["branch"] = "develop" },
			wantErr: `branch must be "main" when using a Spacelift repo`,
		},
		{
			name:   "branch is only constrained by the spacelift block",
			mutate: func(raw map[string]any) { delete(raw, "spacelift"); raw["branch"] = "develop" },
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := resourceModule().Diff(context.Background(), nil, terraform.NewResourceConfigRaw(config(tc.mutate)), nil)

			switch {
			case tc.wantErr == "" && err != nil:
				t.Fatalf("expected no error, got %v", err)
			case tc.wantErr != "" && err == nil:
				t.Fatalf("expected an error containing %q, got none", tc.wantErr)
			case tc.wantErr != "" && !strings.Contains(err.Error(), tc.wantErr):
				t.Fatalf("expected an error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}
