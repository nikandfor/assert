package deep

import (
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/nikandfor/errors"
)

type (
	prefixWriter struct {
		io.Writer
		pref []byte
		add  bool
	}

	visit struct {
		a, b unsafe.Pointer
		typ  reflect.Type
	}

	// rtype is the common implementation of most values.
	// It is embedded in other struct types.
	//
	// rtype must be kept in sync with ../runtime/type.go:/^type._type.
	rtype struct {
		size       uintptr
		ptrdata    uintptr // number of bytes in the type that can contain pointers
		hash       uint32  // hash of type; avoids computation in hash tables
		tflag      tflag   // extra type information flags
		align      uint8   // alignment of variable with this type
		fieldAlign uint8   // alignment of struct field with this type
		kind       uint8   // enumeration for C
		// function for comparing objects of this type
		// (ptr to object A, ptr to object B) -> ==?
		equal     func(unsafe.Pointer, unsafe.Pointer) bool
		gcdata    *byte   // garbage collection data
		str       nameOff // string form
		ptrToThis typeOff // type for pointer to this type, may be zero
	}

	tflag uint8

	nameOff int32
	typeOff int32

	value struct {
		typ  *rtype
		ptr  unsafe.Pointer
		flag uintptr
	}
)

var spaces = "                                                                          "

var stop = map[reflect.Type]struct{}{
	reflect.TypeOf(time.Time{}): struct{}{},
	reflect.TypeOf(&big.Int{}):  struct{}{},
}

func Equal(a, b interface{}) bool {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	return equal(av, bv, nil)
}

func Diff(w io.Writer, a, b interface{}) bool {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	return equal(av, bv, nil)
}

func equal(a, b reflect.Value, visited map[visit]struct{}) bool {
	if !a.IsValid() || !b.IsValid() {
		return a.IsValid() == b.IsValid()
	}
	if a.Type() != b.Type() {
		return false
	}

	// The hard part is taken from reflect.DeepEqual

	// We want to avoid putting more in the visited map than we need to.
	// For any possible reference cycle that might be encountered,
	// hard(v1, v2) needs to return true for at least one of the types in the cycle,
	// and it's safe and valid to get Value's internal pointer.
	hard := func(v1, v2 reflect.Value) bool {
		switch v1.Kind() {
		case reflect.Ptr:
			if ptrdata(v1) == 0 {
				// go:notinheap pointers can't be cyclic.
				// At least, all of our current uses of go:notinheap have
				// that property. The runtime ones aren't cyclic (and we don't use
				// DeepEqual on them anyway), and the cgo-generated ones are
				// all empty structs.
				return false
			}

			fallthrough
		case reflect.Map, reflect.Slice, reflect.Interface:
			// Nil pointers cannot be cyclic. Avoid putting them in the visited map.
			return !v1.IsNil() && !v2.IsNil()
		}

		return false
	}

	if hard(a, b) {
		// For a Ptr or Map value, we need to check flagIndir,
		// which we do by calling the pointer method.
		// For Slice or Interface, flagIndir is always set,
		// and using v.ptr suffices.
		ptrval := func(v reflect.Value) unsafe.Pointer {
			switch v.Kind() {
			case reflect.Ptr, reflect.Map:
				return valuePointer(v)
			default:
				return (*value)(unsafe.Pointer(&v)).ptr
			}
		}

		addr1 := ptrval(a)
		addr2 := ptrval(b)
		if uintptr(addr1) > uintptr(addr2) {
			// Canonicalize order to reduce number of entries in visited.
			// Assumes non-moving garbage collector.
			addr1, addr2 = addr2, addr1
		}

		// Short circuit if references are already seen.
		typ := a.Type()
		v := visit{a: addr1, b: addr2, typ: typ}
		if _, ok := visited[v]; ok {
			return true
		}

		if visited == nil {
			visited = make(map[visit]struct{})
		}

		// Remember for later.
		visited[v] = struct{}{}
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
		reflect.Float64, reflect.Float32,
		reflect.Complex128, reflect.Complex64,
		reflect.String,
		reflect.Chan,
		reflect.Bool:

		return eface(a) == eface(b)

	case reflect.Interface:
		ai := a.InterfaceData()
		bi := b.InterfaceData()

		if ai[0] != bi[0] {
			return false
		}

		return equal(a.Elem(), b.Elem(), visited)
	case reflect.Slice, reflect.Array:
		return equalSlice(a, b, visited)

	case reflect.Struct:
		return equalStructFields(a, b, visited)

	default:
		panic(fmt.Sprintf("%v", a.Kind()))
	}
}

func equalStructFields(a, b reflect.Value, visited map[visit]struct{}) bool {
	t := a.Type()

	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		if ft.Tag.Get("deep") == "-" {
			continue
		}
		if v, ok := getTag(ft, "deep", "compare"); ok && v == "false" {
			continue
		}

		if !equal(a.Field(i), b.Field(i), visited) {
			return false
		}
	}

	return true
}

func equalSlice(a, b reflect.Value, visited map[visit]struct{}) bool {
	if a.Len() != b.Len() {
		return false
	}

	for i := 0; i < a.Len(); i++ {
		if !equal(a.Index(i), b.Index(i), visited) {
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

	tp := x.Type()

	if _, ok := stop[tp]; ok {
		return writef(w, n, "%#v", x.Interface())
	}

	for x.Kind() == reflect.Ptr {
		if x.IsNil() {
			return writef(w, n, "(%v)(nil)", x.Type())
		}

		n, err = writef(w, n, "&")
		if err != nil {
			return
		}

		x = x.Elem()
	}

	if _, ok := stop[tp]; ok {
		return writef(w, n, "%#v", x.Interface())
	}

	switch x.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr, reflect.UnsafePointer:

		n, err = writef(w, n, "%v(0x%x)", x.Type(), x)
	case reflect.String:
		n, err = writef(w, n, "%q", x.String())
	case reflect.Slice, reflect.Array:
		if x.Kind() == reflect.Slice && x.IsNil() {
			return writef(w, n, `%v(nil)`, tp)
		}

		if tp := x.Type(); tp.Elem().Kind() == reflect.Uint8 {
			return writef(w, n, `%v(unhex("%x"))`, tp, x.Slice(0, x.Len()).Bytes())
		}

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

		n, err = writef(w, n, " (kind: %v)", x.Kind())
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

		if l := len(ft.Name); l < 14 {
			n, err = writef(w, n, "%v", spaces[:14-l])
			if err != nil {
				return
			}
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

func ptrdata(v reflect.Value) uintptr {
	return (*value)(unsafe.Pointer(&v)).typ.ptrdata
}

//go:linkname valuePointer reflect.Value.pointer
func valuePointer(v reflect.Value) unsafe.Pointer
