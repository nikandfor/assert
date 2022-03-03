package is

import (
	"errors"
	"fmt"
	"io"
)

type (
	Checker interface {
		Check(io.Writer) bool
	}

	CheckerFunc func(w io.Writer) bool

	equal struct {
		a, b interface{}
	}
)

func NoError(err error) Checker {
	return CheckerFunc(func(w io.Writer) bool {
		if err == nil {
			return true
		}

		fmt.Fprintf(w, "Error: %+v (type: %[1]T)", err)

		return false
	})
}

func ErrorIs(err, target error) Checker {
	return CheckerFunc(func(w io.Writer) bool {
		if errors.Is(err, target) {
			return true
		}

		fmt.Fprintf(w, "Error chain\n")

		for e := err; e != nil; e = errors.Unwrap(e) {
			fmt.Fprintf(w, "%q (type %T)\n", e.Error(), e)
		}

		fmt.Fprintf(w, "is not %q (type %T)", target.Error(), target)

		return false
	})
}

func (f CheckerFunc) Check(w io.Writer) bool { return f(w) }
