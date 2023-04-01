// Package assert provides convenience function to make assertions at runtime.
package assert

import (
	"errors"
)

// Assert panics if condition is false.
func Assert(condition bool) {
	if !condition {
		panic(errors.New("assertion failed"))
	}
}
