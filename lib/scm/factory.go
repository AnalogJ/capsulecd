package scm

import (
	"capsulecd/lib/config"
	"net/http"
	"capsulecd/lib/errors"
	"fmt"
)


type Scm interface {
	Options() *ScmOptions
	Configure(client *http.Client) error
	RetrievePayload() (*ScmPayload, error)
	ProcessPushPayload(payload *ScmPayload) error
	ProcessPullRequestPayload(payload *ScmPayload) error
	Publish() error //create release.
	Notify() error
}

func Create() (Scm, error) {

	switch scmType := config.Get("scm"); scmType {
	case "bitbucket":
		return &scmBitbucket{}, nil
	case "github":
		return &scmGithub{}, nil
	default:
		return nil, errors.ScmUnspecifiedError(fmt.Sprintf("Unknown Scm Type: %s", scmType))
	}
}