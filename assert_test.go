package assert_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/nikandfor/assert"
)

type (
	wbuf []byte

	TestT struct {
		testing.T

		failed int
		b      wbuf
	}
)

func TestNoError(t *testing.T) {
	tt := &TestT{}

	assert.NoError(tt, nil)
	checkOK(t, tt)

	tt.reset()

	assert.NoError(tt, errors.New("test_error"))
	checkFailed(t, tt, 1)
}

func TestErrorIs(t *testing.T) {
	tt := &TestT{}

	assert.ErrorIs(tt, io.EOF, io.EOF)
	checkOK(t, tt)

	tt.reset()

	assert.ErrorIs(tt, fmt.Errorf("wrapped: %w", io.EOF), io.EOF)
	checkOK(t, tt)

	tt.reset()

	assert.NoError(tt, errors.New("test_error"), io.EOF)
	checkFailed(t, tt, 1)
}

func TestEqual(t *testing.T) {
	tt := &TestT{}

	assert.Equal(tt, "asd", "asd")
	checkOK(t, tt)

	assert.Equal(tt, "asd", "qwe")
	checkFailed(t, tt, 1)
}

func checkOK(t *testing.T, tt *TestT) {
	if tt.failed == 0 && len(tt.b) == 0 {
		return
	}

	failTT(t, tt)
}

func checkFailed(t *testing.T, tt *TestT, failed int) {
	if tt.failed == failed && len(tt.b) != 0 {
		t.Logf("TEST OUTPUT:\n%s", tt.b)

		return
	}

	failTT(t, tt)
}

func failTT(t *testing.T, tt *TestT) {
	t.Errorf("failed: %v\n%s", tt.failed, tt.b)
}

func (tt *TestT) Fail() {
	tt.failed = 1
}

func (tt *TestT) FailNow() {
	tt.failed = 2
}

func (tt *TestT) Logf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(&tt.b, format, args...)
}

func (tt *TestT) reset() {
	tt.failed = 0
	tt.b = tt.b[:0]
}

func (w *wbuf) Write(p []byte) (int, error) {
	*w = append(*w, p...)

	return len(p), nil
}
