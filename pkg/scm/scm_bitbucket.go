package scm

import (
	"capsulecd/pkg/pipeline"
	"context"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"net/http"
	"capsulecd/pkg/config"
)

type scmBitbucket struct {
	Config config.Interface
	Client       *github.Client
	PipelineData *pipeline.Data
}

// configure method will generate an authenticated client that can be used to comunicate with Github
// MUST set @git_parent_path
// MUST set @client field
func (b *scmBitbucket) init(pipelineData *pipeline.Data, config config.Interface, client *http.Client) error {
	b.PipelineData = pipelineData
	b.Config = config

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "... your access token ..."},
	)
	tc := oauth2.NewClient(ctx, ts)

	b.Client = github.NewClient(tc)
	return nil
}

func (b *scmBitbucket) RetrievePayload() (*Payload, error) {
	return new(Payload), nil
}

func (b *scmBitbucket) CheckoutPushPayload(payload *Payload) error {
	return nil
}

func (b *scmBitbucket) CheckoutPullRequestPayload(payload *Payload) error {
	return nil
}

func (b *scmBitbucket) Publish() error {
	return nil
}

func (b *scmBitbucket) Notify(ref string, state string, message string) error {
	return nil
}
