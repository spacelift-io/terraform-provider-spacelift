package testhelpers

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/pkg/errors"
	"strings"
)

// AttributeCheck is a check on a single attribute.
type AttributeCheck func(map[string]string) error

// ValueCheck is a check on an attribute value.
type ValueCheck func(string) error

// Equals checks for equality against the expected value.
func Equals(expected string) resource.CheckResourceAttrWithFunc {
	return func(actual string) error {
		if actual == expected {
			return nil
		}

		return errors.Errorf("expected %q, got %q instead", expected, actual)
	}
}

// IsEmpty checks that the expected value is empty.
func IsEmpty() resource.CheckResourceAttrWithFunc {
	return func(actual string) error {
		if actual == "" {
			return nil
		}

		return errors.Errorf("expected %q to be empty", actual)
	}
}

// IsNotEmpty checks that the expected value is not empty.
func IsNotEmpty() resource.CheckResourceAttrWithFunc {
	return func(actual string) error {
		if actual != "" {
			return nil
		}

		return errors.Errorf("expected %q to not be empty", actual)
	}
}

// Contains checks for a partial match.
func Contains(needle string) resource.CheckResourceAttrWithFunc {
	return func(haystack string) error {
		if strings.Contains(haystack, needle) {
			return nil
		}

		return errors.Errorf("expected %q to contain %q", haystack, needle)
	}
}
