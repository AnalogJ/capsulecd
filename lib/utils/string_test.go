package utils_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"capsulecd/lib/utils"
)


var snakeCaseTests = []struct {
	n        string // input
	expected string // expected result
}{
	{"this_is_an_input", "ThisIsAnInput"},
	{"",""},
	{"hello", "Hello"},
}

func TestSnakeCaseToCamelCase(t *testing.T) {
	t.Parallel()
	for _, tt := range snakeCaseTests {
		actual := utils.SnakeCaseToCamelCase(tt.n)

		assert.Equal(t, tt.expected, actual, "should convert to camel case correctly")
	}
}