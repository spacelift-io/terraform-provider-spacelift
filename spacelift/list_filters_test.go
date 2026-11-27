package spacelift

import "testing"

func TestNamePattern(t *testing.T) {
	for _, tc := range []struct {
		pattern, name string
		want          bool
	}{
		{"", "anything", true},
		{"prod-*", "Prod-api", true},
		{"prod-*", "my-prod-api", false},
		{"*api", "prod-API", true},
		{"a.b", "aXb", false},
		{"a.b", "A.B", true},
		{"*", "", true},
		{"prod", "production", false},
	} {
		if got := namePattern(tc.pattern).MatchString(tc.name); got != tc.want {
			t.Errorf("pattern %q, name %q: got %v, want %v", tc.pattern, tc.name, got, tc.want)
		}
	}
}

func TestListFilterMatches(t *testing.T) {
	// root > a > a1 > a2, and root > b.
	spaces := spaceTree{"root": "", "a": "root", "a1": "a", "a2": "a1", "b": "root"}

	for _, tc := range []struct {
		desc   string
		filter listFilter
		space  string
		labels []string
		want   bool
	}{
		{"no filter", listFilter{}, "b", nil, true},
		{"name matches", listFilter{name: "N*"}, "b", nil, true},
		{"name differs", listFilter{name: "x*"}, "b", nil, false},
		{"same space", listFilter{spaceID: "a"}, "a", nil, true},
		{"child space", listFilter{spaceID: "a"}, "a1", nil, false},
		{"child space included", listFilter{spaceID: "a", includeChildSpaces: true}, "a2", nil, true},
		{"other branch", listFilter{spaceID: "a", includeChildSpaces: true}, "b", nil, false},
		{"parent space", listFilter{spaceID: "a1", includeChildSpaces: true}, "a", nil, false},
		{"all labels", listFilter{labels: []string{"x", "y"}}, "b", []string{"y", "z", "x"}, true},
		{"missing label", listFilter{labels: []string{"x", "y"}}, "b", []string{"x"}, false},
	} {
		tc.filter.namePattern = namePattern(tc.filter.name)
		tc.filter.spaces = spaces
		if got := tc.filter.matches("Name", tc.space, tc.labels); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.desc, got, tc.want)
		}
	}
}

func TestSpaceTreePath(t *testing.T) {
	spaces := spaceTree{"root": "", "a": "root", "a1": "a"}
	assertStrings(t, []string{spaces.path("a1"), spaces.path("root")}, "root/a/a1", "root")
}
