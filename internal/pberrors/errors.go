// Package pberrors defines errors specific to PushBits
package pberrors

import "errors"

// ErrorMessageNotFound indicates that a message does not exist
var ErrorMessageNotFound = errors.New("message not found")
