// +build golang

package engine_test

import (
	"github.com/analogj/capsulecd/pkg/config"
	"github.com/analogj/capsulecd/pkg/engine"
	"github.com/analogj/capsulecd/pkg/pipeline"
	"github.com/analogj/capsulecd/pkg/scm"
	"github.com/analogj/capsulecd/pkg/utils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"path"
	//"path/filepath"
	"github.com/analogj/capsulecd/pkg/config/mock"
	"github.com/analogj/capsulecd/pkg/scm/mock"
	"os"
	"testing"
)

func TestEngineGolang_Create(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	require.NoError(t, err)

	testConfig.Set("scm", "github")
	testConfig.Set("package_type", "golang")
	testConfig.Set("scm_github_access_token", "placeholder")
	pipelineData := new(pipeline.Data)
	githubScm, err := scm.Create("github", pipelineData, testConfig, nil)
	require.NoError(t, err)

	//test
	golangEngine, err := engine.Create("golang", pipelineData, testConfig, githubScm)

	//assert
	require.NoError(t, err)
	require.NotNil(t, golangEngine)
	require.Equal(t, "exit 0", testConfig.GetString("engine_cmd_security_check"), "should load engine defaults")
}

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EngineGolangTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Scm          *mock_scm.MockInterface
	Config       *mock_config.MockInterface
	PipelineData *pipeline.Data
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *EngineGolangTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Scm = mock_scm.NewMockInterface(suite.MockCtrl)

}

func (suite *EngineGolangTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEngineGolang_TestSuite(t *testing.T) {
	suite.Run(t, new(EngineGolangTestSuite))
}

func (suite *EngineGolangTestSuite) TestEngineGolang_ValidateTools() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("scm").Return("github").MinTimes(1)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("github.com/analogj/golang_analogj_test").MinTimes(1)

	golangEngine, err := engine.Create("golang", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := golangEngine.ValidateTools()

	//assert
	require.NoError(suite.T(), berr)
}

func (suite *EngineGolangTestSuite) TestEngineGolang_AssembleStep() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("engine_version_bump_type").Return("patch").MinTimes(1)
	suite.Config.EXPECT().GetString("scm").Return("github").MinTimes(1)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("github.com/analogj/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_version_metadata_path").Return("pkg/version/version.go").MinTimes(1)


	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "golang_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "golang", "golang_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	golangEngine, err := engine.Create("golang", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := golangEngine.AssembleStep()
	require.NoError(suite.T(), berr)

	//assert
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, ".gitignore")))
}

func (suite *EngineGolangTestSuite) TestEngineGolang_AssembleStep_WithMinimalCookbook() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("engine_version_bump_type").Return("patch").MinTimes(1)
	suite.Config.EXPECT().GetString("scm").Return("github").MinTimes(1)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("github.com/analogj/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_version_metadata_path").Return("pkg/version/version.go").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "golang_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "golang", "minimal_golang_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	golangEngine, err := engine.Create("golang", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := golangEngine.AssembleStep()
	require.NoError(suite.T(), berr)

	//assert
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, ".gitignore")))
}

func (suite *EngineGolangTestSuite) TestEngineGolang_AssembleStep_WithoutVersion() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("scm").Return("github").MinTimes(1)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("github.com/analogj/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_version_metadata_path").Return("pkg/version/version.go").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "golang_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "golang", "minimal_golang_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)
	os.Remove(path.Join(suite.PipelineData.GitLocalPath, "pkg", "version", "version.go"))

	golangEngine, err := engine.Create("golang", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := golangEngine.AssembleStep()

	//assert
	require.Error(suite.T(), berr, "should return an error")
}

func (suite *EngineGolangTestSuite) TestEngineGolang_TestStep_AllDisabled() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("scm").Return("github").MinTimes(1)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("github.com/analogj/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetBool(gomock.Any()).Return(true).MinTimes(1)
	suite.Config.EXPECT().GetString("engine_cmd_test").Return("exit 0").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "golang_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "golang", "golang_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	golangEngine, err := engine.Create("golang", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := golangEngine.TestStep()

	//assert
	require.NoError(suite.T(), berr)
}

func (suite *EngineGolangTestSuite) TestEngineGolang_TestStep_LintFailure() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("scm").Return("github").MinTimes(1)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("github.com/analogj/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetBool(gomock.Any()).Return(false).MinTimes(1)
	suite.Config.EXPECT().GetString("engine_cmd_lint").Return("exit 1").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "golang_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "golang", "golang_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	golangEngine, err := engine.Create("golang", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := golangEngine.TestStep()

	//assert
	require.Error(suite.T(), berr)
}

func (suite *EngineGolangTestSuite) TestEngineGolang_TestStep_TestFailure() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("scm").Return("github").MinTimes(1)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("github.com/analogj/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetBool(gomock.Any()).Return(false).MinTimes(1)
	suite.Config.EXPECT().GetString("engine_cmd_lint").Return("exit 0").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_cmd_test").Return("exit 1").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "golang_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "golang", "golang_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	golangEngine, err := engine.Create("golang", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := golangEngine.TestStep()

	//assert
	require.Error(suite.T(), berr)
}

func (suite *EngineGolangTestSuite) TestEngineGolang_TestStep_SecurityCheckFailure() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("scm").Return("github").MinTimes(1)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("github.com/analogj/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetBool(gomock.Any()).Return(false).MinTimes(1)
	suite.Config.EXPECT().GetString("engine_cmd_lint").Return("exit 0").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_cmd_test").Return("exit 0").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_cmd_security_check").Return("exit 1").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "golang_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "golang", "golang_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	golangEngine, err := engine.Create("golang", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := golangEngine.TestStep()

	//assert
	require.Error(suite.T(), berr)
}

func (suite *EngineGolangTestSuite) TestEngineGolang_PackageStep_WithoutLockFiles() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("scm").Return("github").MinTimes(1)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("github.com/analogj/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_version_bump_msg").Return("Automated packaging of release by CapsuleCD").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_git_author_name").Return("CapsuleCD").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_git_author_email").Return("CapsuleCD@users.noreply.github.com").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	cpath, cerr := utils.GitClone(parentPath, "golang_analogj_test", "https://github.com/AnalogJ/golang_analogj_test.git")
	require.NoError(suite.T(), cerr)
	suite.PipelineData.GitLocalPath = cpath

	golangEngine, err := engine.Create("golang", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := golangEngine.PackageStep()

	//assert
	require.NoError(suite.T(), berr)
	require.False(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "glide.lock")))
}
