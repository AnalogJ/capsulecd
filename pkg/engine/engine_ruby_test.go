// +build ruby

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

func TestEngineRuby_Create(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	require.NoError(t, err)

	testConfig.Set("scm", "github")
	testConfig.Set("package_type", "ruby")
	testConfig.Set("scm_github_access_token", "placeholder")
	pipelineData := new(pipeline.Data)
	githubScm, err := scm.Create("github", pipelineData, testConfig, nil)
	require.NoError(t, err)

	//test
	rubyEngine, err := engine.Create("ruby", pipelineData, testConfig, githubScm)

	//assert
	require.NoError(t, err)
	require.NotNil(t, rubyEngine)
}

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EngineRubyTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Scm          *mock_scm.MockInterface
	Config       *mock_config.MockInterface
	PipelineData *pipeline.Data
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *EngineRubyTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Scm = mock_scm.NewMockInterface(suite.MockCtrl)

}

func (suite *EngineRubyTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEngineRuby_TestSuite(t *testing.T) {
	suite.Run(t, new(EngineRubyTestSuite))
}

func (suite *EngineRubyTestSuite) TestEngineRuby_ValidateTools() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	rubyEngine, err := engine.Create("ruby", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := rubyEngine.ValidateTools()

	//assert
	require.NoError(suite.T(), berr)
}

func (suite *EngineRubyTestSuite) TestEngineRuby_AssembleStep() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("engine_version_bump_type").Return("patch").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "gem_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "ruby", "gem_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	rubyEngine, err := engine.Create("ruby", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := rubyEngine.AssembleStep()
	require.NoError(suite.T(), berr)

	//assert
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "Rakefile")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "spec")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, ".gitignore")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "Gemfile")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "gem_analogj_test-0.1.4.gem")))
}

func (suite *EngineRubyTestSuite) TestEngineRuby_AssembleStep_WithMinimalGem() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("engine_version_bump_type").Return("patch").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "gem_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "ruby", "minimal_gem_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	rubyEngine, err := engine.Create("ruby", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := rubyEngine.AssembleStep()
	require.NoError(suite.T(), berr)

	//assert
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "Rakefile")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "spec")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, ".gitignore")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "gem_analogj_test-0.1.4.gem")))
}

func (suite *EngineRubyTestSuite) TestEngineRuby_AssembleStep_WithoutGemspec() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "gem_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "ruby", "minimal_gem_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)
	os.Remove(path.Join(suite.PipelineData.GitLocalPath, "gem_analogj_test.gemspec"))

	rubyEngine, err := engine.Create("ruby", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := rubyEngine.AssembleStep()

	//assert
	require.Error(suite.T(), berr, "should return an error")
}

func (suite *EngineRubyTestSuite) TestEngineRuby_TestStep_AllDisabled() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetBool(gomock.Any()).Return(true).MinTimes(1)
	suite.Config.EXPECT().GetString("engine_cmd_test").Return("exit 0").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "gem_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "ruby", "gem_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	rubyEngine, err := engine.Create("ruby", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := rubyEngine.TestStep()

	//assert
	require.NoError(suite.T(), berr)
}

func (suite *EngineRubyTestSuite) TestEngineRuby_TestStep_LintFailure() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetBool(gomock.Any()).Return(false).MinTimes(1)
	suite.Config.EXPECT().GetString("engine_cmd_lint").Return("exit 1").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "gem_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "ruby", "gem_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	rubyEngine, err := engine.Create("ruby", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := rubyEngine.TestStep()

	//assert
	require.Error(suite.T(), berr)
}

func (suite *EngineRubyTestSuite) TestEngineRuby_TestStep_TestFailure() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetBool(gomock.Any()).Return(false).MinTimes(1)
	suite.Config.EXPECT().GetString("engine_cmd_lint").Return("exit 0").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_cmd_test").Return("exit 1").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "gem_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "ruby", "gem_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	rubyEngine, err := engine.Create("ruby", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := rubyEngine.TestStep()

	//assert
	require.Error(suite.T(), berr)
}

func (suite *EngineRubyTestSuite) TestEngineRuby_TestStep_SecurityCheckFailure() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetBool(gomock.Any()).Return(false).MinTimes(1)
	suite.Config.EXPECT().GetString("engine_cmd_lint").Return("exit 0").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_cmd_test").Return("exit 0").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_cmd_security_check").Return("exit 1").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "gem_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "ruby", "gem_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	rubyEngine, err := engine.Create("ruby", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := rubyEngine.TestStep()

	//assert
	require.Error(suite.T(), berr)
}

//func (suite *EngineRubyTestSuite) TestEngineRuby_PackageStep_WithoutLockFiles() {
//	//setup
//	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
//	suite.Config.EXPECT().GetBool("mgr_keep_lock_file").MinTimes(1).Return(false)
//
//	//copy cookbook fixture into a temp directory.
//	parentPath, err := ioutil.TempDir("", "")
//	require.NoError(suite.T(), err)
//	defer os.RemoveAll(parentPath)
//	suite.PipelineData.GitParentPath = parentPath
//	cpath, cerr := utils.GitClone(parentPath, "gem_analogj_test", "https://github.com/AnalogJ/gem_analogj_test.git")
//	require.NoError(suite.T(), cerr)
//	suite.PipelineData.GitLocalPath = cpath
//
//	rubyEngine, err := engine.Create("ruby", suite.PipelineData, suite.Config, suite.Scm)
//	require.NoError(suite.T(), err)
//
//	//test
//	berr := rubyEngine.PackageStep()
//
//	//assert
//	require.NoError(suite.T(), berr)
//	require.False(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "Gemfile.lock")))
//}
