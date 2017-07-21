package scm

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/errors"
	"capsulecd/pkg/pipeline"
	"fmt"
	"net/http"
)

func Create(scmType string, pipelineData *pipeline.Data, config config.Interface, client *http.Client) (Interface, error) {

	switch scmType {
	case "bitbucket":
		scm := new(scmBitbucket)
		if err := scm.init(pipelineData, config, client); err != nil {
			return nil, err
		}
		return scm, nil
	case "github":
		scm := new(scmGithub)
		if err := scm.init(pipelineData, config, client); err != nil {
			return nil, err
		}
		return scm, nil
	default:
		return nil, errors.ScmUnspecifiedError(fmt.Sprintf("Unknown Scm Type: %s", scmType))
	}
}
