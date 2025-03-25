package structs

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"
)

type TaskInput struct {
	StackID string
	Command string
	Init    bool
	Wait    WaitConfiguration
}

type Task struct {
	ID graphql.ID `json:"id"`
}

func NewTaskInput(d *schema.ResourceData) (*TaskInput, error) {
	cfg := &TaskInput{}

	command, ok := d.GetOk("command")
	if ok && command != nil {
		cfg.Command = command.(string)
	}

	init, ok := d.GetOk("init")
	if ok && init != nil {
		cfg.Init = init.(bool)
	}

	stackID, ok := d.GetOk("stack_id")
	if ok && stackID != nil {
		cfg.StackID = stackID.(string)
	}

	return cfg, nil
}
