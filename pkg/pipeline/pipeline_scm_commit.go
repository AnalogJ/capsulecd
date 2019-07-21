package pipeline

import "github.com/analogj/capsulecd/pkg/errors"

type ScmRepoInfo struct {
	CloneUrl string
	Name     string
	FullName string
}

type ScmCommitInfo struct {
	Sha  string //Commit Sha
	Ref  string //Commit Branch
	Repo *ScmRepoInfo
}

// TODO: validation almost needs to be source specific (or inherit from this base function), because source methods
// may require additional attributes, while these base payload keys are required for general step functions.
func (i *ScmCommitInfo) Validate() error {
	if i.Sha == "" {
		return errors.ScmPayloadFormatError("Incorrectly formatted payload, missing 'sha' key")
	} else if i.Ref == "" {
		return errors.ScmPayloadFormatError("Incorrectly formatted payload, missing 'Ref' key")
	} else if i.Repo == nil {
		return errors.ScmPayloadFormatError("Incorrectly formatted payload, missing 'Repo' key")
	} else if i.Repo.CloneUrl == "" {
		return errors.ScmPayloadFormatError("Incorrectly formatted payload, missing 'Repo.CloneUrl' key")
	} else if i.Repo.Name == "" {
		return errors.ScmPayloadFormatError("Incorrectly formatted payload, missing 'Repo.Name' key")
	} else if i.Repo.FullName == "" {
		return errors.ScmPayloadFormatError("Incorrectly formatted payload, missing 'Repo.FullName' key")
	} else {
		return nil
	}
}
