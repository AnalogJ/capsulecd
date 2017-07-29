package scm_test

import (
	"capsulecd/pkg/config/mock"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/scm"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ScmTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Config       *mock_config.MockInterface
	PipelineData *pipeline.Data
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *ScmTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)

}

func (suite *ScmTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

func (suite *ScmTestSuite) TestCreate_Invalid() {
	//test
	testEngine, cerr := scm.Create("invalidtype", suite.PipelineData, suite.Config, nil)

	//assert
	require.Error(suite.T(), cerr, "should return an erro")
	require.Nil(suite.T(), testEngine, "engine should be nil")
}

func (suite *ScmTestSuite) TestCreate_Github() {
	//setup
	suite.Config.EXPECT().GetString("scm_github_access_token").Return("placeholder")
	suite.Config.EXPECT().IsSet("scm_github_api_endpoint").Return(false)
	suite.Config.EXPECT().IsSet("scm_github_access_token").Return(true)
	suite.Config.EXPECT().IsSet("scm_git_parent_path").Return(false)

	//test
	testScm, cerr := scm.Create("github", suite.PipelineData, suite.Config, nil)

	//assert
	require.NoError(suite.T(), cerr)
	require.NotNil(suite.T(), testScm)
}

func (suite *ScmTestSuite) TestCreate_Bitbucket() {
	//setup
	//suite.Config.EXPECT().SetDefault(gomock.Any(),gomock.Any()).MinTimes(1)

	//test
	testScm, cerr := scm.Create("bitbucket", suite.PipelineData, suite.Config, nil)

	//assert
	require.NoError(suite.T(), cerr)
	require.NotNil(suite.T(), testScm)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(ScmTestSuite))
}
