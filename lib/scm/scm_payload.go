package scm

import "capsulecd/lib/pipeline"

type ScmPayload struct {
	Head *pipeline.PipelineScmCommitInfo
	Base *pipeline.PipelineScmCommitInfo

	//Pull Request specific fields
	Title string
	PullRequestNumber string
}
