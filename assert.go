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

func All(t TestingT, c []Checker, args ...interface{}) (ok bool) {
	if h, ok := t.(helper); ok {
		h.Helper()
	}

	bs := make([]wbuf, len(c))

	ok = true

	for i, c := range c {
		if !c.Check(&bs[i]) {
			ok = false
			break
		}
	}

	if !ok {
		Fail(t, append([]interface{}{bs}, args...)...)
	}

	return ok
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
			b.Newline()
		case []wbuf:
			for _, a := range a {
				b = append(b, a...)
				b.Newline()
			}
		case string:
			fmt.Fprintf(&b, a, args[i+1:]...)

			b.Newline()

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

func True(t TestingT, ok bool, args ...interface{}) bool {
	if h, ok := t.(helper); ok {
		h.Helper()
	}

	return Eval(t, is.True(ok), args...)
}

func False(t TestingT, ok bool, args ...interface{}) bool {
	if h, ok := t.(helper); ok {
		h.Helper()
	}

	return Eval(t, is.False(ok), args...)
}

func Nil(t TestingT, x interface{}, args ...interface{}) bool {
	if h, ok := t.(helper); ok {
		h.Helper()
	}

	return Eval(t, is.Nil(x), args...)
}

func NotNil(t TestingT, x interface{}, args ...interface{}) bool {
	if h, ok := t.(helper); ok {
		h.Helper()
	}

	return Eval(t, is.NotNil(x), args...)
}

func NoError(t TestingT, err error, args ...interface{}) bool {
	if h, ok := t.(helper); ok {
		h.Helper()
	}

	return Eval(t, is.NoError(err), args...)
}

func Error(t TestingT, err error, args ...interface{}) bool {
	if h, ok := t.(helper); ok {
		h.Helper()
	}

	return Eval(t, is.Error(err), args...)
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

func (w *wbuf) Newline() {
	if l := len(*w); l == 0 || (*w)[l-1] == '\n' {
		return
	}

	*w = append(*w, '\n')
}
