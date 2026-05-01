package core

import "errors"

var (
	ErrInternal         = errors.New("internal server error")
	ErrDeadlineExceeded = errors.New("deadline is exceeded")
	ErrCanceled         = errors.New("context is canceled")
)
