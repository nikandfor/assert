package deep

import (
	"bytes"
	"testing"
)

type (
	A struct {
		A int
		B string
		C uint64
		D []int
	}

	B struct {
		A A
		B int `deep:"print:omit"`
		C []byte
		D []int
		E interface{}
	}
)

func TestFprint(t *testing.T) {
	b := B{
		A: A{
			A: 1,
			B: "second",
			C: 5,
		},
		B: 9,
		C: []byte("123"),
		D: []int{1, 2, 3},
	}

	var buf bytes.Buffer

	n, err := Fprint(&buf, &b)
	if err != nil {
		t.Errorf("fprint error: %v", err)
	}
	if n != buf.Len() {
		t.Errorf("bad size: %d != %d", n, buf.Len())
	}

	if b := buf.Bytes(); len(b) == 0 {
		t.Errorf("empty result")
	}

	t.Logf("result:\n%s", buf.Bytes())
}

func TestEqual(t *testing.T) {
	x := B{
		A: A{
			A: 1,
			B: "second",
			C: 5,
		},
		B: 9,
		C: []byte("123"),
		D: []int{1, 2, 3},
		E: 5,
	}

	y := B{
		A: A{
			A: 1,
			B: "second",
			C: 5,
		},
		B: 9,
		C: []byte("123"),
		D: []int{1, 2, 3},
		E: 5,
	}

	eq := Equal(x, y)
	if !eq {
		t.Errorf("excepted to be equal")
	}

	//

	y.D[2]++

	eq = Equal(x, y)
	if eq {
		t.Errorf("excepted to not to be equal")
	}

	y.D[2]--

	//

	y.E = 6

	eq = Equal(x, y)
	if eq {
		t.Errorf("excepted to not to be equal")
	}

	//

	y.E = int64(5)

	eq = Equal(x, y)
	if eq {
		t.Errorf("excepted to not to be equal")
	}
}
