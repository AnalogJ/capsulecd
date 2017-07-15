package scm


type ScmRepoInfo struct {
	CloneUrl string
	Name string
	FullName string
}

type ScmCommitInfo struct {
	Sha string
	Ref string
	Repo *ScmRepoInfo
}

type ScmPayload struct {
	Title string
	Head *ScmCommitInfo
	Base *ScmCommitInfo
}