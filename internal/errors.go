package internal

import (
	"errors"
	"runtime/debug"
)

var (
	ErrNotFound = errors.New("not found")
	ErrInvalid  = errors.New("invalid")
)

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

// PubErr stands for Sensitive Error
type PubErr struct {
	Public, Private error
}

func NewPubErr(public, private error) PubErr {
	return PubErr{
		Public:  public,
		Private: private,
	}
}

func (e PubErr) Error() string {
	return e.Public.Error()
}

func (e PubErr) Unwrap() error {
	return e.Public
}
