// +build chef

package engine_test

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/engine"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/scm"
	"capsulecd/pkg/utils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"path"
	//"path/filepath"
	"capsulecd/pkg/config/mock"
	"capsulecd/pkg/scm/mock"
	"os"
	"testing"
)

func TestEngineChef_Create(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	require.NoError(t, err)

	testConfig.Set("scm", "github")
	testConfig.Set("package_type", "chef")
	testConfig.Set("scm_github_access_token", "placeholder")
	pipelineData := new(pipeline.Data)
	githubScm, err := scm.Create("github", pipelineData, testConfig, nil)
	require.NoError(t, err)

	//test
	chefEngine, err := engine.Create("chef", pipelineData, testConfig, githubScm)

	//assert
	require.NoError(t, err)
	require.NotNil(t, chefEngine)
	require.Equal(t, "Other", testConfig.GetString("chef_supermarket_type"), "should load engine defaults")
}

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EngineChefTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Scm          *mock_scm.MockInterface
	Config       *mock_config.MockInterface
	PipelineData *pipeline.Data
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *EngineChefTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Scm = mock_scm.NewMockInterface(suite.MockCtrl)

}

func (suite *EngineChefTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEngineChef_TestSuite(t *testing.T) {
	suite.Run(t, new(EngineChefTestSuite))
}

func (suite *EngineChefTestSuite) TestEngineChef_ValidateTools() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	chefEngine, err := engine.Create("chef", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := chefEngine.ValidateTools()

	//assert
	require.NoError(suite.T(), berr)
}

func (suite *EngineChefTestSuite) TestEngineChef_AssembleStep() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("engine_version_bump_type").Return("patch")

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "cookbook_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "chef", "cookbook_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	chefEngine, err := engine.Create("chef", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := chefEngine.AssembleStep()
	require.NoError(suite.T(), berr)

	//assert
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "Rakefile")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "Berksfile")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, ".gitignore")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "Gemfile")))
}

func (suite *EngineChefTestSuite) TestEngineChef_AssembleStep_WithMinimalCookbook() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("engine_version_bump_type").Return("patch")

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "cookbook_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "chef", "minimal_cookbook_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	chefEngine, err := engine.Create("chef", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := chefEngine.AssembleStep()
	require.NoError(suite.T(), berr)

	//assert
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "Rakefile")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "Berksfile")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, ".gitignore")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "Gemfile")))
}

func (suite *EngineChefTestSuite) TestEngineChef_AssembleStep_WithoutMetadata() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "cookbook_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "chef", "minimal_cookbook_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)
	os.Remove(path.Join(suite.PipelineData.GitLocalPath, "metadata.rb"))

	chefEngine, err := engine.Create("chef", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := chefEngine.AssembleStep()

	//assert
	require.Error(suite.T(), berr, "should return an error")
}

func (suite *EngineChefTestSuite) TestEngineChef_DependenciesStep() {
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

	chefEngine, err := engine.Create("chef", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := chefEngine.DependenciesStep()

	//assert
	require.NoError(suite.T(), berr)
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "Berksfile.lock")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "Gemfile.lock")))
}

func (suite *EngineChefTestSuite) TestEngineChef_TestStep_AllDisabled() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetBool(gomock.Any()).MinTimes(1).Return(true)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "cookbook_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "chef", "cookbook_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	chefEngine, err := engine.Create("chef", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := chefEngine.TestStep()

	//assert
	require.NoError(suite.T(), berr)
}

func (suite *EngineChefTestSuite) TestEngineChef_TestStep_LintFailure() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetBool(gomock.Any()).MinTimes(1).Return(false)
	suite.Config.EXPECT().GetString("engine_cmd_lint").Return("exit 1")

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "cookbook_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "chef", "cookbook_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	chefEngine, err := engine.Create("chef", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := chefEngine.TestStep()

	//assert
	require.Error(suite.T(), berr)
}

func (suite *EngineChefTestSuite) TestEngineChef_TestStep_TestFailure() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetBool(gomock.Any()).MinTimes(1).Return(false)
	suite.Config.EXPECT().GetString("engine_cmd_lint").Return("exit 0")
	suite.Config.EXPECT().GetString("engine_cmd_test").Return("exit 1")

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "cookbook_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "chef", "cookbook_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	chefEngine, err := engine.Create("chef", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := chefEngine.TestStep()

	//assert
	require.Error(suite.T(), berr)
}

func (suite *EngineChefTestSuite) TestEngineChef_TestStep_SecurityCheckFailure() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetBool(gomock.Any()).MinTimes(1).Return(false)
	suite.Config.EXPECT().GetString("engine_cmd_lint").Return("exit 0")
	suite.Config.EXPECT().GetString("engine_cmd_test").Return("exit 0")
	suite.Config.EXPECT().GetString("engine_cmd_security_check").Return("exit 1")

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "cookbook_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "chef", "cookbook_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	chefEngine, err := engine.Create("chef", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := chefEngine.TestStep()

	//assert
	require.Error(suite.T(), berr)
}

func (suite *EngineChefTestSuite) TestEngineChef_PackageStep_WithoutLockFiles() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetBool("engine_package_keep_lock_file").MinTimes(1).Return(false)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	cpath, cerr := utils.GitClone(parentPath, "cookbook_analogj_test", "https://github.com/AnalogJ/cookbook_analogj_test.git")
	require.NoError(suite.T(), cerr)
	suite.PipelineData.GitLocalPath = cpath

	chefEngine, err := engine.Create("chef", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := chefEngine.PackageStep()

	//assert
	require.NoError(suite.T(), berr)
	require.False(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "Berksfile.lock")))
	require.False(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "Gemfile.lock")))
}

func (suite *EngineChefTestSuite) TestEngineChef_DistStep_WithoutCredentials() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().IsSet("chef_supermarket_username").MinTimes(1).Return(false)

	chefEngine, err := engine.Create("chef", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := chefEngine.DistStep()

	//assert
	require.Error(suite.T(), berr)
}
