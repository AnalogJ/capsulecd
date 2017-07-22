package scm

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestScmGithub(t *testing.T) {
	scm := new(scmGithub)
	require.Implements(t, (*Interface)(nil), scm, "should implement the Scm interface")
}

func TestScmBitbucket(t *testing.T) {
	eng := new(scmBitbucket)
	require.Implements(t, (*Interface)(nil), eng, "should implement the Scm interface")
}
