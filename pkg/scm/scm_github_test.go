package scm_test

import (
	"capsulecd/pkg/scm"
	"github.com/stretchr/testify/require"
	"testing"

	"capsulecd/pkg/config"
	"capsulecd/pkg/pipeline"
	"context"
	"crypto/tls"
	"github.com/seborama/govcr"
	"golang.org/x/oauth2"
	"net/http"
	"path"
	"io/ioutil"
	"os"
	"capsulecd/pkg/utils"
)

func vcrSetup(t *testing.T) *http.Client {
	accessToken := "PLACEHOLDER"
	if(os.Getenv("CI") != "true"){
		accessToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	}

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

	vcr := govcr.NewVCR(t.Name(),
		&govcr.VCRConfig{
			CassettePath: path.Join("testdata", "govcr-fixtures"),
			Client:       tc,
		})
	return vcr.Client
}


func TestScmGithub_Init_WithoutAccessToken(t *testing.T) {

	//setup
	testConfig, err := config.Create()
	require.NoError(t, err)
	testConfig.Set("scm", "github")
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	testScm, err := scm.Create("github", pipelineData, testConfig, client)

	//assert
	require.Nil(t, testScm)
	require.Error(t, err, "should raise an auth error")
}

func TestScmGithub_Init_WithGitParentPath(t *testing.T) {

	//setup
	testConfig, err := config.Create()
	require.NoError(t, err)
	testConfig.Set("scm", "github")
	testConfig.Set("scm_github_access_token", "placeholder")

	dirPath, err := ioutil.TempDir("", "")
	defer os.RemoveAll(dirPath)
	testConfig.Set("scm_git_parent_path", dirPath)
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	testScm, err := scm.Create("github", pipelineData, testConfig, client)
	require.Equal(t, pipelineData.GitParentPath, dirPath, "should correctly set parent path to existing")

	//assert
	require.NotNil(t, testScm)
	require.Nil(t, err, "should not have an error")
}

func TestScmGithub_Init_WithDefaults(t *testing.T) {

	//setup
	testConfig, err := config.Create()
	require.NoError(t, err)
	testConfig.Set("scm", "github")
	testConfig.Set("scm_github_access_token", "placeholder")
	pipelineData := new(pipeline.Data)

	//test
	testScm, err := scm.Create("github", pipelineData, testConfig, nil)
	require.NotEmpty(t, pipelineData.GitParentPath, "should correctly generate a temporary parent path")

	//assert
	require.NotNil(t, testScm)
	require.Nil(t, err, "should not have an error")
}


func TestScmGithub_RetrievePayload_PullRequest(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	testConfig.Set("scm", "github")
	testConfig.Set("scm_pull_request", "12")
	testConfig.Set("scm_repo_full_name", "AnalogJ/cookbook_analogj_test")
	testConfig.Set("scm_github_access_token", "placeholder")
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	githubScm, err := scm.Create("github", pipelineData, testConfig, client)
	require.NoError(t, err)
	payload, perr := githubScm.RetrievePayload()
	require.NoError(t, perr)

	//assert
	require.NotEmpty(t, payload, "payload must be set after source Init")
	require.True(t, pipelineData.IsPullRequest)
}

func TestScmGithub_RetrievePayload_PullRequest_InvalidState(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	testConfig.Set("scm", "github")
	testConfig.Set("scm_pull_request", "11")
	testConfig.Set("scm_repo_full_name", "AnalogJ/cookbook_analogj_test")
	testConfig.Set("scm_github_access_token", "placeholder")
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	githubScm, err := scm.Create("github", pipelineData, testConfig, client)
	require.NoError(t, err)
	payload, perr := githubScm.RetrievePayload()

	//assert
	require.Error(t, perr, "should return an error when PR is closed")
	require.Nil(t, payload)
}

