package utils_test


import (
	"testing"
	"github.com/stretchr/testify/assert"
	"capsulecd/lib/utils"
	"io/ioutil"
)


func TestGitClone(t *testing.T) {

	dirPath, err := ioutil.TempDir("testdata","")
	assert.NoError(t, err)

	clonePath, cerr := utils.GitClone(dirPath, "test", "https://github.com/AnalogJ/test.git")
	assert.NoError(t, cerr)
	assert.NotEmpty(t, clonePath)

	//TODO add file cleanup after clone.
}

func TestGitFetch(t *testing.T) {

	dirPath, err := ioutil.TempDir("testdata","")
	assert.NoError(t, err)

	clonePath, cerr := utils.GitClone(dirPath, "cookbook_analogj_test", "https://github.com/AnalogJ/cookbook_analogj_test.git")
	assert.NoError(t, cerr)

	ferr := utils.GitFetch(clonePath, "refs/pull/12/merge", "localBranchName")
	assert.NoError(t, ferr)

	//TODO add file cleanup after clone.

}

func TestGitCheckout(t *testing.T) {

	dirPath, err := ioutil.TempDir("testdata","")
	assert.NoError(t, err)

	clonePath, cerr := utils.GitClone(dirPath, "npm_analogj_test", "https://github.com/AnalogJ/npm_analogj_test.git")
	assert.NoError(t, cerr)

	ferr := utils.GitCheckout(clonePath, "branch_test")
	assert.NoError(t, ferr)

	//TODO add file cleanup after clone.

}

func TestGitCommit(t *testing.T) {

	dirPath, err := ioutil.TempDir("testdata","")
	assert.NoError(t, err)

	clonePath, cerr := utils.GitClone(dirPath, "commit_to_npm_analogj_test", "https://github.com/AnalogJ/npm_analogj_test.git")
	assert.NoError(t, cerr)

	ferr := utils.GitCheckout(clonePath, "branch_test")
	assert.NoError(t, ferr)

	//create a new file
	d1 := []byte("hello\nworld\n")
	werr := ioutil.WriteFile(clonePath + "/commit_testfile.txt", d1, 0644)
	assert.NoError(t, werr)

	gcerr := utils.GitCommit(clonePath, "Added New File")
	assert.NoError(t, gcerr)

	//TODO add file cleanup after clone.

}


func TestGitTag(t *testing.T) {

	dirPath, err := ioutil.TempDir("testdata","")
	assert.NoError(t, err)

	clonePath, cerr := utils.GitClone(dirPath, "add_tag_npm_analogj_test", "https://github.com/AnalogJ/npm_analogj_test.git")
	assert.NoError(t, cerr)

	ferr := utils.GitCheckout(clonePath, "branch_test")
	assert.NoError(t, ferr)

	//create a new file
	d1 := []byte("hello\nworld\n")
	werr := ioutil.WriteFile(clonePath + "/tag_testfile.txt", d1, 0644)
	assert.NoError(t, werr)

	gcerr := utils.GitCommit(clonePath, "Added New File")
	assert.NoError(t, gcerr)

	tid, terr := utils.GitTag(clonePath, "v9.9.9")
	assert.NoError(t, terr)
	assert.NotEmpty(t, tid)

	//TODO add file cleanup after clone.
}

func TestGitPush(t *testing.T) {
	t.Skip() //Skipping because access_token not available during remote testing.
	dirPath, err := ioutil.TempDir("testdata","")
	assert.NoError(t, err)

	clonePath, cerr := utils.GitClone(dirPath, "push_npm_analogj_test", "https://access_token_here:@github.com/AnalogJ/npm_analogj_test.git")
	assert.NoError(t, cerr)

	ferr := utils.GitCheckout(clonePath, "branch_test")
	assert.NoError(t, ferr)

	//create a new file
	d1 := []byte("hello\nworld\n")
	werr := ioutil.WriteFile(clonePath + "/push_testfile.txt", d1, 0644)
	assert.NoError(t, werr)

	gcerr := utils.GitCommit(clonePath, "Added New File")
	assert.NoError(t, gcerr)

	perr := utils.GitPush(clonePath, "branch_test", "branch_test")
	assert.NoError(t, perr)

	//TODO add file cleanup after clone.

}

func TestGitPush_PullRequest(t *testing.T) {
	t.Skip() //Skipping because access_token not available during remote testing.
	dirPath, err := ioutil.TempDir("testdata","")
	assert.NoError(t, err)


	clonePath, cerr := utils.GitClone(dirPath, "cookbook_analogj_test", "https://access_token_here:@github.com/AnalogJ/cookbook_analogj_test.git")
	assert.NoError(t, cerr)

	ferr := utils.GitFetch(clonePath, "refs/pull/13/merge", "localBranchName")
	assert.NoError(t, ferr)

	//create a new file
	d1 := []byte("hello\nworld\n")
	werr := ioutil.WriteFile(clonePath + "/push_testfile.txt", d1, 0644)
	assert.NoError(t, werr)

	gcerr := utils.GitCommit(clonePath, "Added New File")
	assert.NoError(t, gcerr)

	perr := utils.GitPush(clonePath, "localBranchName", "master")
	assert.NoError(t, perr)

	//TODO add file cleanup after clone.

}


func TestGitLatestTaggedCommit(t *testing.T) {
	dirPath, err := ioutil.TempDir("testdata","")
	assert.NoError(t, err)


	clonePath, cerr := utils.GitClone(dirPath, "cookbook_analogj_test", "https://github.com/AnalogJ/cookbook_analogj_test.git")
	assert.NoError(t, cerr)

	tag, ferr := utils.GitLatestTaggedCommit(clonePath)
	assert.NoError(t, ferr)

	assert.Equal(t, "", tag)

	//TODO add file cleanup after clone.

}