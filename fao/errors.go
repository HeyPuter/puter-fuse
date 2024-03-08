package fao

import (
	"fmt"
	"syscall"
)

type ErrDoesNotExist struct {
	Path string
}

func (e *ErrDoesNotExist) Error() string {
	return "Does not exist: " + e.Path
}

type ErrNotDirectory struct {
	Path string
}

func (e *ErrNotDirectory) Error() string {
	return "Not a directory: " + e.Path
}

type FAOError struct {
	Errno syscall.Errno
	From  error
}

func (e *FAOError) Error() string {
	return e.From.Error()
}

func Errorf(errno syscall.Errno, format string, a ...interface{}) error {
	return &FAOError{
		Errno: errno,
		From:  fmt.Errorf(format, a...),
	}
}
