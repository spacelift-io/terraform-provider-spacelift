package internal

import (
	"fmt"
	"strings"
)

// FromSpaceliftError wraps the error with a helpful message when encountering a Spacelift error.
// In this case an unauthorized error.
func FromSpaceliftError(err error) error {
	if err == nil || !strings.Contains(err.Error(), "unauthorized") {
		return err
	}
	return fmt.Errorf("%w - is it an administrative stack?", err)
}
