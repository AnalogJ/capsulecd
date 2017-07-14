package capsulecd

import (
	"github.com/google/go-github/github"
	"context"
	"golang.org/x/oauth2"
)

type SourceGithub struct {
	client *github.Client
	gitParentPath string
}

// configure method will generate an authenticated client that can be used to comunicate with Github
// MUST set @git_parent_path
// MUST set @client field
func (sourceGithub SourceGithub) Configure() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "... your access token ..."},
	)
	tc := oauth2.NewClient(ctx, ts)

	sourceGithub.client = github.NewClient(tc)
	return
}

func (sourceGithub SourceGithub) ProcessPushPayload() {
	return
}

func (sourceGithub SourceGithub) ProcessPullRequestPayload() {
	return
}

func (sourceGithub SourceGithub) Publish() {
	return
}

func (sourceGithub SourceGithub) Notify() {
	return
}
