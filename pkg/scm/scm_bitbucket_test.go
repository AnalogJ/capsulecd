package scm_test

import (
	"capsulecd/pkg/config/mock"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/scm"
	"capsulecd/pkg/utils"
	"crypto/tls"
	"github.com/golang/mock/gomock"
	"github.com/seborama/govcr"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
)

//TODO: set to true when development is complete and no new recordings need to be created (CI mode enabled)
const DISABLE_RECORDINGS = false

func bitbucketVcrSetup(t *testing.T) *http.Client {
	tr := http.DefaultTransport.(*http.Transport)
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true, //disable certificate validation because we're playing back http requests.
	}
	insecureClient := http.Client{
		Transport: tr,
	}

	vcrConfig := govcr.VCRConfig{
		Logging:      true,
		CassettePath: path.Join("testdata", "govcr-fixtures"),
		Client:       &insecureClient,
		ExcludeHeaderFunc: func(key string) bool {
			// HTTP headers are case-insensitive
			//return strings.ToLower(key) == "user-agent" || strings.ToLower(key) == "accept"
			return strings.ToLower(key) == "user-agent" || strings.ToLower(key) == "authorization"
		},
		RequestFilterFunc: func(reqHeader http.Header, reqBody []byte) (*http.Header, *[]byte) {
			reqHeader.Set("Authorization", "Basic UExBQ0VIT0xERVI6UExBQ0VIT0xERVI=") //placeholder:placeholder

			return &reqHeader, &reqBody
		},

		//this line ensures that we do not attempt to create new recordings.
		//Comment this out if you would like to make recordings.
		DisableRecording: DISABLE_RECORDINGS,
	}

	vcr := govcr.NewVCR(t.Name(), &vcrConfig)
	return vcr.Client
}

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ScmBitbucketTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Config       *mock_config.MockInterface
	PipelineData *pipeline.Data
	Client       *http.Client

	Username string
	Password string
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *ScmBitbucketTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())
	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Client = bitbucketVcrSetup(suite.T())
	suite.PipelineData = new(pipeline.Data)

	if DISABLE_RECORDINGS {
		suite.Username = "PLACEHOLDER"
		suite.Password = "PLACEHOLDER"
	} else {
		suite.Username = os.Getenv("BITBUCKET_USERNAME")
		suite.Password = os.Getenv("BITBUCKET_PASSWORD")
	}
}

