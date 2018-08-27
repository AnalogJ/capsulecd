// +build chef
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
	"capsulecd/pkg/mgr"
	"capsulecd/pkg/utils"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type MgrChefBerkshelfTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Mgr          *mock_mgr.MockInterface
	Config       *mock_config.MockInterface
	PipelineData *pipeline.Data
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *MgrChefBerkshelfTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Mgr = mock_mgr.NewMockInterface(suite.MockCtrl)

}

func (suite *MgrChefBerkshelfTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestMgrChefBerkshelf_TestSuite(t *testing.T) {
	suite.Run(t, new(MgrChefBerkshelfTestSuite))
}

func (suite *MgrChefBerkshelfTestSuite) TestMgrChefBerkshelfTestSuite_DependenciesStep() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "cookbook_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "chef", "cookbook_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	mgrChefBerkshelf, err := mgr.Create("berkshelf", suite.PipelineData, suite.Config, nil)
	require.NoError(suite.T(), err)

	currentVersion := new(metadata.ChefMetadata)
	nextVersion := new(metadata.ChefMetadata)

	//test
	berr := mgrChefBerkshelf.MgrDependenciesStep(currentVersion, nextVersion)

	//assert
	require.NoError(suite.T(), berr)
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "Berksfile.lock")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "Gemfile.lock")))

}


func (suite *MgrChefBerkshelfTestSuite) TestMgrChefBerkshelfTestSuite_WithoutCredentials() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().IsSet("chef_supermarket_username").MinTimes(1).Return(false)

	mgrChefBerkshelf, err := mgr.Create("berkshelf", suite.PipelineData, suite.Config, nil)
	require.NoError(suite.T(), err)

	currentVersion := new(metadata.ChefMetadata)
	nextVersion := new(metadata.ChefMetadata)

	//test
	berr := mgrChefBerkshelf.MgrDistStep(currentVersion, nextVersion)

	//assert
	require.Error(suite.T(), berr)
}

//
//// junk username/password only for use on test.pypi.org
//// username: capsulecd
//// password: capsulecd$23$
//// we're not going to mock out this test, as we want to ensure that package manager integration works correctly, so we'll just
//// communicate with the test pypi server.
//func (suite *MgrPythonPipTestSuite) TestMgrChefBerkshelfTestSuite_DistStep_WithCredentials() {
//	//setup
//	//suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
//	suite.Config.EXPECT().IsSet("pypi_username").MinTimes(1).Return(true)
//	suite.Config.EXPECT().IsSet("pypi_password").MinTimes(1).Return(true)
//	suite.Config.EXPECT().GetString("pypi_username").MinTimes(1).Return("capsulecd")
//	suite.Config.EXPECT().GetString("pypi_password").MinTimes(1).Return("capsulecd$23$")
//	suite.Config.EXPECT().GetString("pypi_repository").MinTimes(1).Return("https://test.pypi.org/legacy/") //using test repo
//
//
//	//copy cookbook fixture into a temp directory.
//	parentPath, err := ioutil.TempDir("", "")
//	require.NoError(suite.T(), err)
//	defer os.RemoveAll(parentPath)
//	suite.PipelineData.GitParentPath = parentPath
//	suite.PipelineData.GitLocalPath = path.Join(parentPath, "pip_analogj_test")
//	cerr := utils.CopyDir(path.Join("testdata", "python", "pip_analogj_test"), suite.PipelineData.GitLocalPath)
//	require.NoError(suite.T(), cerr)
//
//
//	//using current date/time as a pseudo version number
//	t := time.Now()
//	dateVersion := t.Format("20060102.1504.05") //yyyymmdd.HHMM.SS
//	werr := ioutil.WriteFile(path.Join(suite.PipelineData.GitLocalPath, "VERSION"), []byte(dateVersion), 0644)
//	require.NoError(suite.T(), werr)
//
//
//	mgrPythonPip, err := mgr.Create("pip", suite.PipelineData, suite.Config, nil)
//	require.NoError(suite.T(), err)
//
//	currentVersion := new(metadata.PythonMetadata)
//	nextVersion := new(metadata.PythonMetadata)
//
//	//test
//	berr := mgrPythonPip.MgrDistStep(currentVersion, nextVersion)
//
//	//assert
//	require.NoError(suite.T(), berr)
//}

