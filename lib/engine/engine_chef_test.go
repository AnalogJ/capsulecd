package engine_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"capsulecd/lib/config"
	"capsulecd/lib/engine"
)

func TestEngineChef(t *testing.T) {
	config.Init()
	config.Set("scm","github")
	config.Set("package_type","chef")

	//githubScm, err := scm.Create()
	//assert.NoError(t, err)
	chefEngine, err := engine.Create()
	assert.NoError(t, err)

	assert.Implements(t, (*engine.Engine)(nil), chefEngine, "should implement the Engine interface")

}
