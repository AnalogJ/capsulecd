// +build node
package mgr_test

import (
	"github.com/stretchr/testify/suite"
	"github.com/golang/mock/gomock"
	"capsulecd/pkg/mgr/mock"
	"capsulecd/pkg/config/mock"
	"capsulecd/pkg/pipeline"
	"testing"
	"io/ioutil"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"capsulecd/pkg/metadata"
	"capsulecd/pkg/utils"
	"capsulecd/pkg/mgr"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type MgrNodeNpmTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Mgr          *mock_mgr.MockInterface
	Config       *mock_config.MockInterface
	PipelineData *pipeline.Data
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *MgrNodeNpmTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Mgr = mock_mgr.NewMockInterface(suite.MockCtrl)

}

func (suite *MgrNodeNpmTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestMgrNodeNpm_TestSuite(t *testing.T) {
	suite.Run(t, new(MgrNodeNpmTestSuite))
}

func (suite *MgrNodeNpmTestSuite) TestMgrNodeNpmTestSuite_DependenciesStep() {
	//setup
	//suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "npm_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "node", "npm_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	mgrNodeNpm, err := mgr.Create("npm", suite.PipelineData, suite.Config, nil)
	require.NoError(suite.T(), err)
	currentVersion := new(metadata.NodeMetadata)
	nextVersion := new(metadata.NodeMetadata)

	//test
	berr := mgrNodeNpm.MgrDependenciesStep(currentVersion, nextVersion)

	//assert
	require.NoError(suite.T(), berr)
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "package.json")))

}


func (suite *MgrNodeNpmTestSuite) TestMgrNodeNpmTestSuite_MgrDistStep_WithoutCredentials() {
	//setup
	//suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	mgrNodeNpm, err := mgr.Create("npm", suite.PipelineData, suite.Config, nil)
	require.NoError(suite.T(), err)
	currentVersion := new(metadata.NodeMetadata)
	nextVersion := new(metadata.NodeMetadata)

	//test
	berr := mgrNodeNpm.MgrDistStep(currentVersion, nextVersion)

	//assert
	require.Error(suite.T(), berr)
}