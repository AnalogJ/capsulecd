package scm

import "capsulecd/lib/pipeline"

type Payload struct {
	Head *pipeline.ScmCommitInfo
	Base *pipeline.ScmCommitInfo

	//Pull Request specific fields
	Title string
	PullRequestNumber string
}
