package internal

import (
	"errors"
	"runtime/debug"
)

var ErrNotFound = errors.New("not found")

type ErrWithStackTrace struct {
	error
	Stack []byte
}

func (e ErrWithStackTrace) Unwrap() error {
	return e.error
}

func NewErrWithStackTrace(err error) error {
	if IsProd {
		return err
	}

	var errst ErrWithStackTrace
	if errors.As(err, &errst) {
		// already stack strace
		return err
	}

	return ErrWithStackTrace{
		error: err,
		Stack: debug.Stack(),
	}
}

// SErr stands for Sensitive Error
type SErr struct {
	Public, Private error
}

func NewSErr(msg string, errs ...error) SErr {
	return SErr{
		Public:  errors.New(msg),
		Private: errors.Join(append(errs, errors.New(msg))...),
	}
}

func (e SErr) Error() string {
	return e.Public.Error()
}

func (e SErr) Unwrap() error {
	return e.Public
}
