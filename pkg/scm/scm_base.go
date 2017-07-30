package scm

import (
	"net/url"
	"capsulecd/pkg/pipeline"
)

type scmBase struct {
	PipelineData *pipeline.Data
}


func (b *scmBase)authGitRemote(cloneUrl string, username string, password string) (string, error) {
	if username != "" || password != "" {
		// set the remote url, with embedded token
		u, err := url.Parse(cloneUrl)
		if err != nil {
			return "", err
		}
		u.User = url.UserPassword(username, password)
		return u.String(), nil
	} else {
		return cloneUrl, nil
	}
}