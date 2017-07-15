package scm_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"capsulecd/lib/scm"

	"capsulecd/lib/config"
	"log"
)

func TestScmGithub(t *testing.T) {

	config.Init()
	config.Set("scm","github")

	githubScm := scm.Create()
	assert.Implements(t, (*scm.Scm)(nil), githubScm, "should implement the Scm interface")

}


func TestScmGithub_Configure(t *testing.T) {

	config.Init()
	config.Set("scm","github")
	config.Set("scm_github_access_token","github")

	githubScm := scm.Create()

	githubScm.Configure()
	assert.NotEmpty(t, githubScm.Options().GitParentPath, "GitParentPath must be set after source Configure")

}


func TestScmGithub_RetrievePayload_PullRequest(t *testing.T) {

	config.Init()
	config.Set("scm","github")
	config.Set("scm_pull_request","12")
	config.Set("scm_repo_full_name","AnalogJ/cookbook_analogj_test")
	config.Set("scm_github_access_token","7fb05afec5fafb011a28b8651e20aebde1f990cb")

	githubScm := scm.Create()

	githubScm.Configure()
	payload := githubScm.RetrievePayload()
	log.Print(payload)

	assert.NotEmpty(t, payload, "payload must be set after source Configure")

}
