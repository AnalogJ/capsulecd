// +build node

package engine_test

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/engine"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/scm"
	"capsulecd/pkg/utils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/golang/mock/gomock"
	"io/ioutil"
	"path"
	//"path/filepath"
	"testing"
	"capsulecd/pkg/config/mock"
	"capsulecd/pkg/scm/mock"
	"os"
)


func TestEngineNode_Create(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	require.NoError(t, err)

	testConfig.Set("scm", "github")
	testConfig.Set("package_type", "node")
	testConfig.Set("scm_github_access_token","placeholder")
	pipelineData := new(pipeline.Data)
	githubScm, err := scm.Create("github", pipelineData, testConfig, nil)
	require.NoError(t, err)


	//test
	nodeEngine, err := engine.Create("node", pipelineData, testConfig, githubScm)

	//assert
	require.NoError(t, err)
	require.NotNil(t, nodeEngine)
}


// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EngineNodeTestSuite struct {
	suite.Suite
	MockCtrl *gomock.Controller
	Scm *mock_scm.MockInterface
	Config *mock_config.MockInterface
	PipelineData *pipeline.Data
}


// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *EngineNodeTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Scm = mock_scm.NewMockInterface(suite.MockCtrl)

}

func  (suite *EngineNodeTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEngineNode_TestSuite(t *testing.T) {
	suite.Run(t, new(EngineNodeTestSuite))
}

func (suite *EngineNodeTestSuite)TestEngineNode_AssembleStep() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(),gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("engine_version_bump_type").Return("patch")

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "npm_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "node", "npm_analogj_test"), suite.PipelineData.GitLocalPath )
	require.NoError(suite.T(), cerr)

	nodeEngine, err := engine.Create("node", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := nodeEngine.AssembleStep()
	require.NoError(suite.T(), berr)

	//assert
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, ".gitignore")))
}

func (suite *EngineNodeTestSuite)TestEngineNode_AssembleStep_WithoutPackageJson() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(),gomock.Any()).MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "npm_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "node", "npm_analogj_test"), suite.PipelineData.GitLocalPath )
	require.NoError(suite.T(), cerr)
	os.Remove(path.Join(suite.PipelineData.GitLocalPath, "package.json"))

	nodeEngine, err := engine.Create("node", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := nodeEngine.AssembleStep()

	//assert
	require.Error(suite.T(), berr, "should return an error")
}

func (suite *EngineNodeTestSuite)TestEngineNode_DependenciesStep() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(),gomock.Any()).MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "npm_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "node", "npm_analogj_test"), suite.PipelineData.GitLocalPath )
	require.NoError(suite.T(), cerr)

	nodeEngine, err := engine.Create("node", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := nodeEngine.DependenciesStep()

	//assert
	require.NoError(suite.T(), berr)
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "npm-shrinkwrap.json")))
}

func (suite *EngineNodeTestSuite)TestEngineNode_TestStep_AllDisabled() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(),gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetBool(gomock.Any()).MinTimes(1).Return(true)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "npm_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "node", "npm_analogj_test"), suite.PipelineData.GitLocalPath )
	require.NoError(suite.T(), cerr)

	nodeEngine, err := engine.Create("node", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := nodeEngine.TestStep()

	//assert
	require.NoError(suite.T(), berr)
}

func (suite *EngineNodeTestSuite)TestEngineNode_TestStep_LintFailure() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(),gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetBool(gomock.Any()).MinTimes(1).Return(false)
	suite.Config.EXPECT().GetString("engine_cmd_lint").Return("exit 1")

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "npm_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "node", "npm_analogj_test"), suite.PipelineData.GitLocalPath )
	require.NoError(suite.T(), cerr)

	nodeEngine, err := engine.Create("node", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := nodeEngine.TestStep()

	//assert
	require.Error(suite.T(), berr)
}

func (suite *EngineNodeTestSuite)TestEngineNode_TestStep_TestFailure() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(),gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetBool(gomock.Any()).MinTimes(1).Return(false)
	suite.Config.EXPECT().GetString("engine_cmd_lint").Return("exit 0")
	suite.Config.EXPECT().GetString("engine_cmd_test").Return("exit 1")

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "npm_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "node", "npm_analogj_test"), suite.PipelineData.GitLocalPath )
	require.NoError(suite.T(), cerr)

	nodeEngine, err := engine.Create("node", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := nodeEngine.TestStep()

	//assert
	require.Error(suite.T(), berr)
}

func (suite *EngineNodeTestSuite)TestEngineNode_TestStep_SecurityCheckFailure() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(),gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetBool(gomock.Any()).MinTimes(1).Return(false)
	suite.Config.EXPECT().GetString("engine_cmd_lint").Return("exit 0")
	suite.Config.EXPECT().GetString("engine_cmd_test").Return("exit 0")
	suite.Config.EXPECT().GetString("engine_cmd_security_check").Return("exit 1")

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "npm_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "node", "npm_analogj_test"), suite.PipelineData.GitLocalPath )
	require.NoError(suite.T(), cerr)

	nodeEngine, err := engine.Create("node", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := nodeEngine.TestStep()

	//assert
	require.Error(suite.T(), berr)
}

func (suite *EngineNodeTestSuite)TestEngineNode_PackageStep_WithoutLockFiles() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(),gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetBool("engine_package_keep_lock_file").MinTimes(1).Return(false)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	cpath, cerr := utils.GitClone(parentPath, "npm_analogj_test", "https://github.com/AnalogJ/npm_analogj_test.git")
	require.NoError(suite.T(), cerr)
	suite.PipelineData.GitLocalPath = cpath

	nodeEngine, err := engine.Create("node", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := nodeEngine.PackageStep()

	//assert
	require.NoError(suite.T(), berr)
	require.False(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "npm-shrinkwrap.json")))
}

func (suite *EngineNodeTestSuite)TestEngineNode_DistStep_WithoutCredentials() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(),gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().IsSet("npm_auth_token").MinTimes(1).Return(false)

	nodeEngine, err := engine.Create("node", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := nodeEngine.DistStep()

	//assert
	require.Error(suite.T(), berr)
}