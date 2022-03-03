package is

import (
	"bytes"
	"fmt"
	"io"

	"github.com/nikandfor/assert/deep"
)

type ()

func Equal(a, b interface{}) Checker {
	return CheckerFunc(func(w io.Writer) bool {
		if a == b {
			return true
		}

		var buf bytes.Buffer

		eq := deep.Diff(&buf, a, b)
		if eq {
			return true
		}

		fmt.Fprintf(w, "Not equal:\nExpected: %#v\nActual:   %#v\nDiff:\n%s", a, b, buf.Bytes())

		return false
	})
}
