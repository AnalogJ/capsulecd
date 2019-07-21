package scm

import "github.com/analogj/capsulecd/pkg/pipeline"

type Payload struct {
	Head *pipeline.ScmCommitInfo
	Base *pipeline.ScmCommitInfo

	//Pull Request specific fields
	Title             string
	PullRequestNumber string
}
