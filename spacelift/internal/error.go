package internal

import (
	"fmt"
	"strings"
)

// WrapUnauthorized wraps the error with a helpful message when encountering an unauthorized error
func WrapUnauthorized(err error) error {
	if err == nil || !strings.Contains(err.Error(), "unauthorized") {
		return err
	}
	return fmt.Errorf("%w - is it an administrative stack?", err)
}
