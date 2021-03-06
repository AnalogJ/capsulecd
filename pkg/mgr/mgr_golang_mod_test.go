// +build golang

package mgr_test

import (
	"github.com/stretchr/testify/suite"
	"github.com/golang/mock/gomock"
	"github.com/analogj/capsulecd/pkg/mgr/mock"
	"github.com/analogj/capsulecd/pkg/config/mock"
	"github.com/analogj/capsulecd/pkg/pipeline"
	"testing"
	"io/ioutil"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"github.com/analogj/capsulecd/pkg/metadata"
	"github.com/analogj/capsulecd/pkg/utils"
	"github.com/analogj/capsulecd/pkg/mgr"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type MgrGolangModTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Mgr          *mock_mgr.MockInterface
	Config       *mock_config.MockInterface
	PipelineData *pipeline.Data
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *MgrGolangModTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Mgr = mock_mgr.NewMockInterface(suite.MockCtrl)

}

func (suite *MgrGolangModTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestMgrGolangMod_TestSuite(t *testing.T) {
	suite.Run(t, new(MgrGolangModTestSuite))
}

func (suite *MgrGolangModTestSuite) TestMgrGolangModTestSuite_DependenciesStep() {
	//setup
	//suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GolangGoPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "src", "mod_analogj_test")
	os.MkdirAll(path.Join(parentPath, "src"),0666)
	cerr := utils.CopyDir(path.Join("testdata", "golang", "mod_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	mgrGolangMod, err := mgr.Create("mod", suite.PipelineData, suite.Config, nil)
	require.NoError(suite.T(), err)
	currentVersion := new(metadata.GolangMetadata)
	nextVersion := new(metadata.GolangMetadata)

	//test
	berr := mgrGolangMod.MgrDependenciesStep(currentVersion, nextVersion)

	//assert
	require.NoError(suite.T(), berr)
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "go.mod")))

}


func (suite *MgrGolangModTestSuite) TestMgrGolangModTestSuite_MgrDistStep_WithoutCredentials() {
	//setup
	//suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	mgrGolangMod, err := mgr.Create("mod", suite.PipelineData, suite.Config, nil)
	require.NoError(suite.T(), err)
	currentVersion := new(metadata.GolangMetadata)
	nextVersion := new(metadata.GolangMetadata)

	//test
	berr := mgrGolangMod.MgrDistStep(currentVersion, nextVersion)

	//assert
	require.NoError(suite.T(), berr)
}
