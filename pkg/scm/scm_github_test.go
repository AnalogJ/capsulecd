package scm_test

import (
	"capsulecd/pkg/scm"
	"github.com/stretchr/testify/require"
	"testing"

	"capsulecd/pkg/config/mock"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/utils"
	"context"
	"crypto/tls"
	"github.com/golang/mock/gomock"
	"github.com/seborama/govcr"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

func vcrSetup(t *testing.T) *http.Client {
	accessToken := "PLACEHOLDER"
	//if(os.Getenv("CI") != "true"){
	//	accessToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	//}

	ts := oauth2.StaticTokenSource(
		//setting a real access token here will allow API calls to connect successfully
		&oauth2.Token{AccessToken: accessToken},
	)

	tr := http.DefaultTransport.(*http.Transport)
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true, //disable certificate validation because we're playing back http requests.
	}
	insecureClient := http.Client{
		Transport: tr,
	}

	ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, insecureClient)
	tc := oauth2.NewClient(ctx, ts)

	vcrConfig := govcr.VCRConfig{
		CassettePath: path.Join("testdata", "govcr-fixtures"),
		Client:       tc,
		ExcludeHeaderFunc: func(key string) bool {
			// HTTP headers are case-insensitive
			return strings.ToLower(key) == "user-agent"
		},
	}
	if(os.Getenv("CI") != "true"){
		vcrConfig.DisableRecording = true
	}

	vcr := govcr.NewVCR(t.Name(), &vcrConfig)
	return vcr.Client
}

func TestScmGithub_Init_WithoutAccessToken(t *testing.T) {

	//setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_config.NewMockInterface(mockCtrl)
	mockConfig.EXPECT().IsSet("scm_github_access_token").Return(false)

	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	testScm, err := scm.Create("github", pipelineData, mockConfig, client)

	//assert
	require.Nil(t, testScm)
	require.Error(t, err, "should raise an auth error")

}

func TestScmGithub_Init_WithGitParentPath(t *testing.T) {

	//setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_config.NewMockInterface(mockCtrl)
	mockConfig.EXPECT().IsSet("scm_github_access_token").Return(true)
	mockConfig.EXPECT().IsSet("scm_github_api_endpoint").Return(false)
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	dirPath, err := ioutil.TempDir("", "")
	defer os.RemoveAll(dirPath)
	mockConfig.EXPECT().IsSet("scm_git_parent_path").Return(true)
	mockConfig.EXPECT().GetString("scm_git_parent_path").Return(dirPath)

	//test
	testScm, err := scm.Create("github", pipelineData, mockConfig, client)
	require.Equal(t, pipelineData.GitParentPath, dirPath, "should correctly set parent path to existing")

	//assert
	require.NotNil(t, testScm)
	require.Nil(t, err, "should not have an error")

}

func TestScmGithub_Init_WithDefaults(t *testing.T) {

	//setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_config.NewMockInterface(mockCtrl)
	mockConfig.EXPECT().IsSet("scm_github_access_token").Return(true)
	mockConfig.EXPECT().IsSet("scm_github_api_endpoint").Return(false)
	mockConfig.EXPECT().GetString("scm_github_access_token").Return("placeholder")
	mockConfig.EXPECT().IsSet("scm_git_parent_path").Return(false)
	pipelineData := new(pipeline.Data)

	//test
	testScm, err := scm.Create("github", pipelineData, mockConfig, nil)
	require.NotEmpty(t, pipelineData.GitParentPath, "should correctly generate a temporary parent path")

	//assert
	require.NotNil(t, testScm)
	require.Nil(t, err, "should not have an error")

}

func TestScmGithub_RetrievePayload_PullRequest(t *testing.T) {
	//setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_config.NewMockInterface(mockCtrl)
	mockConfig.EXPECT().IsSet("scm_github_access_token").Return(true)
	mockConfig.EXPECT().IsSet("scm_github_api_endpoint").Return(false)
	mockConfig.EXPECT().IsSet("scm_git_parent_path").Return(false)
	mockConfig.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/cookbook_analogj_test")
	mockConfig.EXPECT().GetInt("scm_pull_request").Return(12)
	mockConfig.EXPECT().IsSet("scm_pull_request").Return(true)
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	githubScm, err := scm.Create("github", pipelineData, mockConfig, client)
	require.NoError(t, err)
	payload, perr := githubScm.RetrievePayload()
	require.NoError(t, perr)

	//assert
	require.NotEmpty(t, payload, "payload must be set after source Init")
	require.True(t, pipelineData.IsPullRequest)

}

