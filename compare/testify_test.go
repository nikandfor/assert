package compare

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestifyEqual(t *testing.T) {
	assert.Equal(t, 1, 2, "msg and %v", "args")

	assert.Equal(t, "str", "str1", "msg and %v", "args")

	assert.Equal(t, []byte("aaabbbccc"), []byte("aaavvvccc"), "msg and %v", "args")
}
