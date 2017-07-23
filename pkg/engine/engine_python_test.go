// +build python

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


func TestEnginePython_Create(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	require.NoError(t, err)

	testConfig.Set("scm", "github")
	testConfig.Set("package_type", "python")
	testConfig.Set("scm_github_access_token","placeholder")
	pipelineData := new(pipeline.Data)
	githubScm, err := scm.Create("github", pipelineData, testConfig, nil)
	require.NoError(t, err)


	//test
	pythonEngine, err := engine.Create("python", pipelineData, testConfig, githubScm)

	//assert
	require.NoError(t, err)
	require.NotNil(t, pythonEngine)
	require.Equal(t, "https://upload.pypi.org/legacy/", testConfig.GetString("pypi_repository"), "should load engine defaults")
}


// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EnginePythonTestSuite struct {
	suite.Suite
	MockCtrl *gomock.Controller
	Scm *mock_scm.MockInterface
	Config *mock_config.MockInterface
	PipelineData *pipeline.Data
}


// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *EnginePythonTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Scm = mock_scm.NewMockInterface(suite.MockCtrl)

}

func  (suite *EnginePythonTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEnginePython_TestSuite(t *testing.T) {
	suite.Run(t, new(EnginePythonTestSuite))
}

func (suite *EnginePythonTestSuite)TestEnginePython_AssembleStep() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(),gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("engine_version_bump_type").Return("patch")

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "pip_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "python", "pip_analogj_test"), suite.PipelineData.GitLocalPath )
	require.NoError(suite.T(), cerr)

	pythonEngine, err := engine.Create("python", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := pythonEngine.AssembleStep()
	require.NoError(suite.T(), berr)

	//assert
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "tox.ini")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "tests","__init__.py")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, ".gitignore")))
}

func (suite *EnginePythonTestSuite)TestEnginePython_AssembleStep_WithMinimalCookbook() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(),gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("engine_version_bump_type").Return("patch")

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "pip_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "python", "minimal_pip_analogj_test"), suite.PipelineData.GitLocalPath )
	require.NoError(suite.T(), cerr)

	pythonEngine, err := engine.Create("python", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := pythonEngine.AssembleStep()
	require.NoError(suite.T(), berr)

	//assert
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "VERSION")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "tox.ini")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "tests","__init__.py")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, ".gitignore")))
}

func (suite *EnginePythonTestSuite)TestEnginePython_AssembleStep_WithoutSetupPy() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(),gomock.Any()).MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "pip_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "python", "minimal_pip_analogj_test"), suite.PipelineData.GitLocalPath )
	require.NoError(suite.T(), cerr)
	os.Remove(path.Join(suite.PipelineData.GitLocalPath, "setup.py"))

	pythonEngine, err := engine.Create("python", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := pythonEngine.AssembleStep()

	//assert
	require.Error(suite.T(), berr, "should return an error")
}

func (suite *EnginePythonTestSuite)TestEnginePython_DependenciesStep() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(),gomock.Any()).MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "pip_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "python", "pip_analogj_test"), suite.PipelineData.GitLocalPath )
	require.NoError(suite.T(), cerr)

	pythonEngine, err := engine.Create("python", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := pythonEngine.DependenciesStep()

	//assert
	require.NoError(suite.T(), berr)
	//require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "Berksfile.lock")))
	//no lock files created by Python engine, and dependencies are installed by Tox in TestStep
	//should be a noop
}

func (suite *EnginePythonTestSuite)TestEnginePython_TestStep_AllDisabled() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(),gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetBool(gomock.Any()).MinTimes(1).Return(true)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "pip_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "python", "pip_analogj_test"), suite.PipelineData.GitLocalPath )
	require.NoError(suite.T(), cerr)

	pythonEngine, err := engine.Create("python", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := pythonEngine.TestStep()

	//assert
	require.NoError(suite.T(), berr)
}

func (suite *EnginePythonTestSuite)TestEnginePython_TestStep_LintFailure() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(),gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetBool(gomock.Any()).MinTimes(1).Return(false)
	suite.Config.EXPECT().GetString("engine_cmd_lint").Return("exit 1")

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "pip_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "python", "pip_analogj_test"), suite.PipelineData.GitLocalPath )
	require.NoError(suite.T(), cerr)

	pythonEngine, err := engine.Create("python", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := pythonEngine.TestStep()

	//assert
	require.Error(suite.T(), berr)
}

func (suite *EnginePythonTestSuite)TestEnginePython_TestStep_TestFailure() {
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
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "pip_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "python", "pip_analogj_test"), suite.PipelineData.GitLocalPath )
	require.NoError(suite.T(), cerr)

	pythonEngine, err := engine.Create("python", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := pythonEngine.TestStep()

	//assert
	require.Error(suite.T(), berr)
}

func (suite *EnginePythonTestSuite)TestEnginePython_TestStep_SecurityCheckFailure() {
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
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "pip_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "python", "pip_analogj_test"), suite.PipelineData.GitLocalPath )
	require.NoError(suite.T(), cerr)

	pythonEngine, err := engine.Create("python", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := pythonEngine.TestStep()

	//assert
	require.Error(suite.T(), berr)
}

func (suite *EnginePythonTestSuite)TestEnginePython_PackageStep() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(),gomock.Any()).MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	cpath, cerr := utils.GitClone(parentPath, "pip_analogj_test", "https://github.com/AnalogJ/pip_analogj_test.git")
	require.NoError(suite.T(), cerr)
	suite.PipelineData.GitLocalPath = cpath

	pythonEngine, err := engine.Create("python", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := pythonEngine.PackageStep()

	//assert
	require.NoError(suite.T(), berr)
}

func (suite *EnginePythonTestSuite)TestEnginePython_DistStep_WithoutCredentials() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(),gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().IsSet("pypi_username").MinTimes(1).Return(false)

	pythonEngine, err := engine.Create("python", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := pythonEngine.DistStep()

	//assert
	require.Error(suite.T(), berr)
}