func TestScmGithub_RetrievePayload_PullRequest_InvalidState(t *testing.T) {
	//setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_config.NewMockInterface(mockCtrl)
	mockConfig.EXPECT().IsSet("scm_github_access_token").Return(true)
	mockConfig.EXPECT().IsSet("scm_github_api_endpoint").Return(false)
	mockConfig.EXPECT().IsSet("scm_git_parent_path").Return(false)
	mockConfig.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/cookbook_analogj_test")
	mockConfig.EXPECT().GetInt("scm_pull_request").Return(11)
	mockConfig.EXPECT().IsSet("scm_pull_request").Return(true)
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	githubScm, err := scm.Create("github", pipelineData, mockConfig, client)
	require.NoError(t, err)
	payload, perr := githubScm.RetrievePayload()

	//assert
	require.Error(t, perr, "should return an error when PR is closed")
	require.Nil(t, payload)

}

func TestScmGithub_RetrievePayload_Push(t *testing.T) {

	//setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_config.NewMockInterface(mockCtrl)
	mockConfig.EXPECT().IsSet("scm_github_access_token").Return(true)
	mockConfig.EXPECT().IsSet("scm_github_api_endpoint").Return(false)
	mockConfig.EXPECT().IsSet("scm_git_parent_path").Return(false)
	mockConfig.EXPECT().IsSet("scm_pull_request").Return(false)
	mockConfig.EXPECT().GetString("scm_sha").Return("0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c")
	mockConfig.EXPECT().GetString("scm_branch").Return("master")
	mockConfig.EXPECT().GetString("scm_clone_url").Return("https://github.com/analogj/capsulecd.git")
	mockConfig.EXPECT().GetString("scm_repo_name").Return("capsulecd")
	mockConfig.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/capsulecd")
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	githubScm, err := scm.Create("github", pipelineData, mockConfig, client)
	require.NoError(t, err)
	payload, perr := githubScm.RetrievePayload()
	require.NoError(t, perr)

	//assert
	require.Equal(t, payload.Head.Sha, "0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c")
	require.Equal(t, payload.Head.Ref, "master")
	require.Equal(t, payload.Head.Repo.CloneUrl, "https://github.com/analogj/capsulecd.git")
	require.Equal(t, payload.Head.Repo.Name, "capsulecd")
	require.Equal(t, payload.Head.Repo.FullName, "AnalogJ/capsulecd")
	require.NotEmpty(t, payload, "payload must be set after source Init")
	require.False(t, pipelineData.IsPullRequest)
}

func TestScmGithub_ProcessPushPayload(t *testing.T) {
	//setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_config.NewMockInterface(mockCtrl)
	mockConfig.EXPECT().IsSet("scm_github_access_token").Return(true) //used by the init function
	mockConfig.EXPECT().IsSet("scm_github_api_endpoint").Return(false)
	mockConfig.EXPECT().GetString("scm_github_access_token").Return("") //set the Access Token to empty string before doing checkout
	// (so that git doesnt fail on placeholder token)
	mockConfig.EXPECT().IsSet("scm_git_parent_path").Return(false)
	mockConfig.EXPECT().IsSet("scm_pull_request").Return(false)
	mockConfig.EXPECT().GetString("scm_sha").Return("0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c")
	mockConfig.EXPECT().GetString("scm_branch").Return("master")
	mockConfig.EXPECT().GetString("scm_clone_url").Return("https://github.com/analogj/capsulecd.git")
	mockConfig.EXPECT().GetString("scm_repo_name").Return("capsulecd")
	mockConfig.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/capsulecd")
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	githubScm, err := scm.Create("github", pipelineData, mockConfig, client)
	require.NoError(t, err)
	payload, perr := githubScm.RetrievePayload()
	require.NoError(t, perr)
	pperr := githubScm.CheckoutPushPayload(payload)
	require.NoError(t, pperr)

	//assert
	require.NotEmpty(t, pipelineData.GitLocalPath, "should set checkout path")
	require.Equal(t, "master", pipelineData.GitLocalBranch, "should set local branch correctly")
	require.NotNil(t, pipelineData.GitHeadInfo)
}

func TestScmGithub_ProcessPushPayload_WithInvalidPayload(t *testing.T) {
	//setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_config.NewMockInterface(mockCtrl)
	mockConfig.EXPECT().IsSet("scm_github_access_token").Return(true) //used by the init function
	mockConfig.EXPECT().IsSet("scm_github_api_endpoint").Return(false)
	mockConfig.EXPECT().IsSet("scm_git_parent_path").Return(false)
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	githubScm, err := scm.Create("github", pipelineData, mockConfig, client)
	require.NoError(t, err)
	payload := &scm.Payload{
		Head: new(pipeline.ScmCommitInfo),
	}
	pperr := githubScm.CheckoutPushPayload(payload)

	//assert
	require.Error(t, pperr, "should return an error")

}

