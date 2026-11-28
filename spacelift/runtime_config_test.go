package spacelift

import (
	"context"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testRunResourceDataWithRawConfig(t *testing.T, runtimeConfig map[string]cty.Value) *schema.ResourceData {
	t.Helper()

	r := resourceRun()
	configSchema := r.CoreConfigSchema()

	attributes := map[string]cty.Value{
		"stack_id": cty.StringVal("stack-id"),
	}

	if runtimeConfig != nil {
		blockSchema := configSchema.BlockTypes["runtime_config"].Block
		blockAttrs := make(map[string]cty.Value, len(blockSchema.Attributes)+len(blockSchema.BlockTypes))
		for name, attr := range blockSchema.Attributes {
			if val, ok := runtimeConfig[name]; ok {
				blockAttrs[name] = val
			} else {
				blockAttrs[name] = cty.NullVal(attr.Type)
			}
		}
		for name, block := range blockSchema.BlockTypes {
			if val, ok := runtimeConfig[name]; ok {
				blockAttrs[name] = val
			} else {
				blockAttrs[name] = cty.NullVal(block.Block.ImpliedType())
			}
		}
		attributes["runtime_config"] = cty.ListVal([]cty.Value{cty.ObjectVal(blockAttrs)})
	}

	for name, attr := range configSchema.Attributes {
		if _, ok := attributes[name]; !ok {
			attributes[name] = cty.NullVal(attr.Type)
		}
	}
	for name, block := range configSchema.BlockTypes {
		if _, ok := attributes[name]; !ok {
			attributes[name] = cty.NullVal(block.Block.ImpliedType())
		}
	}

	configValue := cty.ObjectVal(attributes)
	diff, err := r.Diff(context.Background(), &terraform.InstanceState{}, terraform.NewResourceConfigShimmed(configValue, configSchema), nil)
	if err != nil {
		t.Fatalf("could not diff resource: %v", err)
	}
	diff.RawConfig = configValue

	d, err := schema.InternalMap(r.Schema).Data(nil, diff)
	if err != nil {
		t.Fatalf("could not build ResourceData: %v", err)
	}
	return d
}

func TestParseRuntimeConfigInput(t *testing.T) {
	t.Run("omitted runtime_config returns nil", func(t *testing.T) {
		d := testRunResourceDataWithRawConfig(t, nil)
		if got := parseRuntimeConfigInput(d, "runtime_config"); got != nil {
			t.Fatalf("expected nil, got %+v", got)
		}
	})

	t.Run("omitted hook lists stay nil when runtime_config is set", func(t *testing.T) {
		d := testRunResourceDataWithRawConfig(t, map[string]cty.Value{
			"environment": cty.SetVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"key":   cty.StringVal("PR884_TEST"),
					"value": cty.StringVal("runtime-config-present"),
				}),
			}),
		})

		got := parseRuntimeConfigInput(d, "runtime_config")
		if got == nil {
			t.Fatal("expected non-nil RuntimeConfigInput")
		}
		if got.Environment == nil || len(*got.Environment) != 1 || (*got.Environment)[0].Key != "PR884_TEST" {
			t.Fatalf("expected environment to be populated, got %+v", got.Environment)
		}

		if got.AfterApply != nil || got.AfterDestroy != nil || got.AfterInit != nil ||
			got.AfterPerform != nil || got.AfterPlan != nil || got.AfterRun != nil ||
			got.BeforeApply != nil || got.BeforeDestroy != nil || got.BeforeInit != nil ||
			got.BeforePerform != nil || got.BeforePlan != nil {
			t.Fatalf("expected all omitted hook lists to be nil, got %+v", got)
		}
	})

	t.Run("explicit empty list sends empty slice while omitted hooks stay nil", func(t *testing.T) {
		d := testRunResourceDataWithRawConfig(t, map[string]cty.Value{
			"after_apply": cty.ListValEmpty(cty.String),
			"before_plan": cty.ListValEmpty(cty.String),
			"after_plan":  cty.ListVal([]cty.Value{cty.StringVal("echo ok")}),
		})

		got := parseRuntimeConfigInput(d, "runtime_config")
		if got == nil {
			t.Fatal("expected non-nil RuntimeConfigInput")
		}
		if got.AfterApply == nil || len(*got.AfterApply) != 0 {
			t.Fatalf("expected AfterApply to be empty slice, got %+v", got.AfterApply)
		}
		if got.BeforePlan == nil || len(*got.BeforePlan) != 0 {
			t.Fatalf("expected BeforePlan to be empty slice, got %+v", got.BeforePlan)
		}
		if got.AfterPlan == nil || len(*got.AfterPlan) != 1 || (*got.AfterPlan)[0] != "echo ok" {
			t.Fatalf("expected AfterPlan to be [echo ok], got %+v", got.AfterPlan)
		}
		if got.BeforeInit != nil || got.AfterInit != nil || got.BeforeApply != nil {
			t.Fatalf("expected omitted hooks to stay nil, got %+v", got)
		}
	})

	t.Run("runtime_config with only empty hook lists is preserved", func(t *testing.T) {
		d := testRunResourceDataWithRawConfig(t, map[string]cty.Value{
			"before_init": cty.ListValEmpty(cty.String),
			"after_init":  cty.ListValEmpty(cty.String),
		})

		got := parseRuntimeConfigInput(d, "runtime_config")
		if got == nil {
			t.Fatal("expected non-nil RuntimeConfigInput when only empty hook lists are set")
		}
		if got.BeforeInit == nil || len(*got.BeforeInit) != 0 {
			t.Fatalf("expected BeforeInit to be empty slice, got %+v", got.BeforeInit)
		}
		if got.AfterInit == nil || len(*got.AfterInit) != 0 {
			t.Fatalf("expected AfterInit to be empty slice, got %+v", got.AfterInit)
		}
		if got.AfterPlan != nil || got.BeforePlan != nil || got.AfterApply != nil || got.BeforeApply != nil {
			t.Fatalf("expected omitted hooks to stay nil, got %+v", got)
		}
	})

	t.Run("empty runtime_config block with no attributes returns nil", func(t *testing.T) {
		d := testRunResourceDataWithRawConfig(t, map[string]cty.Value{})
		if got := parseRuntimeConfigInput(d, "runtime_config"); got != nil {
			t.Fatalf("expected nil, got %+v", got)
		}
	})
}
