package scm
import "capsulecd/lib/errors"

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

// TODO: validation almost needs to be source specific (or inherit from this base function), because source methods
// may require additional attributes, while these base payload keys are required for general step functions.
func (i *ScmCommitInfo) Validate() error {
	if i.Sha == "" {
		return errors.ScmPayloadFormatError("Incorrectly formatted payload, missing 'sha' key")
	} else if(i.Ref == ""){
		return errors.ScmPayloadFormatError("Incorrectly formatted payload, missing 'Ref' key")
	} else if(i.Repo.CloneUrl == ""){
		return errors.ScmPayloadFormatError("Incorrectly formatted payload, missing 'Repo.CloneUrl' key")
	} else if(i.Repo.Name == ""){
		return errors.ScmPayloadFormatError("Incorrectly formatted payload, missing 'Repo.Name' key")
	} else {
		return nil
	}
}