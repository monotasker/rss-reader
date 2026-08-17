package store

import "errors"

// ErrNotFound is returned when a requested row does not exist.
var ErrNotFound = errors.New("not found")

// ErrDuplicate is returned when a uniqueness constraint is violated
var ErrDuplicate = errors.New("already exists")