func TestScmGithub_ProcessPullRequestPayload(t *testing.T) {
	//setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_config.NewMockInterface(mockCtrl)
	mockConfig.EXPECT().IsSet("scm_github_access_token").Return(true)
	mockConfig.EXPECT().IsSet("scm_github_api_endpoint").Return(false)
	mockConfig.EXPECT().GetString("scm_github_access_token").Return("")
	mockConfig.EXPECT().IsSet("scm_git_parent_path").Return(false)
	mockConfig.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/cookbook_analogj_test").MinTimes(1)
	mockConfig.EXPECT().GetInt("scm_pull_request").Return(12)
	mockConfig.EXPECT().IsSet("scm_pull_request").Return(true)
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	githubScm, err := scm.Create("github", pipelineData, mockConfig, client)
	require.NoError(t, err)
	payload, perr := githubScm.RetrievePayload()
	require.NoError(t, perr)
	pperr := githubScm.CheckoutPullRequestPayload(payload)
	require.NoError(t, pperr)

	//assert
	require.NotEmpty(t, pipelineData.GitLocalPath)
	require.NotEmpty(t, pipelineData.GitLocalBranch)
	require.NotNil(t, pipelineData.GitHeadInfo)
	require.NotNil(t, pipelineData.GitBaseInfo)
}

//cant test publish because it'll continuously push to github repo.
//func TestScmGithub_Publish(t *testing.T) {
//
//}

func TestScmGithub_PublishAssets(t *testing.T) {
	//setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_config.NewMockInterface(mockCtrl)
	mockConfig.EXPECT().IsSet("scm_github_access_token").Return(true)
	mockConfig.EXPECT().IsSet("scm_github_api_endpoint").Return(false)
	mockConfig.EXPECT().IsSet("scm_git_parent_path").Return(false)
	mockConfig.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/gem_analogj_test").MinTimes(1)
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)
	pipelineData.ReleaseAssets = []pipeline.ScmReleaseAsset{
		{
			LocalPath:    path.Join("test_nested_dir", "gem_analogj_test-0.1.4.gem"),
			ArtifactName: "gem_analogj_test.gem",
		},
	}
	githubScm, err := scm.Create("github", pipelineData, mockConfig, client)
	require.NoError(t, err)
	defer os.Remove(pipelineData.GitParentPath)

	pipelineData.GitLocalPath = path.Join(pipelineData.GitParentPath, "gem_analogj_test")

	cerr := utils.CopyDir(path.Join("testdata", "gem_analogj_test"), pipelineData.GitLocalPath)
	require.NoError(t, cerr)
	//test

	paerr := githubScm.PublishAssets(3663503)
	require.NoError(t, paerr)
}

func TestScmGithub_Cleanup_WithoutEnablingBranchCleanup(t *testing.T) {
	//setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_config.NewMockInterface(mockCtrl)
	mockConfig.EXPECT().IsSet("scm_github_access_token").Return(true)
	mockConfig.EXPECT().IsSet("scm_github_api_endpoint").Return(false)
	mockConfig.EXPECT().IsSet("scm_git_parent_path").Return(false)
	mockConfig.EXPECT().GetBool("scm_enable_branch_cleanup").Return(false)
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)
	githubScm, err := scm.Create("github", pipelineData, mockConfig, client)
	require.NoError(t, err)
	defer os.Remove(pipelineData.GitParentPath)
	//test

	paerr := githubScm.Cleanup()

	//
	require.Error(t, paerr, "should raise an error")
}

func TestScmGithub_Cleanup_WithDifferentOrgs(t *testing.T) {
	//setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_config.NewMockInterface(mockCtrl)
	mockConfig.EXPECT().IsSet("scm_github_access_token").Return(true)
	mockConfig.EXPECT().IsSet("scm_github_api_endpoint").Return(false)
	mockConfig.EXPECT().IsSet("scm_git_parent_path").Return(false)
	mockConfig.EXPECT().GetBool("scm_enable_branch_cleanup").Return(true)
	pipelineData := new(pipeline.Data)
	pipelineData.GitHeadInfo = &pipeline.ScmCommitInfo{
		Ref: "AnalogJ-patch-6",
		Repo: &pipeline.ScmRepoInfo{
			CloneUrl: "https://github.com/AnalogJ/gem_analogj_test.git",
			Name:     "gem_analogj_test",
			FullName: "AnalogJ/gem_analogj_test",
		},
		Sha: "12345",
	}
	pipelineData.GitBaseInfo = &pipeline.ScmCommitInfo{
		Ref: "master",
		Repo: &pipeline.ScmRepoInfo{
			CloneUrl: "https://github.com/AnalogJ/gem_analogj_test.git",
			Name:     "gem_analogj_test",
			FullName: "DifferentOrg/gem_analogj_test",
		},
		Sha: "12345",
	}
	client := vcrSetup(t)
	githubScm, err := scm.Create("github", pipelineData, mockConfig, client)
	require.NoError(t, err)
	defer os.Remove(pipelineData.GitParentPath)
	//test

	paerr := githubScm.Cleanup()

	//
	require.Error(t, paerr, "should raise an error")
}

