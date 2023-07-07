package spacelift

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/structs"
	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestSpaceByPathData(t *testing.T) {
	t.Run("creates and reads a space", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		spaceName := fmt.Sprintf("My first space %s", randomID)
		testSteps(t, []resource.TestStep{{
			Config: fmt.Sprintf(`
				resource "spacelift_space" "test" {
					name = "%s"
					inherit_entities = true
					parent_space_id = "root"
					description = "some valid description"
					labels = ["label1", "label2"]
				}
	
				data "spacelift_space_by_path" "test" {
					space_path = "root/%s"
					depends_on = [spacelift_space.test]
				}
			`, spaceName, spaceName),
			Check: Resource(
				"data.spacelift_space_by_path.test",
				Attribute("id", Contains("my-first-space")),
				Attribute("parent_space_id", Equals("root")),
				Attribute("description", Equals("some valid description")),
				SetEquals("labels", "label1", "label2"),
			),
		}})
	})

	t.Run("invalid space path should return an error", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: `
					data "spacelift_space_by_path" "test" {
						space_path = "root123"
					}
				`,
				ExpectError: regexp.MustCompile("space path must start with `root`"),
			},
			{
				Config: `
					data "spacelift_space_by_path" "test" {
						space_path = "test123/test"
					}
				`,
				ExpectError: regexp.MustCompile("space path must start with `root`"),
			},
		})
	})
}

func Test_findSpaceByPath(t *testing.T) {
	type args struct {
		spaces []*structs.Space
		path   string
	}

	var root = &structs.Space{
		ID:   "root",
		Name: "root",
	}
	var rootChild = &structs.Space{
		ID:          "rootChild-randomsuffix1",
		Name:        "rootChild",
		ParentSpace: &root.ID,
	}
	var rootChild2 = &structs.Space{
		ID:          "rootChild2-randomsuffix2",
		Name:        "rootChild2",
		ParentSpace: &root.ID,
	}
	var rootGrandchild = &structs.Space{
		ID:          "rootGrandchild-randomsuffix3",
		Name:        "rootGrandchild",
		ParentSpace: &rootChild.ID,
	}

	var rootChildSameName = &structs.Space{
		ID:          "rootChild-randomsuffix4",
		Name:        "rootChild",
		ParentSpace: &root.ID,
	}

	tests := []struct {
		name    string
		args    args
		want    *structs.Space
		wantErr bool
	}{
		{
			name: "just root should be found",
			args: args{
				spaces: []*structs.Space{
					root,
				},
				path: "root",
			},
			want:    root,
			wantErr: false,
		},
		{
			name: "root child should be found",
			args: args{
				spaces: []*structs.Space{
					root,
					rootChild,
					rootChild2,
				},
				path: "root/rootChild",
			},
			want:    rootChild,
			wantErr: false,
		},
		{
			name: "root child should not be found if name is ambiguous",
			args: args{
				spaces: []*structs.Space{
					root,
					rootChild,
					rootChild2,
					rootChildSameName,
				},
				path: "root/rootChild",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "root grandchild should be found",
			args: args{
				spaces: []*structs.Space{
					root,
					rootChild,
					rootChild2,
					rootGrandchild,
				},
				path: "root/rootChild/rootGrandchild",
			},
			want:    rootGrandchild,
			wantErr: false,
		},
		{
			name: "invalid path should return error",
			args: args{
				spaces: []*structs.Space{
					root,
					rootChild,
					rootChild2,
					rootGrandchild,
				},
				path: "root/rootGrandchild",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findSpaceByPath(tt.args.spaces, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("findSpaceByPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findSpaceByPath() got = %v, want %v", got, tt.want)
			}
		})
	}
}
