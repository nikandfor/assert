package is

import (
	"errors"
	"fmt"
	"io"
	"reflect"
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

func True(ok bool) Checker {
	return CheckerFunc(func(w io.Writer) bool {
		if ok {
			return true
		}

		fmt.Fprintf(w, "Want true")

		return false
	})
}

func False(ok bool) Checker {
	return CheckerFunc(func(w io.Writer) bool {
		if !ok {
			return true
		}

		fmt.Fprintf(w, "Want false")

		return false
	})
}

func Nil(x interface{}) Checker {
	return CheckerFunc(func(w io.Writer) bool {
		if x == nil {
			return true
		}

		r := reflect.ValueOf(x)

		if k := r.Kind(); k == reflect.Ptr || k == reflect.Interface || k == reflect.Slice || k == reflect.Map || k == reflect.Chan || k == reflect.Func {
			return r.IsNil()
		}

		fmt.Fprintf(w, "Want nil, got: %v", x)

		return false
	})
}

func NotNil(x interface{}) Checker {
	return CheckerFunc(func(w io.Writer) bool {
		if x != nil {
			return true
		}

		fmt.Fprintf(w, "Want not nil")

		return false
	})
}

func NoError(err error) Checker {
	return CheckerFunc(func(w io.Writer) bool {
		if err == nil {
			return true
		}

		fmt.Fprintf(w, "Error: %+v (type: %[1]T)", err)

		return false
	})
}

func Error(err error) Checker {
	return CheckerFunc(func(w io.Writer) bool {
		if err != nil {
			return true
		}

		fmt.Fprintf(w, "Want error")

		return false
	})
}

func ErrorIs(err, target error) Checker {
	return CheckerFunc(func(w io.Writer) bool {
		if errors.Is(err, target) {
			return true
		}

		if err == nil {
			fmt.Fprintf(w, "Want error: %q (type %T)", target.Error(), target)

			return false
		}

		fmt.Fprintf(w, "Error chain\n")

		for e := err; e != nil; e = errors.Unwrap(e) {
			fmt.Fprintf(w, "%q (type %T)\n", e.Error(), e)
		}

		fmt.Fprintf(w, "is not %q (type %T)", target.Error(), target)

		return false
	})
}

func Zero(x interface{}) Checker {
	return CheckerFunc(func(w io.Writer) bool {
		if x == nil {
			return true
		}

		r := reflect.ValueOf(x)

		if r.IsZero() {
			return true
		}

		fmt.Fprintf(w, "Want zero value, got: %v", x)

		return false
	})
}

func NotZero(x interface{}) Checker {
	return CheckerFunc(func(w io.Writer) bool {
		r := reflect.ValueOf(x)

		if !r.IsZero() {
			return true
		}

		fmt.Fprintf(w, "Want not zero value, got: %v", x)

		return false
	})
}

func (f CheckerFunc) Check(w io.Writer) bool { return f(w) }
