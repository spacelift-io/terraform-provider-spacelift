package spacelift

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

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
	raw, ok := d.Get(key).([]any)
	if !ok || len(raw) == 0 || raw[0] == nil {
		return nil
	}

	mapped := raw[0].(map[string]any)

	cfg := &structs.RuntimeConfigInput{
		AfterApply:    listToOptionalStringList(mapped["after_apply"]),
		AfterDestroy:  listToOptionalStringList(mapped["after_destroy"]),
		AfterInit:     listToOptionalStringList(mapped["after_init"]),
		AfterPerform:  listToOptionalStringList(mapped["after_perform"]),
		AfterPlan:     listToOptionalStringList(mapped["after_plan"]),
		AfterRun:      listToOptionalStringList(mapped["after_run"]),
		BeforeApply:   listToOptionalStringList(mapped["before_apply"]),
		BeforeDestroy: listToOptionalStringList(mapped["before_destroy"]),
		BeforeInit:    listToOptionalStringList(mapped["before_init"]),
		BeforePerform: listToOptionalStringList(mapped["before_perform"]),
		BeforePlan:    listToOptionalStringList(mapped["before_plan"]),
	}

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

	return cfg
}
