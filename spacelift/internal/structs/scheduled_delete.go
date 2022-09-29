package structs

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"
)

// ScheduledStackDelete represents a scheduled delete.
type ScheduledStackDelete struct {
	ID                    string
	ShouldDeleteResources graphql.Boolean `json:"shouldDeleteResources"`
	TimestampSchedule     graphql.Int     `json:"timestampSchedule"`
}

// ScheduledDeleteInput represents the input for creating/updating a scheduled stack_delete.
type ScheduledDeleteInput struct {
	ShouldDeleteResources graphql.Boolean `json:"shouldDeleteResources"`
	TimestampSchedule     *graphql.Int    `json:"timestampSchedule"`
}

func PopulateDeleteStackSchedule(d *schema.ResourceData, sd *ScheduledStackDelete) error {
	if err := d.Set("delete_resources", sd.ShouldDeleteResources); err != nil {
		return errors.Wrap(err, "could not set \"delete_resources\" attribute")
	}

	if err := d.Set("at", sd.TimestampSchedule); err != nil {
		return errors.Wrap(err, "could not set \"at\" attribute")
	}

	return nil
}
