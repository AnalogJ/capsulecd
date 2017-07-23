package engine

import (
	"capsulecd/pkg/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEngineBase_BumpVersion_Patch(t *testing.T) {

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

func TestEngineBase_BumpVersion_Minor(t *testing.T) {

	//setup
	testConfig, _ := config.Create()
	testConfig.Set("engine_version_bump_type", "minor")
	eng := engineBase{
		Config: testConfig,
	}

	//test
	ver, err := eng.BumpVersion("1.2.2")
	require.Nil(t, err)

	//assert
	require.Equal(t, ver, "1.3.0", "should correctly do a patch bump")
}

func TestEngineBase_BumpVersion_Major(t *testing.T) {

	//setup
	testConfig, _ := config.Create()
	testConfig.Set("engine_version_bump_type", "major")
	eng := engineBase{
		Config: testConfig,
	}

	//test
	ver, err := eng.BumpVersion("1.2.2")
	require.Nil(t, err)

	//assert
	require.Equal(t, ver, "2.0.0", "should correctly do a patch bump")
}

func TestEngineBase_BumpVersion_InvalidCurrentVersion(t *testing.T) {

	//setup
	testConfig, _ := config.Create()
	eng := engineBase{
		Config: testConfig,
	}

	//test
	nextV, err := eng.BumpVersion("abcde")

	//assert
	require.Error(t, err, "should return an error if unparsable version")
	require.Empty(t, nextV, "should be empty next version")
}

func TestEngineBase_BumpVersion_WithVPrefix(t *testing.T) {

	//setup
	testConfig, _ := config.Create()
	eng := engineBase{
		Config: testConfig,
	}

	//test
	nextV, err := eng.BumpVersion("v1.2.3")
	require.Nil(t, err)

	//assert
	require.Equal(t, nextV, "1.2.4", "should correctly do a patch bump")
}
