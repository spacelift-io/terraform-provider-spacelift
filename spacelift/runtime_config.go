package spacelift

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

// runtimeConfigInputSchema returns the nested-block schema for the user-
// configurable fields of a runtime_config. When forceNew is true (e.g., for
// immutable resources like spacelift_run), every field is marked ForceNew.
func runtimeConfigInputSchema(forceNew bool) map[string]*schema.Schema {
	stringList := func(description string) *schema.Schema {
		return &schema.Schema{
			Type:        schema.TypeList,
			Description: description,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
			ForceNew:    forceNew,
		}
	}

	return map[string]*schema.Schema{
		"project_root": {
			Type:        schema.TypeString,
			Description: "Project root is the optional directory relative to the workspace root containing the entrypoint to the Stack.",
			Optional:    true,
			ForceNew:    forceNew,
		},
		"runner_image": {
			Type:        schema.TypeString,
			Description: "Name of the Docker image used to process Runs",
			Optional:    true,
			ForceNew:    forceNew,
		},
		"environment": {
			Type:        schema.TypeSet,
			Description: "Environment variables for the run",
			Optional:    true,
			ForceNew:    forceNew,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"key": {
						Type:        schema.TypeString,
						Description: "Environment variable key",
						Required:    true,
					},
					"value": {
						Type:        schema.TypeString,
						Description: "Environment variable value",
						Required:    true,
					},
				},
			},
		},
		"after_apply":    stringList("List of after-apply scripts"),
		"after_destroy":  stringList("List of after-destroy scripts"),
		"after_init":     stringList("List of after-init scripts"),
		"after_perform":  stringList("List of after-perform scripts"),
		"after_plan":     stringList("List of after-plan scripts"),
		"after_run":      stringList("List of after-run scripts"),
		"before_apply":   stringList("List of before-apply scripts"),
		"before_destroy": stringList("List of before-destroy scripts"),
		"before_init":    stringList("List of before-init scripts"),
		"before_perform": stringList("List of before-perform scripts"),
		"before_plan":    stringList("List of before-plan scripts"),
	}
}

// parseRuntimeConfigInput reads the runtime_config nested block from the
// given ResourceData and returns the corresponding GraphQL input, or nil if
// the block is not set.
func parseRuntimeConfigInput(d *schema.ResourceData, key string) *structs.RuntimeConfigInput {
	var mapped map[string]any
	if raw, ok := d.Get(key).([]any); ok && len(raw) > 0 && raw[0] != nil {
		mapped = raw[0].(map[string]any)
	}

	rawBlock, hasRawBlock := runtimeConfigRawBlock(d, key)
	if mapped == nil && !hasRawBlock {
		return nil
	}

	optionalList := func(field string) *[]graphql.String {
		if hasRawBlock {
			if !rawBlock.Type().HasAttribute(field) || rawBlock.GetAttr(field).IsNull() {
				return nil
			}
			if mapped != nil && mapped[field] != nil {
				return listToOptionalStringList(mapped[field])
			}
			// When every attribute in runtime_config is empty (e.g. only []
			// hook lists), SDK v2's d.Get(key) collapses the block to [].
			// The raw attribute is non-null, so return an empty slice.
			return listToOptionalStringList([]any{})
		}
		// Without raw config (e.g. TestResourceDataRaw), SDK v2 materializes
		// omitted TypeList fields as empty slices, so only non-empty lists can
		// be distinguished from omitted ones.
		if list, ok := mapped[field].([]any); !ok || len(list) == 0 {
			return nil
		}
		return listToOptionalStringList(mapped[field])
	}

	cfg := &structs.RuntimeConfigInput{
		AfterApply:    optionalList("after_apply"),
		AfterDestroy:  optionalList("after_destroy"),
		AfterInit:     optionalList("after_init"),
		AfterPerform:  optionalList("after_perform"),
		AfterPlan:     optionalList("after_plan"),
		AfterRun:      optionalList("after_run"),
		BeforeApply:   optionalList("before_apply"),
		BeforeDestroy: optionalList("before_destroy"),
		BeforeInit:    optionalList("before_init"),
		BeforePerform: optionalList("before_perform"),
		BeforePlan:    optionalList("before_plan"),
	}

	if mapped != nil {
		if envSet, ok := mapped["environment"].(*schema.Set); ok && envSet.Len() > 0 {
			environment := make([]structs.EnvVarInput, 0, envSet.Len())
			for _, e := range envSet.List() {
				envMap := e.(map[string]any)
				environment = append(environment, structs.EnvVarInput{
					Key:   toString(envMap["key"]),
					Value: toString(envMap["value"]),
				})
			}
			cfg.Environment = &environment
		}

		if projectRoot, ok := mapped["project_root"].(string); ok && projectRoot != "" {
			cfg.ProjectRoot = toOptionalString(projectRoot)
		}

		if runnerImage, ok := mapped["runner_image"].(string); ok && runnerImage != "" {
			cfg.RunnerImage = toOptionalString(runnerImage)
		}
	}

	if cfg.Environment == nil && cfg.ProjectRoot == nil && cfg.RunnerImage == nil &&
		cfg.AfterApply == nil && cfg.AfterDestroy == nil && cfg.AfterInit == nil &&
		cfg.AfterPerform == nil && cfg.AfterPlan == nil && cfg.AfterRun == nil &&
		cfg.BeforeApply == nil && cfg.BeforeDestroy == nil && cfg.BeforeInit == nil &&
		cfg.BeforePerform == nil && cfg.BeforePlan == nil {
		return nil
	}

	return cfg
}

func runtimeConfigRawBlock(d *schema.ResourceData, key string) (cty.Value, bool) {
	rawConfig := d.GetRawConfig()
	if rawConfig.IsNull() || !rawConfig.IsKnown() || !rawConfig.Type().IsObjectType() || !rawConfig.Type().HasAttribute(key) {
		return cty.NilVal, false
	}

	blockList := rawConfig.GetAttr(key)
	if blockList.IsNull() || !blockList.IsKnown() || blockList.LengthInt() == 0 {
		return cty.NilVal, false
	}

	blockVal := blockList.Index(cty.NumberIntVal(0))
	if blockVal.IsNull() || !blockVal.IsKnown() || !blockVal.Type().IsObjectType() {
		return cty.NilVal, false
	}

	return blockVal, true
}
