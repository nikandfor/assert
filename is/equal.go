package is

import (
	"bytes"
	"fmt"
	"io"
	"runtime/debug"

	"github.com/nikandfor/assert/deep"
)

func Equal(a, b interface{}) Checker {
	return CheckerFunc(func(w io.Writer) bool {
		var buf bytes.Buffer

		defer func() {
			p := recover()
			if p == nil {
				return
			}

			fmt.Fprintf(w, "PANIC: %v\n", p)
			fmt.Fprintf(w, "%s", debug.Stack())
		}()

		eq := deep.Diff(&buf, a, b)
		if eq {
			return true
		}

		//	fmt.Fprintf(w, "Not equal:\nExpected: %#v\nActual:   %#v\nDiff:\n%s", a, b, buf.Bytes())
		fmt.Fprintf(w, "Not equal:\nExpected: ")

		_, err := deep.Fprint(w, a)
		if err != nil {
			fmt.Fprintf(w, "PRINT ERROR: %v\n", err)
		}

		fmt.Fprintf(w, "\nActual:   ")

		_, err = deep.Fprint(w, b)
		if err != nil {
			fmt.Fprintf(w, "PRINT ERROR: %v\n", err)
		}

		fmt.Fprintf(w, "\nDiff:\n%s", buf.Bytes())

		return false
	})
}

func NotEqual(a, b interface{}) Checker {
	return CheckerFunc(func(w io.Writer) bool {
		var buf bytes.Buffer

		defer func() {
			p := recover()
			if p == nil {
				return
			}

			fmt.Fprintf(w, "PANIC: %v\n", p)
			fmt.Fprintf(w, "%s", debug.Stack())
		}()

		eq := deep.Diff(&buf, a, b)
		if !eq {
			return true
		}

		//	fmt.Fprintf(w, "Not equal:\nExpected: %#v\nActual:   %#v\nDiff:\n%s", a, b, buf.Bytes())
		fmt.Fprintf(w, "Expected not equal: ")

		_, err := deep.Fprint(w, a)
		if err != nil {
			fmt.Fprintf(w, "PRINT ERROR: %v\n", err)
		}

		return false
	})
}
