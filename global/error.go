package global

import (
	"github.com/pkg/errors"
)

var (
	ErrNotInTransaction = errors.New("run error:not in transaction")
)

var (
	ErrNotBelongCurrentUser = errors.New("not belong current user")
	ErrInvalidRequest       = errors.New("invalid request")
	ErrInvalidParameter     = errors.New("invalid parameter")
)
