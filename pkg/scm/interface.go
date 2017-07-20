package scm

import(
	"capsulecd/pkg/pipeline"
	"net/http"
	"capsulecd/pkg/config"
)

type Interface interface {
	init(pipelineData *pipeline.Data, config config.Interface, client *http.Client) error
	RetrievePayload() (*Payload, error)
	ProcessPushPayload(payload *Payload) error
	ProcessPullRequestPayload(payload *Payload) error
	Publish() error //create release.
	Notify(ref string, state string, message string) error
}