package utils_test

import (
	"capsulecd/pkg/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

var snakeCaseTests = []struct {
	n        string // input
	expected string // expected result
}{
	{"this_is_an_input", "ThisIsAnInput"},
	{"", ""},
	{"hello", "Hello"},
}

func TestSnakeCaseToCamelCase(t *testing.T) {
	t.Parallel()
	for _, tt := range snakeCaseTests {
		//test
		actual := utils.SnakeCaseToCamelCase(tt.n)

		//assert
		require.Equal(t, tt.expected, actual, "should convert to camel case correctly")
	}
}


func TestStripIndent(t *testing.T) {
	t.Parallel()

	testString := `
	this is my multi line string
	line2
	line 3`

	require.Equal(t,"\nthis is my multi line string\nline2\nline 3", utils.StripIndent(testString))
}