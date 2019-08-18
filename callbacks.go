package guillotine

import (
	"github.com/cockroachdb/errors"
	"go.stevenxie.me/gopkg/zero"
)

// WithPrefix creates a Callback adds a prefix to the error from a
// Finalizer.
func WithPrefix(msg string) Callback {
	return func(err error) error { return errors.WithMessage(err, msg) }
}

// WithError creates a Callback that creates an Error if the Finalizer didn't
// return one.
func WithError(msg string) Callback {
	return func(err error) error {
		if err != nil {
			return err
		}
		return errors.New(msg)
	}
}

// WithErrorf is like WithError, but creates a formatted error message.
func WithErrorf(format string, args ...zero.Interface) Callback {
	return func(err error) error {
		if err != nil {
			return err
		}
		return errors.Newf(format, args...)
	}
}

// WithFunc creates a Callback that runs an arbitrary function.
func WithFunc(f func()) Callback {
	return func(error) error {
		f()
		return nil
	}
}

// WithEffect creates a Callback that runs a function that receives the
// error created by the Finalizer, and returns nothing.
//
// It is intended for functions with side effects involving the error, like
// logging functions, etc.
func WithEffect(f func(error)) Callback {
	return func(err error) error {
		f(err)
		return err
	}
}
