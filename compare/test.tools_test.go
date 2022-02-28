package compare

import (
	"testing"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

func TestTestToolsEqual(t *testing.T) {
	assert.Check(t, cmp.Equal(1, 2), "msg and %v", "args")

	assert.Check(t, cmp.Equal("str", "str1"), "msg and %v", "args")

	assert.Check(t, cmp.DeepEqual([]byte("aaabbbccc"), []byte("aaavvvccc")), "msg and %v", "args")
}
