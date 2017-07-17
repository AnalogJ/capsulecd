package pipeline

type PipelineData struct {
	IsPullRequest bool
	GitBaseInfo *PipelineScmCommitInfo
	GitHeadInfo *PipelineScmCommitInfo
	GitParentPath string
	GitLocalPath string
	GitLocalBranch string
	GitRemote string

	ReleaseVersion string
	ReleaseCommit string
	ReleaseArtifacts []string
}
