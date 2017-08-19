package engine

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEngineChef(t *testing.T) {
	eng := new(engineChef)
	require.Implements(t, (*Interface)(nil), eng, "should implement the Engine interface")
}

func TestEngineGeneric(t *testing.T) {
	eng := new(engineGeneric)
	require.Implements(t, (*Interface)(nil), eng, "should implement the Engine interface")
}

func TestEngineGolang(t *testing.T) {
	eng := new(engineGolang)
	require.Implements(t, (*Interface)(nil), eng, "should implement the Engine interface")
}

func TestEngineNode(t *testing.T) {
	eng := new(engineNode)
	require.Implements(t, (*Interface)(nil), eng, "should implement the Engine interface")
}

func TestEnginePython(t *testing.T) {
	eng := new(enginePython)
	require.Implements(t, (*Interface)(nil), eng, "should implement the Engine interface")
}

func TestEngineRuby(t *testing.T) {
	eng := new(engineRuby)
	require.Implements(t, (*Interface)(nil), eng, "should implement the Engine interface")
}
