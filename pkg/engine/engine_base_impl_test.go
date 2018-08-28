package engine

import (
	"github.com/stretchr/testify/require"
	"testing"
	"capsulecd/pkg/config/mock"
	"github.com/golang/mock/gomock"
)

func TestEngineBase_BumpVersion_Patch(t *testing.T) {

	//setup
	mockCtrl := gomock.NewController(t)
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	fakeConfig.EXPECT().GetString("engine_version_bump_type").MinTimes(1).Return("patch")
	eng := engineBase{
		Config: fakeConfig,
	}

	//test
	ver, err := eng.BumpVersion("1.2.2")
	require.Nil(t, err)

	ver2, err := eng.BumpVersion("1.0.0")
	require.Nil(t, err)

	//assert
	require.Equal(t, ver, "1.2.3", "should correctly do a patch bump")
	require.Equal(t, ver2, "1.0.1", "should correctly do a patch bump")
}

func TestEngineBase_BumpVersion_Minor(t *testing.T) {

	//setup
	mockCtrl := gomock.NewController(t)
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	fakeConfig.EXPECT().GetString("engine_version_bump_type").MinTimes(1).Return("minor")
	eng := engineBase{
		Config: fakeConfig,
	}


	//test
	ver, err := eng.BumpVersion("1.2.2")
	require.Nil(t, err)

	//assert
	require.Equal(t, ver, "1.3.0", "should correctly do a patch bump")
}

func TestEngineBase_BumpVersion_Major(t *testing.T) {

	//setup
	mockCtrl := gomock.NewController(t)
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	fakeConfig.EXPECT().GetString("engine_version_bump_type").MinTimes(1).Return("major")
	eng := engineBase{
		Config: fakeConfig,
	}

	//test
	ver, err := eng.BumpVersion("1.2.2")
	require.Nil(t, err)

	//assert
	require.Equal(t, ver, "2.0.0", "should correctly do a patch bump")
}

func TestEngineBase_BumpVersion_InvalidCurrentVersion(t *testing.T) {

	//setup
	mockCtrl := gomock.NewController(t)
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	fakeConfig.EXPECT().GetString("engine_version_bump_type").MinTimes(1).Return("patch")
	eng := engineBase{
		Config: fakeConfig,
	}

	//test
	nextV, err := eng.BumpVersion("abcde")

	//assert
	require.Error(t, err, "should return an error if unparsable version")
	require.Empty(t, nextV, "should be empty next version")
}

func TestEngineBase_BumpVersion_WithVPrefix(t *testing.T) {

	//setup
	mockCtrl := gomock.NewController(t)
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	fakeConfig.EXPECT().GetString("engine_version_bump_type").MinTimes(1).Return("patch")
	eng := engineBase{
		Config: fakeConfig,
	}

	//test
	nextV, err := eng.BumpVersion("v1.2.3")
	require.Nil(t, err)

	//assert
	require.Equal(t, nextV, "1.2.4", "should correctly do a patch bump")
}
