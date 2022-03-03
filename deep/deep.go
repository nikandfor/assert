package deep

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"unsafe"

	"github.com/nikandfor/errors"
)

type (
	prefixWriter struct {
		io.Writer
		pref []byte
		add  bool
	}
)

var spaces = "                                                                          "

func Equal(a, b interface{}) bool {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	return equal(av, bv)
}

func Diff(w io.Writer, a, b interface{}) bool {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	return equal(av, bv)
}

func equal(a, b reflect.Value) bool {
	if !a.IsValid() || !b.IsValid() {
		return a.IsValid() == b.IsValid()
	}
	if a.Type() != b.Type() {
		return false
	}

	for a.Kind() == reflect.Ptr {
		if a.IsNil() != b.IsNil() {
			return false
		}

		if a.IsNil() {
			return true
		}

		a = a.Elem()
		b = b.Elem()
	}

	switch a.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr, reflect.UnsafePointer,
		reflect.String:

		return eface(a) == eface(b)

	case reflect.Interface:
		ai := a.InterfaceData()
		bi := b.InterfaceData()

		if ai[0] != bi[0] {
			return false
		}

		return equal(a.Elem(), b.Elem())
	case reflect.Slice, reflect.Array:
		return equalSlice(a, b)

	case reflect.Struct:
		return equalStructFields(a, b)

	default:
		panic(fmt.Sprintf("%v", a.Kind()))
	}
}

func equalStructFields(a, b reflect.Value) bool {
	t := a.Type()

	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		if ft.Tag.Get("deep") == "-" {
			continue
		}
		if v, ok := getTag(ft, "deep", "compare"); ok && v == "false" {
			continue
		}

		if !equal(a.Field(i), b.Field(i)) {
			return false
		}
	}

	return true
}

func equalSlice(a, b reflect.Value) bool {
	if a.Len() != b.Len() {
		return false
	}

	for i := 0; i < a.Len(); i++ {
		if !equal(a.Index(i), b.Index(i)) {
			return false
		}
	}

	return true
}

func Fprint(w io.Writer, x ...interface{}) (n int, err error) {
	for i, x := range x {
		n, err = fprint(w, n, reflect.ValueOf(x), 0)
		if err != nil {
			return n, errors.Wrap(err, "%d", i)
		}
	}

	return
}

func fprint(w io.Writer, n int, x reflect.Value, d int) (m int, err error) {
	//	defer func() {
	//		fmt.Fprintf(os.Stderr, "fprint: n:%v  x:%v  from %v\n", m, x, loc.Caller(1))
	//	}()

	for x.Kind() == reflect.Ptr {
		n, err = writef(w, n, "&")
		if err != nil {
			return
		}

		x = x.Elem()
	}

	switch x.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr, reflect.UnsafePointer:

		n, err = writef(w, n, "%v(%v)", x.Type(), x)
	case reflect.String:
		n, err = writef(w, n, "%q", x.String())
	case reflect.Slice:
		n, err = writef(w, n, "%v", x.Type())
		if err != nil {
			return
		}

		n, err = fprintSlice(w, n, x, d+1)
		if err != nil {
			return
		}
	case reflect.Struct:
		n, err = writef(w, n, "%v{\n", x.Type())
		if err != nil {
			return
		}

		n, err = fprintStructFields(w, n, x, d+1)
		if err != nil {
			return
		}

		n, err = ident(w, n, d, "}")
	default:
		n, err = writef(w, n, "%v", x.Type())
		if err != nil {
			return
		}

		n, err = writef(w, n, " (kind: %v)\n", x.Kind())
	}

	if err != nil {
		return
	}

	return n, nil
}

func fprintStructFields(w io.Writer, n int, x reflect.Value, d int) (_ int, err error) {
	t := x.Type()

	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		if ft.Tag.Get("deep") == "-" {
			continue
		}
		if v, ok := getTag(ft, "deep", "print"); ok && v == "omit" {
			continue
		}

		n, err = ident(w, n, d, "")
		if err != nil {
			return
		}

		n, err = writef(w, n, "%v: ", ft.Name)
		if err != nil {
			return
		}

		n, err = fprint(w, n, x.Field(i), d)
		if err != nil {
			return
		}

		n, err = writef(w, n, "\n")
		if err != nil {
			return
		}
	}

	return n, nil
}

func fprintSlice(w io.Writer, n int, x reflect.Value, d int) (m int, err error) {
	t := x.Type().Elem()
	k := t.Kind()

	if x.IsNil() {
		return writef(w, n, "(nil)")
	}

	if k == reflect.Uint8 {
		ok := 0
		for _, c := range x.Bytes() {
			if c >= 0x20 && c < 0x80 {
				ok++
			}
		}

		if ok*5/4 >= x.Len() {
			return writef(w, n, "(%q)", x.Bytes())
		}
	}

	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr, reflect.UnsafePointer:

		n, err = writef(w, n, "{")
		if err != nil {
			return
		}

		for i := 0; i < x.Len(); i++ {
			xx := x.Index(i)

			if i != 0 {
				n, err = writef(w, n, ", ")
				if err != nil {
					return
				}
			}

			if k == reflect.UnsafePointer {
				n, err = writef(w, n, "0x%x", xx.Pointer())
				if err != nil {
					return
				}

				continue
			}

			var val interface{}

			switch k {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				val = xx.Int()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Uintptr:
				val = xx.Uint()
			}

			n, err = writef(w, n, "%v", val)
			if err != nil {
				return
			}
		}

		n, err = writef(w, n, "}")
	default:
		n, err = writef(w, n, "{elems}")
	}

	return n, nil
}

func ident(w io.Writer, n, d int, fmt string, args ...interface{}) (_ int, err error) {
	n, err = writef(w, n, "%s", spaces[:4*d])
	if err != nil {
		return
	}

	if fmt == "" && len(args) == 0 {
		return n, err
	}

	return writef(w, n, fmt, args...)
}

func writef(w io.Writer, i int, format string, args ...interface{}) (n int, err error) {
	n, err = fmt.Fprintf(w, format, args...)
	return i + n, err
}

func getTag(x reflect.StructField, t, k string) (string, bool) {
	tags := strings.Split(x.Tag.Get(t), ",")

	for _, tag := range tags {
		kv := strings.SplitN(tag, ":", 2)
		if kv[0] == k {
			if len(kv) == 1 {
				return "", true
			}

			return kv[1], true
		}
	}

	return "", false
}

func (w *prefixWriter) Write(p []byte) (n int, err error) {
	i := 0

	for i < len(p) {
		if w.add {
			_, err = w.Writer.Write(w.pref)
			if err != nil {
				return
			}
		}

		st := i

		for i < len(p) && p[i] != '\n' {
			i++
		}

		if i < len(p) && p[i] == '\n' {
			i++

			w.add = true
		}

		var m int
		m, err = w.Writer.Write(p[st:i])
		n += m
		if err != nil {
			return
		}
	}

	return
}

func eface(x reflect.Value) interface{} {
	return *(*interface{})(unsafe.Pointer(&x))
}
