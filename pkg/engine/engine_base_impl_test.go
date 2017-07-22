package engine

import (
	"capsulecd/pkg/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEngineBase_BumpVersion(t *testing.T) {

	//setup
	testConfig, _ := config.Create()
	eng := engineBase{
		Config: testConfig,
	}

	//test
	ver, err := eng.BumpVersion("1.2.2")
	require.Nil(t, err)

	//assert
	require.Equal(t, ver, "1.2.3", "should correctly do a patch bump")
}
