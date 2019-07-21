package engine

import (
	"github.com/stretchr/testify/suite"
	"github.com/golang/mock/gomock"
	"github.com/analogj/capsulecd/pkg/scm/mock"
	"github.com/analogj/capsulecd/pkg/config/mock"
	"github.com/analogj/capsulecd/pkg/pipeline"
	"testing"
	"github.com/stretchr/testify/require"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EngineGolangImplTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Scm          *mock_scm.MockInterface
	Config       *mock_config.MockInterface
	PipelineData *pipeline.Data
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *EngineGolangImplTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Scm = mock_scm.NewMockInterface(suite.MockCtrl)

}

func (suite *EngineGolangImplTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEngineGolang_TestSuite(t *testing.T) {
	suite.Run(t, new(EngineGolangImplTestSuite))
}

func (suite *EngineGolangImplTestSuite) TestEngineGolang_Init_GithubPackagePath() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("scm").Return("github")
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/golang_analogj_test")
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("github.com/analogj/golang_analogj_test")

	//test
	golangEngine := new(engineGolang)
	err := golangEngine.Init(suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//assert
	require.Equal(suite.T(), "src/github.com/analogj", golangEngine.PipelineData.GitParentPath)
}

func (suite *EngineGolangImplTestSuite) TestEngineGolang_Init_BitbucketPackagePath() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("scm").Return("bitbucket")
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("sparktree/golang_analogj_test")
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("bitbucket.org/sparktree/golang_analogj_test")

	//test
	golangEngine := new(engineGolang)
	err := golangEngine.Init(suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//assert
	require.Equal(suite.T(), "src/bitbucket.org/sparktree", golangEngine.PipelineData.GitParentPath)
}

func (suite *EngineGolangImplTestSuite) TestEngineGolang_Init_SimplePackagePath() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("scm").Return("bitbucket")
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("analogj/golang_analogj_test")
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("capsulecd")

	//test
	golangEngine := new(engineGolang)
	err := golangEngine.Init(suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//assert
	require.Equal(suite.T(), "src", golangEngine.PipelineData.GitParentPath)
}
