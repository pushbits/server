// Package pberrors defines errors specific to PushBits
package pberrors

import "errors"

// ErrMessageNotFound indicates that a message does not exist
var ErrMessageNotFound = errors.New("message not found")

// ErrConfigTLSFilesInconsistent indicates that either just a certfile or a keyfile was provided
var ErrConfigTLSFilesInconsistent = errors.New("TLS certfile and keyfile must either both be provided or omitted")
