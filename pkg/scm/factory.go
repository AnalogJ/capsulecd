package scm

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/errors"
	"capsulecd/pkg/pipeline"
	"fmt"
	"net/http"
)

func Create(scmType string, pipelineData *pipeline.Data, config config.Interface, client *http.Client) (Interface, error) {

	var scm Interface
	switch scmType {
	case "bitbucket":
		scm = new(scmBitbucket)
	case "github":
		scm = new(scmGithub)
	default:
		return nil, errors.ScmUnspecifiedError(fmt.Sprintf("Unknown Scm Type: %s", scmType))
	}

	if err := scm.Init(pipelineData, config, client); err != nil {
		return nil, err
	}
	return scm, nil
}
