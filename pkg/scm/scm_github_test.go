package scm_test

import (
	"capsulecd/pkg/scm"
	"github.com/stretchr/testify/assert"
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
)

func vcrSetup(t *testing.T) *http.Client {

	ts := oauth2.StaticTokenSource(
		//setting a real access token here will allow API calls to connect successfully
		&oauth2.Token{AccessToken: "PLACEHOLDER"},
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


func TestScmGithub_init_WithoutAccessToken(t *testing.T) {

	//setup
	testConfig, err := config.Create()
	assert.NoError(t, err)
	testConfig.Set("scm", "github")
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	testScm, err := scm.Create("github", pipelineData, testConfig, client)

	//assert
	assert.Nil(t, testScm)
	assert.Error(t, err, "should raise an auth error")
}

func TestScmGithub_init_WithGitParentPath(t *testing.T) {

	//setup
	testConfig, err := config.Create()
	assert.NoError(t, err)
	testConfig.Set("scm", "github")
	testConfig.Set("scm_github_access_token", "placeholder")

	dirPath, err := ioutil.TempDir("", "")
	defer os.RemoveAll(dirPath)
	testConfig.Set("scm_git_parent_path", dirPath)
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	testScm, err := scm.Create("github", pipelineData, testConfig, client)
	assert.Equal(t, pipelineData.GitParentPath, dirPath, "should correctly set parent path to existing")

	//assert
	assert.NotNil(t, testScm)
	assert.Nil(t, err, "should not have an error")
}

func TestScmGithub_init_WithDefaults(t *testing.T) {

	//setup
	testConfig, err := config.Create()
	assert.NoError(t, err)
	testConfig.Set("scm", "github")
	testConfig.Set("scm_github_access_token", "placeholder")
	pipelineData := new(pipeline.Data)

	//test
	testScm, err := scm.Create("github", pipelineData, testConfig, nil)
	assert.NotEmpty(t, pipelineData.GitParentPath, "should correctly generate a temporary parent path")

	//assert
	assert.NotNil(t, testScm)
	assert.Nil(t, err, "should not have an error")
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
	assert.NoError(t, err)
	payload, perr := githubScm.RetrievePayload()
	assert.NoError(t, perr)

	//assert
	assert.NotEmpty(t, payload, "payload must be set after source Init")
	assert.True(t, pipelineData.IsPullRequest)
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
	assert.NoError(t, err)
	payload, perr := githubScm.RetrievePayload()

	//assert
	assert.Error(t, perr, "should return an error when PR is closed")
	assert.Nil(t, payload)
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
	assert.NoError(t, err)
	payload, perr := githubScm.RetrievePayload()
	assert.NoError(t, perr)

	//assert
	assert.Equal(t, payload.Head.Sha, "0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c")
	assert.Equal(t, payload.Head.Ref, "master")
	assert.Equal(t, payload.Head.Repo.CloneUrl, "https://github.com/analogj/capsulecd.git")
	assert.Equal(t, payload.Head.Repo.Name, "capsulecd")
	assert.Equal(t, payload.Head.Repo.FullName, "AnalogJ/capsulecd")
	assert.NotEmpty(t, payload, "payload must be set after source Init")
	assert.False(t, pipelineData.IsPullRequest)
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
	assert.NoError(t, err)
	payload, perr := githubScm.RetrievePayload()
	assert.NoError(t, perr)
	testConfig.Set("scm_github_access_token", "") //set the Access Token to empty string before doing checkout
	// (so that git doesnt fail on placeholder token)
	pperr := githubScm.CheckoutPushPayload(payload)
	assert.NoError(t, pperr)

	//assert
	assert.NotEmpty(t, pipelineData.GitLocalPath, "should set checkout path")
	assert.Equal(t, "master",pipelineData.GitLocalBranch, "should set local branch correctly")
	assert.NotNil(t, pipelineData.GitHeadInfo)
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
	assert.NoError(t, err)
	payload := &scm.Payload{
		Head: new(pipeline.ScmCommitInfo),
	}
	pperr := githubScm.CheckoutPushPayload(payload)

	//assert
	assert.Error(t, pperr, "should return an error")

}

func TestScmGithub_ProcessPullRequestPayload(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	assert.NoError(t, err)
	testConfig.Set("scm", "github")
	testConfig.Set("scm_pull_request", "12")
	testConfig.Set("scm_repo_full_name", "AnalogJ/cookbook_analogj_test")
	testConfig.Set("scm_github_access_token", "placeholder")
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	githubScm, err := scm.Create("github", pipelineData, testConfig, client)
	assert.NoError(t, err)
	payload, perr := githubScm.RetrievePayload()
	assert.NoError(t, perr)
	testConfig.Set("scm_github_access_token", "") //set the Access Token to empty string before doing checkout
	// (so that git doesnt fail on placeholder token)
	pperr := githubScm.CheckoutPullRequestPayload(payload)
	assert.NoError(t, pperr)

	//assert
	assert.NotEmpty(t, pipelineData.GitLocalPath)
	assert.NotEmpty(t, pipelineData.GitLocalBranch)
	assert.NotNil(t, pipelineData.GitHeadInfo)
	assert.NotNil(t, pipelineData.GitBaseInfo)
}

//cant test publish becasue it'll continuously push to github repo.
//func TestScmGithub_Publish(t *testing.T) {
//
//}

func TestScmGithub_Notify(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	assert.NoError(t, err)
	testConfig.Set("scm", "github")
	testConfig.Set("scm_repo_full_name", "AnalogJ/cookbook_analogj_test")
	testConfig.Set("scm_github_access_token", "placeholder")
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	githubScm, err := scm.Create("github", pipelineData, testConfig, client)
	assert.NoError(t, err)
	pperr := githubScm.Notify("49f5bfbf4610f0c2a54d33945521051ba92b2eac","success", "test message")
	assert.NoError(t, pperr)
}