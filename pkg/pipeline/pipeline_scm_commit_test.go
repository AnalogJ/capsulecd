package pipeline_test

import (
	"capsulecd/pkg/pipeline"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestScmCommitInfo_Validate_Empty(t *testing.T) {
	//setup
	commit := pipeline.ScmCommitInfo{}

	//test
	err := commit.Validate()

	//assert
	require.Error(t, err, "Should raise error if anything/everything is missing")
}

func TestScmCommitInfo_Validate_MissingRef(t *testing.T) {
	//setup
	commit := pipeline.ScmCommitInfo{
		Sha: "1234",
		Repo: &pipeline.ScmRepoInfo{
			Name:     "reponame",
			CloneUrl: "clone_url",
		},
	}

	//test
	err := commit.Validate()

	//assert
	require.Error(t, err, "Should raise error if missing Ref")
}

func TestScmCommitInfo_Validate_MissingSha(t *testing.T) {
	//setup
	commit := pipeline.ScmCommitInfo{
		Ref: "1234",
		Repo: &pipeline.ScmRepoInfo{
			Name:     "reponame",
			CloneUrl: "clone_url",
		},
	}

	//test
	err := commit.Validate()

	//assert
	require.Error(t, err, "Should raise error if missing Sha")
}

func TestScmCommitInfo_Validate_MissingRepo(t *testing.T) {
	//setup
	commit := pipeline.ScmCommitInfo{
		Ref: "1234",
		Sha: "1234",
	}

	//test
	err := commit.Validate()

	//assert
	require.Error(t, err, "Should raise error if missing Repo Info")
}

func TestScmCommitInfo_Validate_MissingRepoCloneUrl(t *testing.T) {
	//setup
	commit := pipeline.ScmCommitInfo{
		Ref: "1234",
		Sha: "1234",
		Repo: &pipeline.ScmRepoInfo{
			Name: "reponame",
		},
	}

	//test
	err := commit.Validate()

	//assert
	require.Error(t, err, "Should raise error if missing Repo Clone Url")
}

func TestScmCommitInfo_Validate_MissingRepoName(t *testing.T) {
	//setup
	commit := pipeline.ScmCommitInfo{
		Ref: "1234",
		Sha: "1234",
		Repo: &pipeline.ScmRepoInfo{
			CloneUrl: "clone_url",
		},
	}

	//test
	err := commit.Validate()

	//assert
	require.Error(t, err, "Should raise error if missing Repo Name")
}

func TestScmCommitInfo_Validate(t *testing.T) {
	//setup
	commit := pipeline.ScmCommitInfo{
		Ref: "1234",
		Sha: "1234",
		Repo: &pipeline.ScmRepoInfo{
			Name:     "reponame",
			CloneUrl: "clone_url",
		},
	}

	//test
	err := commit.Validate()

	//assert
	require.NoError(t, err, "Should validate object successfully")
}
