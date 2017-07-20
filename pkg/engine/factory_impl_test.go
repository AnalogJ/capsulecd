package engine

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestEngineChef(t *testing.T) {
	eng := new(engineChef)
	assert.Implements(t, (*Interface)(nil), eng, "should implement the Engine interface")
}

func TestEngineGolang(t *testing.T) {
	eng := new(engineGolang)
	assert.Implements(t, (*Interface)(nil), eng, "should implement the Engine interface")
}

func TestEngineNode(t *testing.T) {
	eng := new(engineNode)
	assert.Implements(t, (*Interface)(nil), eng, "should implement the Engine interface")
}

func TestEnginePython(t *testing.T) {
	eng := new(enginePython)
	assert.Implements(t, (*Interface)(nil), eng, "should implement the Engine interface")
}

func TestEngineRuby(t *testing.T) {
	eng := new(engineRuby)
	assert.Implements(t, (*Interface)(nil), eng, "should implement the Engine interface")
}