func TestScmGithub_RetrievePayload_Push(t *testing.T) {

	//setup
	testConfig, err := config.Create()
	testConfig.Set("scm", "github")
	testConfig.Set("scm_sha", "0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c")
	testConfig.Set("scm_branch", "master")
	testConfig.Set("scm_clone_url", "https://github.com/analogj/capsulecd.git")
	testConfig.Set("scm_repo_name", "capsulecd")
	testConfig.Set("scm_repo_full_name", "AnalogJ/capsulecd")
	testConfig.Set("scm_github_access_token", "placeholder")
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	githubScm, err := scm.Create("github", pipelineData, testConfig, client)
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
	testConfig, err := config.Create()
	testConfig.Set("scm", "github")
	testConfig.Set("scm_sha", "0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c")
	testConfig.Set("scm_branch", "master")
	testConfig.Set("scm_clone_url", "https://github.com/analogj/capsulecd.git")
	testConfig.Set("scm_repo_name", "capsulecd")
	testConfig.Set("scm_repo_full_name", "AnalogJ/capsulecd")
	testConfig.Set("scm_github_access_token", "placeholder")
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	githubScm, err := scm.Create("github", pipelineData, testConfig, client)
	require.NoError(t, err)
	payload, perr := githubScm.RetrievePayload()
	require.NoError(t, perr)
	testConfig.Set("scm_github_access_token", "") //set the Access Token to empty string before doing checkout
	// (so that git doesnt fail on placeholder token)
	pperr := githubScm.CheckoutPushPayload(payload)
	require.NoError(t, pperr)

	//assert
	require.NotEmpty(t, pipelineData.GitLocalPath, "should set checkout path")
	require.Equal(t, "master",pipelineData.GitLocalBranch, "should set local branch correctly")
	require.NotNil(t, pipelineData.GitHeadInfo)
}

func TestScmGithub_ProcessPushPayload_WithInvalidPayload(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	testConfig.Set("scm", "github")
	testConfig.Set("scm_sha", "0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c")
	testConfig.Set("scm_branch", "master")
	testConfig.Set("scm_clone_url", "https://github.com/analogj/capsulecd.git")
	testConfig.Set("scm_repo_name", "capsulecd")
	testConfig.Set("scm_repo_full_name", "AnalogJ/capsulecd")
	testConfig.Set("scm_github_access_token", "placeholder")
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	githubScm, err := scm.Create("github", pipelineData, testConfig, client)
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
	testConfig, err := config.Create()
	require.NoError(t, err)
	testConfig.Set("scm", "github")
	testConfig.Set("scm_pull_request", "12")
	testConfig.Set("scm_repo_full_name", "AnalogJ/cookbook_analogj_test")
	testConfig.Set("scm_github_access_token", "placeholder")
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	githubScm, err := scm.Create("github", pipelineData, testConfig, client)
	require.NoError(t, err)
	payload, perr := githubScm.RetrievePayload()
	require.NoError(t, perr)
	testConfig.Set("scm_github_access_token", "") //set the Access Token to empty string before doing checkout
	// (so that git doesnt fail on placeholder token)
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
	testConfig, err := config.Create()
	require.NoError(t, err)
	testConfig.Set("scm", "github")
	testConfig.Set("scm_repo_full_name", "AnalogJ/gem_analogj_test")
	testConfig.Set("scm_github_access_token", "placeholder")
	pipelineData := new(pipeline.Data)
	pipelineData.ReleaseAssets = []pipeline.ScmReleaseAsset{
		{
			LocalPath: path.Join("test_nested_dir", "gem_analogj_test-0.1.4.gem"),
			ArtifactName: "gem_analogj_test.gem",
		},
	}
	client := vcrSetup(t)
	githubScm, err := scm.Create("github", pipelineData, testConfig, client)
	require.NoError(t, err)
	defer os.Remove(pipelineData.GitParentPath)

	pipelineData.GitLocalPath = path.Join(pipelineData.GitParentPath, "gem_analogj_test")

	cerr := utils.CopyDir(path.Join("testdata", "gem_analogj_test"), pipelineData.GitLocalPath )
	require.NoError(t, cerr)
	//test

	paerr := githubScm.PublishAssets(3663503)
	require.NoError(t, paerr)
}

func TestScmGithub_Notify(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	require.NoError(t, err)
	testConfig.Set("scm", "github")
	testConfig.Set("scm_repo_full_name", "AnalogJ/cookbook_analogj_test")
	testConfig.Set("scm_github_access_token", "placeholder")
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	githubScm, err := scm.Create("github", pipelineData, testConfig, client)
	require.NoError(t, err)
	pperr := githubScm.Notify("49f5bfbf4610f0c2a54d33945521051ba92b2eac","success", "test message")
	require.NoError(t, pperr)
}