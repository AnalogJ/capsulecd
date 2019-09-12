package pipeline

type Data struct {
	IsPullRequest  bool
	GitBaseInfo    *ScmCommitInfo
	GitHeadInfo    *ScmCommitInfo
	GitParentPath  string
	GitLocalPath   string
	GitLocalBranch string
	GitRemote      string
	GitNearestTag  *GitTagDetails

	ReleaseVersion string
	ReleaseCommit  string
	ReleaseAssets  []ScmReleaseAsset

	//Engine specific pipeline data
	GolangGoPath string
}