func (suite *ScmBitbucketTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestScmBitbucket_TestSuite(t *testing.T) {
	suite.Run(t, new(ScmBitbucketTestSuite))
}

func (suite *ScmBitbucketTestSuite) TestScmBitbucket_Init_WithoutUsername() {

	//setup
	suite.Config.EXPECT().IsSet("scm_bitbucket_username").Return(false)

	//test
	testScm, err := scm.Create("bitbucket", suite.PipelineData, suite.Config, suite.Client)

	//assert
	require.Nil(suite.T(), testScm)
	require.Error(suite.T(), err, "should raise an auth error")

}

func (suite *ScmBitbucketTestSuite) TestScmBitbucket_Init_WithoutPassword() {

	//setup
	suite.Config.EXPECT().IsSet("scm_bitbucket_username").Return(true)
	suite.Config.EXPECT().IsSet("scm_bitbucket_password").MinTimes(1).Return(false)
	suite.Config.EXPECT().IsSet("scm_bitbucket_access_token").MinTimes(1).Return(false)

	//test
	testScm, err := scm.Create("bitbucket", suite.PipelineData, suite.Config, suite.Client)

	//assert
	require.Nil(suite.T(), testScm)
	require.Error(suite.T(), err, "should raise an auth error")

}

func (suite *ScmBitbucketTestSuite) TestScmBitbucket_Init_WithGitParentPath() {

	//setup
	suite.Config.EXPECT().IsSet("scm_bitbucket_username").Return(true)
	suite.Config.EXPECT().IsSet("scm_bitbucket_password").MinTimes(1).Return(true)
	suite.Config.EXPECT().GetString("scm_bitbucket_username").Return(suite.Username)
	suite.Config.EXPECT().GetString("scm_bitbucket_password").MinTimes(1).Return(suite.Password)

	dirPath, err := ioutil.TempDir("", "")
	defer os.RemoveAll(dirPath)
	suite.Config.EXPECT().IsSet("scm_git_parent_path").Return(true)
	suite.Config.EXPECT().GetString("scm_git_parent_path").Return(dirPath)

	//test
	testScm, err := scm.Create("bitbucket", suite.PipelineData, suite.Config, suite.Client)
	require.Equal(suite.T(), suite.PipelineData.GitParentPath, dirPath, "should correctly set parent path to existing")

	//assert
	require.NotNil(suite.T(), testScm)
	require.Nil(suite.T(), err, "should not have an error")

}

func (suite *ScmBitbucketTestSuite) TestScmBitbucket_Init_WithDefaults() {

	//setup
	suite.Config.EXPECT().IsSet("scm_bitbucket_username").Return(true)
	suite.Config.EXPECT().IsSet("scm_bitbucket_password").MinTimes(1).Return(true)
	suite.Config.EXPECT().GetString("scm_bitbucket_username").Return(suite.Username)
	suite.Config.EXPECT().GetString("scm_bitbucket_password").MinTimes(1).Return(suite.Password)
	suite.Config.EXPECT().IsSet("scm_git_parent_path").Return(false)

	//test
	testScm, err := scm.Create("bitbucket", suite.PipelineData, suite.Config, nil)
	require.NotEmpty(suite.T(), suite.PipelineData.GitParentPath, "should correctly generate a temporary parent path")

	//assert
	require.NotNil(suite.T(), testScm)
	require.Nil(suite.T(), err, "should not have an error")

}

func (suite *ScmBitbucketTestSuite) TestScmBitbucket_RetrievePayload_PullRequest() {
	//setup
	suite.Config.EXPECT().IsSet("scm_bitbucket_username").Return(true)
	suite.Config.EXPECT().IsSet("scm_bitbucket_password").MinTimes(1).Return(true)
	suite.Config.EXPECT().GetString("scm_bitbucket_username").Return(suite.Username)
	suite.Config.EXPECT().GetString("scm_bitbucket_password").MinTimes(1).Return(suite.Password)
	suite.Config.EXPECT().IsSet("scm_git_parent_path").Return(false)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("sparktree/gem_analogj_test")
	suite.Config.EXPECT().GetString("scm_pull_request").Return("1")
	suite.Config.EXPECT().IsSet("scm_pull_request").Return(true)

	//test
	testScm, err := scm.Create("bitbucket", suite.PipelineData, suite.Config, suite.Client)
	require.NoError(suite.T(), err)
	payload, perr := testScm.RetrievePayload()
	require.NoError(suite.T(), perr)

	//assert
	require.NotEmpty(suite.T(), payload, "payload must be set after source Init")
	require.Equal(suite.T(), "1", payload.PullRequestNumber)
	require.True(suite.T(), suite.PipelineData.IsPullRequest)

}

func (suite *ScmBitbucketTestSuite) TestScmBitbucket_RetrievePayload_PullRequest_InvalidState() {
	//setup
	suite.Config.EXPECT().IsSet("scm_bitbucket_username").Return(true)
	suite.Config.EXPECT().IsSet("scm_bitbucket_password").MinTimes(1).Return(true)
	suite.Config.EXPECT().GetString("scm_bitbucket_username").Return(suite.Username)
	suite.Config.EXPECT().GetString("scm_bitbucket_password").MinTimes(1).Return(suite.Password)
	suite.Config.EXPECT().IsSet("scm_git_parent_path").Return(false)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("sparktree/gem_analogj_test")
	suite.Config.EXPECT().GetString("scm_pull_request").Return("2")
	suite.Config.EXPECT().IsSet("scm_pull_request").Return(true)

	//test
	testScm, err := scm.Create("bitbucket", suite.PipelineData, suite.Config, suite.Client)
	require.NoError(suite.T(), err)
	payload, perr := testScm.RetrievePayload()

	//assert
	require.Error(suite.T(), perr, "should return an error when PR is closed")
	require.Nil(suite.T(), payload)

}

func (suite *ScmBitbucketTestSuite) TestScmBitbucket_RetrievePayload_Push() {

	//setup
	suite.Config.EXPECT().IsSet("scm_bitbucket_username").Return(true)
	suite.Config.EXPECT().IsSet("scm_bitbucket_password").MinTimes(1).Return(true)
	suite.Config.EXPECT().GetString("scm_bitbucket_username").Return(suite.Username)
	suite.Config.EXPECT().GetString("scm_bitbucket_password").MinTimes(1).Return(suite.Password)
	suite.Config.EXPECT().IsSet("scm_git_parent_path").Return(false)
	suite.Config.EXPECT().IsSet("scm_pull_request").Return(false)
	suite.Config.EXPECT().GetString("scm_sha").Return("0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c")
	suite.Config.EXPECT().GetString("scm_branch").Return("master")
	suite.Config.EXPECT().GetString("scm_clone_url").Return("https://bitbucket.org/sparktree/gem_analogj_test.git")
	suite.Config.EXPECT().GetString("scm_repo_name").Return("gem_analogj_test")
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("sparktree/gem_analogj_test")

	//test
	testScm, err := scm.Create("bitbucket", suite.PipelineData, suite.Config, suite.Client)
	require.NoError(suite.T(), err)
	payload, perr := testScm.RetrievePayload()
	require.NoError(suite.T(), perr)

	//assert
	require.Equal(suite.T(), payload.Head.Sha, "0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c")
	require.Equal(suite.T(), payload.Head.Ref, "master")
	require.Equal(suite.T(), payload.Head.Repo.CloneUrl, "https://bitbucket.org/sparktree/gem_analogj_test.git")
	require.Equal(suite.T(), payload.Head.Repo.Name, "gem_analogj_test")
	require.Equal(suite.T(), payload.Head.Repo.FullName, "sparktree/gem_analogj_test")
	require.NotEmpty(suite.T(), payload, "payload must be set after source Init")
	require.False(suite.T(), suite.PipelineData.IsPullRequest)
}

func (suite *ScmBitbucketTestSuite) TestScmBitbucket_CheckoutPushPayload() {
	//setup
	suite.Config.EXPECT().IsSet("scm_bitbucket_username").Return(true)
	suite.Config.EXPECT().IsSet("scm_bitbucket_password").MinTimes(1).Return(true)
	suite.Config.EXPECT().GetString("scm_bitbucket_username").MinTimes(1).Return("")
	suite.Config.EXPECT().GetString("scm_bitbucket_password").MinTimes(1).Return("")
	// (so that git doesnt fail on placeholder token)
	suite.Config.EXPECT().IsSet("scm_git_parent_path").Return(false)
	suite.Config.EXPECT().IsSet("scm_pull_request").Return(false)
	suite.Config.EXPECT().GetString("scm_sha").Return("4aa9e889f0beddbc6248f8efa09cecf9a85435a5")
	suite.Config.EXPECT().GetString("scm_branch").Return("master")
	suite.Config.EXPECT().GetString("scm_clone_url").Return("https://bitbucket.org/sparktree/gem_analogj_test.git")
	suite.Config.EXPECT().GetString("scm_repo_name").Return("gem_analogj_test")
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("sparktree/gem_analogj_test")

	//test
	githubScm, err := scm.Create("bitbucket", suite.PipelineData, suite.Config, suite.Client)
	require.NoError(suite.T(), err)
	payload, perr := githubScm.RetrievePayload()
	require.NoError(suite.T(), perr)
	pperr := githubScm.CheckoutPushPayload(payload)
	require.NoError(suite.T(), pperr)

	//assert
	require.NotEmpty(suite.T(), suite.PipelineData.GitLocalPath, "should set checkout path")
	require.Equal(suite.T(), "master", suite.PipelineData.GitLocalBranch, "should set local branch correctly")
	require.NotNil(suite.T(), suite.PipelineData.GitHeadInfo)
}

func (suite *ScmBitbucketTestSuite) TestScmBitbucket_CheckoutPushPayload_WithInvalidPayload() {
	//setup
	suite.Config.EXPECT().IsSet("scm_github_access_token").Return(true) //used by the init function
	suite.Config.EXPECT().IsSet("scm_github_api_endpoint").Return(false)
	suite.Config.EXPECT().IsSet("scm_git_parent_path").Return(false)

	//test
	testScm, err := scm.Create("github", suite.PipelineData, suite.Config, suite.Client)
	require.NoError(suite.T(), err)
	payload := &scm.Payload{
		Head: new(pipeline.ScmCommitInfo),
	}
	pperr := testScm.CheckoutPushPayload(payload)

	//assert
	require.Error(suite.T(), pperr, "should return an error")
}

func (suite *ScmBitbucketTestSuite) TestScmBitbucket_CheckoutPullRequestPayload() {
	//setup

	suite.Config.EXPECT().IsSet("scm_bitbucket_username").Return(true)
	suite.Config.EXPECT().IsSet("scm_bitbucket_password").MinTimes(1).Return(true)
	suite.Config.EXPECT().GetString("scm_bitbucket_username").MinTimes(1).Return("")
	suite.Config.EXPECT().GetString("scm_bitbucket_password").MinTimes(1).Return("")
	suite.Config.EXPECT().IsSet("scm_git_parent_path").Return(false)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("sparktree/gem_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("scm_pull_request").Return("3")
	suite.Config.EXPECT().IsSet("scm_pull_request").Return(true)
	suite.Config.EXPECT().GetString("scm_notify_source").Return("CapsuleCD")
	suite.Config.EXPECT().GetString("scm_notify_target_url").Return("https://www.capsulecd.com")
	
	//test
	githubScm, err := scm.Create("bitbucket", suite.PipelineData, suite.Config, suite.Client)
	require.NoError(suite.T(), err)
	payload, perr := githubScm.RetrievePayload()
	require.NoError(suite.T(), perr)
	pperr := githubScm.CheckoutPullRequestPayload(payload)
	require.NoError(suite.T(), pperr)

	//assert
	require.NotEmpty(suite.T(), suite.PipelineData.GitLocalPath)
	require.NotEmpty(suite.T(), suite.PipelineData.GitLocalBranch)
	require.NotNil(suite.T(), suite.PipelineData.GitHeadInfo)
	require.NotNil(suite.T(), suite.PipelineData.GitBaseInfo)
}

func (suite *ScmBitbucketTestSuite) TestScmBitbucket_Publish() {
	//TODO:cant test publish because it'll continuously push to github repo.
	suite.T().Skip()

	//setup
	suite.Config.EXPECT().IsSet("scm_bitbucket_username").Return(true)
	suite.Config.EXPECT().IsSet("scm_bitbucket_password").MinTimes(1).Return(true)
	suite.Config.EXPECT().GetString("scm_bitbucket_username").MinTimes(1).Return("")
	suite.Config.EXPECT().GetString("scm_bitbucket_password").MinTimes(1).Return("")
	suite.Config.EXPECT().IsSet("scm_git_parent_path").Return(false)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("sparktree/gem_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("scm_pull_request").Return("4")
	suite.Config.EXPECT().IsSet("scm_pull_request").Return(true)

	//test
	testScm, err := scm.Create("bitbucket", suite.PipelineData, suite.Config, suite.Client)
	require.NoError(suite.T(), err)
	payload, perr := testScm.RetrievePayload()
	require.NoError(suite.T(), perr)
	pperr := testScm.CheckoutPullRequestPayload(payload)
	require.NoError(suite.T(), pperr)
	_, terr := utils.GitTag(suite.PipelineData.GitLocalPath, "v1.0.0", "test git tag message")
	require.NoError(suite.T(), terr)
	suite.PipelineData.ReleaseVersion = "1.0.0"
	pberr := testScm.Publish()
	require.NoError(suite.T(), pberr)

	//assert
	require.NotEmpty(suite.T(), suite.PipelineData.GitLocalPath)
	require.NotEmpty(suite.T(), suite.PipelineData.GitLocalBranch)
	require.NotNil(suite.T(), suite.PipelineData.GitHeadInfo)
	require.NotNil(suite.T(), suite.PipelineData.GitBaseInfo)
}

func (suite *ScmBitbucketTestSuite) TestScmBitbucket_PublishAssets() {
	//setup
	suite.Config.EXPECT().IsSet("scm_bitbucket_username").Return(true)
	suite.Config.EXPECT().IsSet("scm_bitbucket_password").MinTimes(1).Return(true)
	suite.Config.EXPECT().GetString("scm_bitbucket_username").MinTimes(1).Return(suite.Username)
	suite.Config.EXPECT().GetString("scm_bitbucket_password").MinTimes(1).Return(suite.Password)
	suite.Config.EXPECT().IsSet("scm_git_parent_path").Return(false)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("sparktree/gem_analogj_test").MinTimes(1)
	suite.PipelineData.ReleaseAssets = []pipeline.ScmReleaseAsset{
		{
			LocalPath:    path.Join("test_nested_dir", "gem_analogj_test-0.1.4.gem"),
			ArtifactName: "gem_analogj_test.gem",
		},
	}
	testScm, err := scm.Create("bitbucket", suite.PipelineData, suite.Config, suite.Client)
	require.NoError(suite.T(), err)
	defer os.Remove(suite.PipelineData.GitParentPath)

	suite.PipelineData.GitLocalPath = path.Join(suite.PipelineData.GitParentPath, "gem_analogj_test")

	cerr := utils.CopyDir(path.Join("testdata", "gem_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	//test
	paerr := testScm.PublishAssets(nil)
	require.NoError(suite.T(), paerr)
}

func (suite *ScmBitbucketTestSuite) TestScmBitbucket_Notify() {
	//setup
	suite.Config.EXPECT().IsSet("scm_bitbucket_username").Return(true)
	suite.Config.EXPECT().IsSet("scm_bitbucket_password").MinTimes(1).Return(true)
	suite.Config.EXPECT().GetString("scm_bitbucket_username").MinTimes(1).Return(suite.Username)
	suite.Config.EXPECT().GetString("scm_bitbucket_password").MinTimes(1).Return(suite.Password)
	suite.Config.EXPECT().IsSet("scm_git_parent_path").Return(false)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("sparktree/gem_analogj_test")
	suite.Config.EXPECT().GetString("scm_notify_source").Return("CapsuleCD")
	suite.Config.EXPECT().GetString("scm_notify_target_url").Return("https://www.capsulecd.com")

	//test
	githubScm, err := scm.Create("bitbucket", suite.PipelineData, suite.Config, suite.Client)
	require.NoError(suite.T(), err)
	pperr := githubScm.Notify("813875f454a9b18121ad1ee3dcb45e667189290b", "pending", "test message")
	require.NoError(suite.T(), pperr)
}
