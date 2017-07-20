package scm

import (
"testing"
"github.com/stretchr/testify/assert"
)

func TestScmGithub(t *testing.T) {
	scm := new(scmGithub)
	assert.Implements(t, (*Interface)(nil), scm, "should implement the Scm interface")
}

func TestScmBitbucket(t *testing.T) {
	eng := new(scmBitbucket)
	assert.Implements(t, (*Interface)(nil), eng, "should implement the Scm interface")
}
