package scm

import (
	"capsulecd/lib/config"
	"net/http"
	"capsulecd/lib/errors"
	"fmt"
	"capsulecd/lib/pipeline"
)


type Scm interface {
	Init(pipelineData *pipeline.PipelineData, client *http.Client) error
	RetrievePayload() (*ScmPayload, error)
	ProcessPushPayload(payload *ScmPayload) error
	ProcessPullRequestPayload(payload *ScmPayload) error
	Publish() error //create release.
	Notify(ref string, state string, message string) error
}

func Create() (Scm, error) {

	switch scmType := config.Get("scm"); scmType {
	case "bitbucket":
		return new(scmBitbucket), nil
	case "github":
		return new(scmGithub), nil
	default:
		return nil, errors.ScmUnspecifiedError(fmt.Sprintf("Unknown Scm Type: %s", scmType))
	}
}