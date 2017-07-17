package scm_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"capsulecd/lib/scm"

	"capsulecd/lib/config"
	"log"
	"github.com/seborama/govcr"
	"path"
	"net/http"
	"golang.org/x/oauth2"
	"context"
	"crypto/tls"
	"os"
	"capsulecd/lib/pipeline"
)


func vcrSetup(t *testing.T) *http.Client {

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")},
	)

	tr := http.DefaultTransport.(*http.Transport)
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true, //disable certificate validation
	}
	insecureClient := http.Client{
		Transport: tr,
	}

	ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, insecureClient)
	tc := oauth2.NewClient(ctx, ts)

	vcr := govcr.NewVCR(t.Name(),
		&govcr.VCRConfig{
			CassettePath: path.Join("testdata", "govcr-fixtures"),
			Client: tc,
		})
	return vcr.Client
}

func TestScmGithub(t *testing.T) {

	config.Init()
	config.Set("scm","github")

	githubScm, err := scm.Create()
	assert.NoError(t, err)
	assert.Implements(t, (*scm.Scm)(nil), githubScm, "should implement the Scm interface")

}


func TestScmGithub_Init(t *testing.T) {

	config.Init()
	config.Set("scm","github")
	config.Set("scm_github_access_token","github")

	pipelineData := new(pipeline.Data)
	githubScm, err := scm.Create()
	assert.NoError(t, err)
	githubScm.Init(pipelineData, nil)
	assert.NotEmpty(t, pipelineData.GitParentPath, "GitParentPath must be set after source Init")

}

func TestScmGithub_Configure_WithNoAuthToken(t *testing.T) {
	config.Init()
	config.Set("scm","github")

	githubScm, err := scm.Create()
	assert.NoError(t, err)

	cerr := githubScm.Init(new(pipeline.Data), nil)
	assert.Error(t, cerr)
}


func TestScmGithub_RetrievePayload_PullRequest(t *testing.T) {

	config.Init()
	config.Set("scm","github")
	config.Set("scm_pull_request","12")
	config.Set("scm_repo_full_name","AnalogJ/cookbook_analogj_test")
	config.Set("scm_github_access_token", "placeholder")

	githubScm, err := scm.Create()
	assert.NoError(t, err)

	client := vcrSetup(t)

	pipelineData := new(pipeline.Data)
	githubScm.Init(pipelineData, client)
	payload, perr := githubScm.RetrievePayload()
	assert.NoError(t, perr)

	log.Print(payload)

	assert.NotEmpty(t, payload, "payload must be set after source Init")

	assert.True(t, pipelineData.IsPullRequest)
}

func TestScmGithub_RetrievePayload_Push(t *testing.T) {

	config.Init()
	config.Set("scm","github")
	config.Set("scm_sha","0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c")
	config.Set("scm_branch","master")
	config.Set("scm_clone_url","https://github.com/analogj/capsulecd.git")
	config.Set("scm_repo_name","capsulecd")
	config.Set("scm_repo_full_name","AnalogJ/capsulecd")
	config.Set("scm_github_access_token", "placeholder")

	githubScm, err := scm.Create()
	assert.NoError(t, err)

	client := vcrSetup(t)

	pipelineData := new(pipeline.Data)

	githubScm.Init(pipelineData, client)
	payload, perr := githubScm.RetrievePayload()
	assert.NoError(t, perr)

	assert.Equal(t, payload.Head.Sha, "0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c")
	assert.Equal(t, payload.Head.Ref, "master")
	assert.Equal(t, payload.Head.Repo.CloneUrl, "https://github.com/analogj/capsulecd.git")
	assert.Equal(t, payload.Head.Repo.Name, "capsulecd")
	assert.Equal(t, payload.Head.Repo.FullName, "AnalogJ/capsulecd")

	assert.NotEmpty(t, payload, "payload must be set after source Init")

	assert.False(t, pipelineData.IsPullRequest)
}

func TestScmGithub_ProcessPushPayload(t *testing.T) {
	config.Init()
	config.Set("scm","github")
	config.Set("scm_sha","0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c")
	config.Set("scm_branch","master")
	config.Set("scm_clone_url","https://github.com/analogj/capsulecd.git")
	config.Set("scm_repo_name","capsulecd")
	config.Set("scm_repo_full_name","AnalogJ/capsulecd")
	config.Set("scm_github_access_token", "")

	githubScm, err := scm.Create()
	assert.NoError(t, err)

	client := vcrSetup(t)

	pipelineData := new(pipeline.Data)

	githubScm.Init(pipelineData, client)
	payload, perr := githubScm.RetrievePayload()
	assert.NoError(t, perr)

	pperr := githubScm.ProcessPushPayload(payload)
	assert.NoError(t, pperr)

	assert.NotEmpty(t, pipelineData.GitLocalPath)
	assert.NotEmpty(t, pipelineData.GitLocalBranch)
	assert.NotNil(t, pipelineData.GitHeadInfo)
}


func TestScmGithub_ProcessPullRequestPayload(t *testing.T) {

	config.Init()
	config.Set("scm","github")
	config.Set("scm_pull_request","12")
	config.Set("scm_repo_full_name","AnalogJ/cookbook_analogj_test")
	config.Set("scm_github_access_token", "")

	githubScm, err := scm.Create()
	assert.NoError(t, err)

	client := vcrSetup(t)
	pipelineData := new(pipeline.Data)

	githubScm.Init(pipelineData, client)
	payload, perr := githubScm.RetrievePayload()
	assert.NoError(t, perr)

	pperr := githubScm.ProcessPullRequestPayload(payload)
	assert.NoError(t, pperr)

	assert.NotEmpty(t, pipelineData.GitLocalPath)
	assert.NotEmpty(t, pipelineData.GitLocalBranch)
	assert.NotNil(t, pipelineData.GitHeadInfo)
	assert.NotNil(t, pipelineData.GitBaseInfo)
}