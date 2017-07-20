package engine

import (
	"testing"
	"capsulecd/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestEngineBase_BumpVersion(t *testing.T) {

	//setup
	testConfig, _ := config.Create()
	eng := engineBase{
		Config: testConfig,
	}

	//test
	ver, err := eng.BumpVersion("1.2.2");
	assert.Nil(t, err)

	//assert
	assert.Equal(t,  ver, "1.2.3", "should correctly do a patch bump")
}