func TestScmGithub_Cleanup_WithHeadBranchMaster(t *testing.T) {
	//setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_config.NewMockInterface(mockCtrl)
	mockConfig.EXPECT().IsSet("scm_github_access_token").Return(true)
	mockConfig.EXPECT().IsSet("scm_github_api_endpoint").Return(false)
	mockConfig.EXPECT().IsSet("scm_git_parent_path").Return(false)
	mockConfig.EXPECT().GetBool("scm_enable_branch_cleanup").Return(true)
	pipelineData := new(pipeline.Data)
	pipelineData.GitHeadInfo = &pipeline.ScmCommitInfo{
		Ref: "master",
		Repo: &pipeline.ScmRepoInfo{
			CloneUrl: "https://github.com/AnalogJ/gem_analogj_test.git",
			Name:     "gem_analogj_test",
			FullName: "AnalogJ/gem_analogj_test",
		},
		Sha: "12345",
	}
	pipelineData.GitBaseInfo = &pipeline.ScmCommitInfo{
		Ref: "master",
		Repo: &pipeline.ScmRepoInfo{
			CloneUrl: "https://github.com/AnalogJ/gem_analogj_test.git",
			Name:     "gem_analogj_test",
			FullName: "AnalogJ/gem_analogj_test",
		},
		Sha: "12345",
	}
	client := vcrSetup(t)
	githubScm, err := scm.Create("github", pipelineData, mockConfig, client)
	require.NoError(t, err)
	defer os.Remove(pipelineData.GitParentPath)
	//test

	paerr := githubScm.Cleanup()

	//
	require.Error(t, paerr, "should raise an error")
}

//func TestScmGithub_Cleanup(t *testing.T) {
//	//setup
//	mockCtrl := gomock.NewController(t)
//	defer mockCtrl.Finish()
//	mockConfig := mock_config.NewMockInterface(mockCtrl)
//	mockConfig.EXPECT().IsSet("scm_github_access_token").Return(true)
//	mockConfig.EXPECT().IsSet("scm_git_parent_path").Return(false)
//	mockConfig.EXPECT().GetBool("scm_enable_branch_cleanup").Return(true)
//	pipelineData := new(pipeline.Data)
//	pipelineData.IsPullRequest = true
//	pipelineData.GitHeadInfo = &pipeline.ScmCommitInfo{
//		Ref: "AnalogJ-patch-3",
//		Repo: &pipeline.ScmRepoInfo{
//			CloneUrl: "https://github.com/AnalogJ/gem_analogj_test.git",
//			Name:     "gem_analogj_test",
//			FullName: "AnalogJ/gem_analogj_test",
//		},
//		Sha: "12345",
//	}
//	pipelineData.GitBaseInfo = &pipeline.ScmCommitInfo{
//		Ref: "master",
//		Repo: &pipeline.ScmRepoInfo{
//			CloneUrl: "https://github.com/AnalogJ/gem_analogj_test.git",
//			Name:     "gem_analogj_test",
//			FullName: "AnalogJ/gem_analogj_test",
//		},
//		Sha: "12345",
//	}
//	client := vcrSetup(t)
//	githubScm, err := scm.Create("github", pipelineData, mockConfig, client)
//	require.NoError(t, err)
//	defer os.Remove(pipelineData.GitParentPath)
//	//test
//
//	paerr := githubScm.Cleanup()
//
//	//
//	require.NoError(t, paerr, "should finish successfully")
//}

func TestScmGithub_Notify(t *testing.T) {
	//setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_config.NewMockInterface(mockCtrl)
	mockConfig.EXPECT().IsSet("scm_github_access_token").Return(true)
	mockConfig.EXPECT().IsSet("scm_github_api_endpoint").Return(false)
	mockConfig.EXPECT().IsSet("scm_git_parent_path").Return(false)
	mockConfig.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/cookbook_analogj_test")
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	githubScm, err := scm.Create("github", pipelineData, mockConfig, client)
	require.NoError(t, err)
	pperr := githubScm.Notify("49f5bfbf4610f0c2a54d33945521051ba92b2eac", "success", "test message")
	require.NoError(t, pperr)
}
