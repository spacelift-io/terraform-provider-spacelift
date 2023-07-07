package framework

import (
	"context"

	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
)

func verifyModule(ctx context.Context, moduleID string, client *internal.Client) error {
	var query struct {
		Module *structs.Module `graphql:"module(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(moduleID)}

	if err := client.Query(ctx, "ModuleVerifyExistence", &query, variables); err != nil {
		return errors.Wrap(err, "could not query for module")
	}

	if query.Module == nil {
		return errors.New("module not found")
	}

	return nil
}

func verifyStack(ctx context.Context, stackID string, client *internal.Client) error {
	var query struct {
		Stack *structs.Stack `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.ID(stackID)}

	if err := client.Query(ctx, "StackVerifyExistence", &query, variables); err != nil {
		return errors.Wrap(err, "could not query for stack")
	}

	if query.Stack == nil {
		return errors.New("stack not found")
	}

	return nil
}
