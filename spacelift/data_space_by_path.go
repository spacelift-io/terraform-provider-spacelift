package spacelift

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/validations"
)

func dataSpaceByPath() *schema.Resource {
	return &schema.Resource{
		Description: "`spacelift_space_by_path` represents a Spacelift **space** - " +
			"a collection of resources such as stacks, modules, policies, etc. Allows for more granular access control. Can have a parent space. In contrary to `spacelift_space`, this resource is identified by a path, not by an ID. " +
			"For this data source to work, path must be unique. If there are multiple spaces with the same path, this datasource will fail. \n" +
			"This data source can be used either with absolute paths (starting with root) or relative paths. When using a relative path, the path is relative to the current run's space. \n" +
			"**Disclaimer:** \n" +
			"This datasource can only be used in a stack that resides in a space with inheritance enabled. In addition, the parent spaces (excluding root) must also have inheritance enabled.",

		ReadContext: dataSpaceByPathRead,

		Schema: map[string]*schema.Schema{
			"space_path": {
				Type:             schema.TypeString,
				Description:      "path to the space - a series of space names separated by `/`",
				Required:         true,
				ValidateDiagFunc: validations.DisallowEmptyString,
			},
			"parent_space_id": {
				Type:        schema.TypeString,
				Description: "immutable ID (slug) of parent space",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "free-form space description for users",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "name of the space",
				Computed:    true,
			},
			"inherit_entities": {
				Type:        schema.TypeBool,
				Description: "indication whether access to this space inherits read access to entities from the parent space",
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "list of labels describing a space",
				Computed:    true,
			},
		},
	}
}

func dataSpaceByPathRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	path := d.Get("space_path").(string)

	if strings.HasPrefix(path, "/") {
		return diag.Errorf("path must not start with a slash")
	}

	var query struct {
		Spaces []*structs.Space `graphql:"spaces"`
	}

	if err := meta.(*internal.Client).Query(ctx, "SpaceRead", &query, map[string]interface{}{}); err != nil {
		return diag.Errorf("could not query for spaces: %v", err)
	}

	startingSpace := "root"
	if !strings.HasPrefix(path, "root/") && path != "root" {
		// if path does not start with root, we think it's a relative path. In this case it's relative to the current space the spacelift run is in

		stackID, err := getStackIDFromToken(meta.(*internal.Client).Token)
		if err != nil {
			return diag.Errorf("couldn't identify the run: %v", err)
		}

		space, err := getSpaceForStack(ctx, stackID, meta)
		if err != nil {
			return diag.Errorf("couldn't determine current space: %v", err)
		}

		startingSpace = space.ID
		path = space.Name + "/" + path // to be consistent with full path search where root is always included in the path
	}

	space, err := findSpaceByPath(query.Spaces, path, startingSpace)
	if err != nil {
		return diag.Errorf("error while traversing space path: %v", err)
	}

	d.SetId(space.ID)
	d.Set("name", space.Name)
	d.Set("description", space.Description)
	d.Set("inherit_entities", space.InheritEntities)

	labels := schema.NewSet(schema.HashString, []interface{}{})
	for _, label := range space.Labels {
		labels.Add(label)
	}
	d.Set("labels", labels)

	if space.ParentSpace != nil {
		d.Set("parent_space_id", *space.ParentSpace)
	}

	return nil
}

func findSpaceByPath(spaces []*structs.Space, path, startingSpace string) (*structs.Space, error) {
	childrenMap := make(map[string][]*structs.Space, len(spaces))
	var currentSpace *structs.Space

	for _, space := range spaces {
		if space.ID == startingSpace {
			currentSpace = space
		}
		if space.ParentSpace != nil {
			childrenMap[*space.ParentSpace] = append(childrenMap[*space.ParentSpace], space)
		}
	}

	if currentSpace == nil {
		return nil, fmt.Errorf("%v space not found", startingSpace)
	}

	pathSplit := strings.Split(path, "/")

	for i := 1; i < len(pathSplit); i++ {
		nameToLookFor := pathSplit[i]
		currentChildren := childrenMap[currentSpace.ID]

		found := false
		for _, child := range currentChildren {
			if child.Name == nameToLookFor {
				if found {
					return nil, fmt.Errorf("path %s is ambiguous", strings.Join(pathSplit[:i+1], "/"))
				}
				currentSpace = child
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("space does not exist")
		}
	}

	return currentSpace, nil
}
