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
	"os"
	"path"
)

func vcrSetup(t *testing.T) *http.Client {

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")},
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

func TestScmGithub_Create_WithNoAuthToken(t *testing.T) {

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
	assert.Error(t, err)
	assert.NotEmpty(t, pipelineData.GitParentPath, "GitParentPath must be set after source Init")

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
	testConfig.Set("scm_github_access_token", "")
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	githubScm, err := scm.Create("github", pipelineData, testConfig, client)
	assert.NoError(t, err)
	payload, perr := githubScm.RetrievePayload()
	assert.NoError(t, perr)
	pperr := githubScm.ProcessPushPayload(payload)
	assert.NoError(t, pperr)

	//assert
	assert.NotEmpty(t, pipelineData.GitLocalPath)
	assert.NotEmpty(t, pipelineData.GitLocalBranch)
	assert.NotNil(t, pipelineData.GitHeadInfo)
}

func TestScmGithub_ProcessPullRequestPayload(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	testConfig.Set("scm", "github")
	testConfig.Set("scm_pull_request", "12")
	testConfig.Set("scm_repo_full_name", "AnalogJ/cookbook_analogj_test")
	testConfig.Set("scm_github_access_token", "")
	pipelineData := new(pipeline.Data)
	client := vcrSetup(t)

	//test
	githubScm, err := scm.Create("github", pipelineData, testConfig, client)
	assert.NoError(t, err)
	payload, perr := githubScm.RetrievePayload()
	assert.NoError(t, perr)
	pperr := githubScm.ProcessPullRequestPayload(payload)
	assert.NoError(t, pperr)

	//assert
	assert.NotEmpty(t, pipelineData.GitLocalPath)
	assert.NotEmpty(t, pipelineData.GitLocalBranch)
	assert.NotNil(t, pipelineData.GitHeadInfo)
	assert.NotNil(t, pipelineData.GitBaseInfo)
}
