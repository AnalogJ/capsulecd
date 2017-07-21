package scm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScmGithub(t *testing.T) {
	scm := new(scmGithub)
	assert.Implements(t, (*Interface)(nil), scm, "should implement the Scm interface")
}

func TestScmBitbucket(t *testing.T) {
	eng := new(scmBitbucket)
	assert.Implements(t, (*Interface)(nil), eng, "should implement the Scm interface")
}
