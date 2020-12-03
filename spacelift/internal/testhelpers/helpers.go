package testhelpers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/pkg/errors"
)

// AttributeCheck is a check on a single attribute.
type AttributeCheck func(map[string]string) error

// ValueCheck is a check on an attribute value.
type ValueCheck func(string) error

// Resource runs an arbitrary number of checks on a resource. The resource is
// assumed to be in the root module.
func Resource(address string, checks ...AttributeCheck) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if len(state.Modules) == 0 {
			return errors.New("no modules present")
		}

		resource, ok := state.Modules[0].Resources[address]
		if !ok {
			return errors.Errorf("resource %s not found", address)
		}

		for index, check := range checks {
			if err := check(resource.Primary.Attributes); err != nil {
				return errors.Wrapf(err, "check %d on resource %s failed", index+1, address)
			}
		}

		return nil
	}
}

// Attribute runs a value check function against an attribute passed by name.
func Attribute(name string, check ValueCheck) AttributeCheck {
	return func(attributes map[string]string) error {
		actual, ok := attributes[name]
		if !ok {
			return errors.Errorf("attribute %s not present on the resource", name)
		}

		return check(actual)
	}
}

// AttributeNotPresent ensures that an attribute is not set on the resource.
func AttributeNotPresent(name string) AttributeCheck {
	return func(attributes map[string]string) error {
		if _, ok := attributes[name]; ok {
			return errors.Errorf("attribute %s is unexpectedly present", name)
		}

		return nil
	}
}

// Equals checks for equality against the expected value.
func Equals(expected string) ValueCheck {
	return func(actual string) error {
		if actual == expected {
			return nil
		}

		return errors.Errorf("expected %q, got %q instead", expected, actual)
	}
}

// IsEmpty checks that the expected value is empty.
func IsEmpty() ValueCheck {
	return func(actual string) error {
		if actual == "" {
			return nil
		}

		return errors.Errorf("expected %q to be empty", actual)
	}
}

// IsNotEmpty checks that the expected value is not empty.
func IsNotEmpty() ValueCheck {
	return func(actual string) error {
		if actual != "" {
			return nil
		}

		return errors.Errorf("expected %q to not be empty", actual)
	}
}

// Contains checks for a patrial match match.
func Contains(needle string) ValueCheck {
	return func(haystack string) error {
		if strings.Contains(haystack, needle) {
			return nil
		}

		return errors.Errorf("expected %q to contain %q", haystack, needle)
	}
}

// StartsWith checks for a prefix match.
func StartsWith(prefix string) ValueCheck {
	return func(actual string) error {
		if strings.HasPrefix(actual, prefix) {
			return nil
		}

		return errors.Errorf("expected %q to start with %q", actual, prefix)
	}
}

// SetEquals checks for complete set equality.
func SetEquals(name string, values ...string) AttributeCheck {
	return func(attributes map[string]string) error {
		countPrefix := fmt.Sprintf("%s.#", name)

		strCount, ok := attributes[countPrefix]
		if !ok {
			return errors.Errorf("%q does not appear to be a set", name)
		}

		count, err := strconv.Atoi(strCount)
		if err != nil {
			return errors.Wrapf(err, "%q has an invalid count value %q", name, strCount)
		}

		if count != len(values) {
			return errors.Errorf("invalid %q set size - %d, expected %d", name, count, len(values))
		}

		attrValues := make(map[string]struct{})
		for attrName, attrVal := range attributes {
			if attrName == countPrefix || !strings.HasPrefix(attrName, fmt.Sprintf("%s.", name)) {
				continue
			}

			attrValues[attrVal] = struct{}{}
		}

		for _, value := range values {
			if _, ok := attrValues[value]; !ok {
				return errors.Errorf("value %q not found in set %q", value, name)
			}
		}

		return nil
	}
}
