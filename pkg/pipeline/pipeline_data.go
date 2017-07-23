package pipeline

type Data struct {
	IsPullRequest  bool
	GitBaseInfo    *ScmCommitInfo
	GitHeadInfo    *ScmCommitInfo
	GitParentPath  string
	GitLocalPath   string
	GitLocalBranch string
	GitRemote      string

	ReleaseVersion   string
	ReleaseCommit    string
	ReleaseAssets []ScmReleaseAsset
}
