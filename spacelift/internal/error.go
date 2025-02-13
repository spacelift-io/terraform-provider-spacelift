package internal

import (
	"errors"
	"fmt"
	"strings"

	"github.com/shurcooL/graphql"
)

// FromSpaceliftError wraps the error with a helpful message when encountering a Spacelift error.
// In this case an unauthorized error.
func FromSpaceliftError(err error) error {
	if err != nil && strings.Contains(err.Error(), "unauthorized") {
		return fmt.Errorf("%w - Is it an administrative stack in the appropriate space? Additionally, have you ensured that you provided the correct space ID rather than the space name?", err)
	}

	var graphErrs graphql.GraphQLErrors
	if errors.As(err, &graphErrs) {
		return parseGraphqlErrors(graphErrs)
	}

	return err
}

func parseGraphqlErrors(graphErrs graphql.GraphQLErrors) error {
	errorParts := make([]string, 0, len(graphErrs))
	for _, err := range graphErrs {
		errorPart := err.Message

		if len(err.Extensions) > 0 {
			errorPart = fmt.Sprintf("%s: %s", errorPart, parseExtensions(err.Extensions))
		}

		errorParts = append(errorParts, errorPart)
	}

	return errors.New(strings.Join(errorParts, ", "))
}

func parseExtensions(ext map[string]interface{}) string {
	errorParts := make([]string, 0, len(ext))
	for k, v := range ext {
		errorParts = append(errorParts, fmt.Sprintf("%s: %v", k, v))
	}

	return strings.Join(errorParts, ", ")
}

// AsError is an inline form of errors.As.
func AsError[TError error](err error) (TError, bool) {
	var as TError
	ok := errors.As(err, &as)
	return as, ok
}

// IsErrorType reports whether or not the type of any error in err's chain matches
// the Error type.
func IsErrorType[TError error](err error) bool {
	_, ok := AsError[TError](err)
	return ok
}
