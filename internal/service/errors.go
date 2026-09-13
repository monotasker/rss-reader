package service

import (
	"errors"
	"fmt"
)

var ErrInvalidURL = errors.New("invalid URL")

type urlValidationError = struct {
	Input string
	Err   error
}

func (u *urlValidationError) Error() string {
	if u.Err != nil {
		return fmt.Errorf("invalid URL %q: %v", u.Input, u.Err)
	}
	return fmt.Sprintf("invalid URL %q", u.Input)
}

func (u *urlValidationError) Is(target error) bool {
	return target == ErrInvalidURL
}
