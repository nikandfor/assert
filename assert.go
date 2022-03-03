package assert

import (
	"fmt"

	"github.com/nikandfor/assert/is"
)

type (
	TestingT interface{}

	Checker = is.Checker

	helper interface {
		Helper()
	}

	fail interface {
		Fail()
	}

	wbuf []byte
)

func Eval(t TestingT, c Checker, args ...interface{}) (ok bool) {
	if h, ok := t.(helper); ok {
		h.Helper()
	}

	//	fargs := source.AssertionArgs(t)

	var b wbuf

	if c.Check(&b) {
		return true
	}

	Fail(t, append([]interface{}{b}, args...)...)

	return false
}

func Any(t TestingT, c []Checker, args ...interface{}) bool {
	if h, ok := t.(helper); ok {
		h.Helper()
	}

	bs := make([]wbuf, len(c))

	for i, c := range c {
		if c.Check(&bs[i]) {
			return true
		}
	}

	Fail(t, append([]interface{}{bs}, args...)...)

	return false
}

func Fail(t TestingT, args ...interface{}) {
	if h, ok := t.(helper); ok {
		h.Helper()
	}

	var b wbuf

loop:
	for i, a := range args {
		switch a := a.(type) {
		case wbuf:
			b = append(b, a...)
		case []wbuf:
			for _, a := range a {
				b = append(b, a...)
			}
		case string:
			fmt.Fprintf(&b, a, args[i+1:]...)
			b = append(b, '\n')

			break loop
		}
	}

	switch t := t.(type) {
	case interface{ Logf(string, ...interface{}) }:
		t.Logf("%s", b)
	case interface{ Log(...interface{}) }:
		t.Log(string(b))
	default:
		panic(fmt.Sprintf("unsupported testing.T: %T", t))
	}

	if t, ok := t.(fail); ok {
		t.Fail()
	}
}

func NoError(t TestingT, err error, args ...interface{}) bool {
	if h, ok := t.(helper); ok {
		h.Helper()
	}

	return Eval(t, is.NoError(err), args...)
}

func ErrorIs(t TestingT, err, target error, args ...interface{}) bool {
	if h, ok := t.(helper); ok {
		h.Helper()
	}

	return Eval(t, is.ErrorIs(err, target), args...)
}

func Equal(t TestingT, exp, act interface{}, args ...interface{}) bool {
	if h, ok := t.(helper); ok {
		h.Helper()
	}

	return Eval(t, is.Equal(exp, act), args...)
}

func (w *wbuf) Write(p []byte) (int, error) {
	*w = append(*w, p...)

	return len(p), nil
}
