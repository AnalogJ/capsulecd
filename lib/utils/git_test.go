package utils_test

import (
	"capsulecd/lib/utils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestGitClone(t *testing.T) {
	t.Parallel()
	dirPath, err := ioutil.TempDir("testdata", "")
	assert.NoError(t, err)
	defer deleteTestRepo(dirPath)

	clonePath, cerr := utils.GitClone(dirPath, "test", "https://github.com/AnalogJ/test.git")
	assert.NoError(t, cerr)
	assert.NotEmpty(t, clonePath)
}

func TestGitFetch(t *testing.T) {
	t.Parallel()

	dirPath, err := ioutil.TempDir("testdata", "")
	assert.NoError(t, err)
	defer deleteTestRepo(dirPath)

	clonePath, cerr := utils.GitClone(dirPath, "cookbook_analogj_test", "https://github.com/AnalogJ/cookbook_analogj_test.git")
	assert.NoError(t, cerr)

	ferr := utils.GitFetch(clonePath, "refs/pull/12/merge", "localBranchName")
	assert.NoError(t, ferr)
}

func TestGitCheckout(t *testing.T) {
	t.Parallel()

	dirPath, err := ioutil.TempDir("testdata", "")
	assert.NoError(t, err)
	defer deleteTestRepo(dirPath)

	clonePath, cerr := utils.GitClone(dirPath, "npm_analogj_test", "https://github.com/AnalogJ/npm_analogj_test.git")
	assert.NoError(t, cerr)

	ferr := utils.GitCheckout(clonePath, "branch_test")
	assert.NoError(t, ferr)
}

func TestGitCommit(t *testing.T) {
	t.Parallel()

	dirPath, err := ioutil.TempDir("testdata", "")
	assert.NoError(t, err)
	defer deleteTestRepo(dirPath)

	clonePath, cerr := utils.GitClone(dirPath, "commit_to_npm_analogj_test", "https://github.com/AnalogJ/npm_analogj_test.git")
	assert.NoError(t, cerr)

	ferr := utils.GitCheckout(clonePath, "branch_test")
	assert.NoError(t, ferr)

	//create a new file
	d1 := []byte("hello\nworld\n")
	werr := ioutil.WriteFile(clonePath+"/commit_testfile.txt", d1, 0644)
	assert.NoError(t, werr)

	gcerr := utils.GitCommit(clonePath, "Added New File")
	assert.NoError(t, gcerr)
}

func TestGitTag(t *testing.T) {
	t.Parallel()

	dirPath, err := ioutil.TempDir("testdata", "")
	assert.NoError(t, err)
	defer deleteTestRepo(dirPath)

	clonePath, cerr := utils.GitClone(dirPath, "add_tag_npm_analogj_test", "https://github.com/AnalogJ/npm_analogj_test.git")
	assert.NoError(t, cerr)

	ferr := utils.GitCheckout(clonePath, "branch_test")
	assert.NoError(t, ferr)

	//create a new file
	d1 := []byte("hello\nworld\n")
	werr := ioutil.WriteFile(clonePath+"/tag_testfile.txt", d1, 0644)
	assert.NoError(t, werr)

	gcerr := utils.GitCommit(clonePath, "Added New File")
	assert.NoError(t, gcerr)

	tid, terr := utils.GitTag(clonePath, "v9.9.9")
	assert.NoError(t, terr)
	assert.NotEmpty(t, tid)
}

func TestGitPush(t *testing.T) {
	t.Skip() //Skipping because access_token not available during remote testing.
	dirPath, err := ioutil.TempDir("testdata", "")
	assert.NoError(t, err)
	defer deleteTestRepo(dirPath)

	clonePath, cerr := utils.GitClone(dirPath, "push_npm_analogj_test", "https://access_token_here:@github.com/AnalogJ/npm_analogj_test.git")
	assert.NoError(t, cerr)

	ferr := utils.GitCheckout(clonePath, "branch_test")
	assert.NoError(t, ferr)

	//create a new file
	d1 := []byte("hello\nworld\n")
	werr := ioutil.WriteFile(clonePath+"/push_testfile.txt", d1, 0644)
	assert.NoError(t, werr)

	gcerr := utils.GitCommit(clonePath, "Added New File")
	assert.NoError(t, gcerr)

	perr := utils.GitPush(clonePath, "branch_test", "branch_test")
	assert.NoError(t, perr)

}

func TestGitPush_PullRequest(t *testing.T) {
	t.Skip() //Skipping because access_token not available during remote testing.
	dirPath, err := ioutil.TempDir("testdata", "")
	assert.NoError(t, err)
	defer deleteTestRepo(dirPath)

	clonePath, cerr := utils.GitClone(dirPath, "cookbook_analogj_test", "https://access_token_here:@github.com/AnalogJ/cookbook_analogj_test.git")
	assert.NoError(t, cerr)

	ferr := utils.GitFetch(clonePath, "refs/pull/13/merge", "localBranchName")
	assert.NoError(t, ferr)

	//create a new file
	d1 := []byte("hello\nworld\n")
	werr := ioutil.WriteFile(clonePath+"/push_testfile.txt", d1, 0644)
	assert.NoError(t, werr)

	gcerr := utils.GitCommit(clonePath, "Added New File")
	assert.NoError(t, gcerr)

	perr := utils.GitPush(clonePath, "localBranchName", "master")
	assert.NoError(t, perr)

}

func TestGitLatestTaggedCommit(t *testing.T) {
	t.Parallel()

	dirPath, err := ioutil.TempDir("testdata", "")
	assert.NoError(t, err)
	defer deleteTestRepo(dirPath)

	clonePath, cerr := utils.GitClone(dirPath, "cookbook_analogj_test", "https://github.com/AnalogJ/cookbook_analogj_test.git")
	assert.NoError(t, cerr)

	tag, ferr := utils.GitLatestTaggedCommit(clonePath)
	assert.NoError(t, ferr)

	assert.Equal(t, "v0.1.11", tag.TagShortName)

}

func TestGitGenerateChangelog(t *testing.T) {
	t.Parallel()

	dirPath, err := ioutil.TempDir("testdata", "")
	assert.NoError(t, err)
	defer deleteTestRepo(dirPath)

	clonePath, cerr := utils.GitClone(dirPath, "cookbook_analogj_test", "https://github.com/AnalogJ/cookbook_analogj_test.git")
	assert.NoError(t, cerr)

	changelog, ferr := utils.GitGenerateChangelog(clonePath, "43adaa328f74fd44abb33d33d8b149ab3780f209", "f3d573aacc59f2a6e2318dd140f3091c16b4b8fe", "")
	assert.NoError(t, ferr)

	assert.Equal(t,
		`Timestamp |  SHA | Message | Author
	------------- | ------------- | ------------- | -------------
	2017-07-16T01:41Z | f3d573aa | Added New File | CapsuleCD
	2017-07-16T01:26Z | 842436c9 | Merge 39e720e37a19716c098757cb5c78ea90b18111d3 into 97cae66b077de3798995342da781c270b1786820 | Jason Kulatunga
	2017-07-16T01:26Z | 39e720e3 | Update README.md | Jason Kulatunga
	2016-02-28T06:59Z | 97cae66b | (v0.1.11) Automated packaging of release by CapsuleCD | CapsuleCD
	2016-02-28T06:52Z | d4b8c3b5 | Update README.md | Jason Kulatunga
	2016-02-28T00:01Z | d0d3fc8f | (v0.1.10) Automated packaging of release by CapsuleCD | CapsuleCD
	`, changelog)
}

func deleteTestRepo(testRepoDirectory string) {
	os.RemoveAll(testRepoDirectory)
}
