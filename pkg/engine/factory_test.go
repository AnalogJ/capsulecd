package engine_test

import (
	"capsulecd/pkg/config/mock"
	"capsulecd/pkg/engine"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/scm/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type FactoryTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Scm          *mock_scm.MockInterface
	Config       *mock_config.MockInterface
	PipelineData *pipeline.Data
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *FactoryTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Scm = mock_scm.NewMockInterface(suite.MockCtrl)

}

func (suite *FactoryTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

func (suite *FactoryTestSuite) TestCreate_Invalid() {
	//test
	testEngine, cerr := engine.Create("invalidtype", suite.PipelineData, suite.Config, suite.Scm)

	//assert
	require.Error(suite.T(), cerr, "should return an erro")
	require.Nil(suite.T(), testEngine, "engine should be nil")
}

func (suite *FactoryTestSuite) TestCreate_Chef() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)

	//test
	testEngine, cerr := engine.Create("chef", suite.PipelineData, suite.Config, suite.Scm)

	//assert
	require.NoError(suite.T(), cerr)
	require.NotNil(suite.T(), testEngine)
}

func (suite *FactoryTestSuite) TestCreate_Golang() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("scm").Return("github")
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/golang_analogj_test")
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("github.com/analogj/golang_analogj_test")

	//test
	testEngine, cerr := engine.Create("golang", suite.PipelineData, suite.Config, suite.Scm)

	//assert
	require.NoError(suite.T(), cerr)
	require.NotNil(suite.T(), testEngine)
}

func (suite *FactoryTestSuite) TestCreate_Node() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)

	//test
	testEngine, cerr := engine.Create("node", suite.PipelineData, suite.Config, suite.Scm)

	//assert
	require.NoError(suite.T(), cerr)
	require.NotNil(suite.T(), testEngine)
}

func (suite *FactoryTestSuite) TestCreate_Python() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)

	//test
	testEngine, cerr := engine.Create("python", suite.PipelineData, suite.Config, suite.Scm)

	//assert
	require.NoError(suite.T(), cerr)
	require.NotNil(suite.T(), testEngine)
}

func (suite *FactoryTestSuite) TestCreate_Ruby() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)

	//test
	testEngine, cerr := engine.Create("ruby", suite.PipelineData, suite.Config, suite.Scm)

	//assert
	require.NoError(suite.T(), cerr)
	require.NotNil(suite.T(), testEngine)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(FactoryTestSuite))
}
