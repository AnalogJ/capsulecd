package utils_test

import (
	"capsulecd/pkg/utils"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestGitClone(t *testing.T) {
	t.Parallel()

	//setup
	dirPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer deleteTestRepo(dirPath)

	//test
	clonePath, cerr := utils.GitClone(dirPath, "test", "https://github.com/AnalogJ/test.git")

	//assert
	require.NoError(t, cerr)
	require.NotEmpty(t, clonePath)
}

func TestGitClone_ExistingPath(t *testing.T) {
	t.Parallel()

	//setup
	dirPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	merr := os.MkdirAll(path.Join(dirPath, "test"), os.ModePerm)
	require.NoError(t, merr)
	defer deleteTestRepo(dirPath)

	//test
	clonePath, cerr := utils.GitClone(dirPath, "test", "https://github.com/AnalogJ/test.git")

	//assert
	require.Error(t, cerr, "should raise an error if cloning to an existing path")
	require.Empty(t, clonePath)
}

func TestGitFetch(t *testing.T) {
	t.Parallel()

	//setup
	dirPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer deleteTestRepo(dirPath)

	//test
	clonePath, cerr := utils.GitClone(dirPath, "cookbook_analogj_test", "https://github.com/AnalogJ/cookbook_analogj_test.git")
	require.NoError(t, cerr)
	ferr := utils.GitFetch(clonePath, "refs/pull/12/merge", "localBranchName")

	//assert
	require.NoError(t, ferr)
}

func TestGitFetch_InvalidDirectory(t *testing.T) {
	t.Parallel()

	//setup
	dirPath := path.Join("this", "path", "does", "not", "exist")

	//test
	ferr := utils.GitFetch(dirPath, "refs/pull/12/merge", "localBranchName")

	//assert
	require.Error(t, ferr)
}

func TestGitCheckout(t *testing.T) {
	t.Parallel()

	//setup
	dirPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer deleteTestRepo(dirPath)
	clonePath, cerr := utils.GitClone(dirPath, "npm_analogj_test", "https://github.com/AnalogJ/npm_analogj_test.git")
	require.NoError(t, cerr)

	//test
	ferr := utils.GitCheckout(clonePath, "branch_test")

	//assert
	require.NoError(t, ferr)
}

func TestGitCheckout_InvalidDirectory(t *testing.T) {
	t.Parallel()

	//setup
	dirPath := path.Join("this", "path", "does", "not", "exist")

	//test
	ferr := utils.GitCheckout(dirPath, "localBranchName")

	//assert
	require.Error(t, ferr)
}

func TestGitCommit(t *testing.T) {
	t.Parallel()

	//setup
	dirPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer deleteTestRepo(dirPath)
	clonePath, cerr := utils.GitClone(dirPath, "commit_to_npm_analogj_test", "https://github.com/AnalogJ/npm_analogj_test.git")
	require.NoError(t, cerr)
	ferr := utils.GitCheckout(clonePath, "branch_test")
	require.NoError(t, ferr)

	//test
	d1 := []byte("hello\nworld\n")
	werr := ioutil.WriteFile(clonePath+"/commit_testfile.txt", d1, 0644)
	require.NoError(t, werr)
	gcerr := utils.GitCommit(clonePath, "Added New File")

	//assert
	require.NoError(t, gcerr)
}

func TestGitCommit_InvalidDirectory(t *testing.T) {
	t.Parallel()

	//setup
	dirPath := path.Join("this", "path", "does", "not", "exist")

	//test
	ferr := utils.GitCommit(dirPath, "message")

	//assert
	require.Error(t, ferr)
}

func TestGitTag(t *testing.T) {
	t.Parallel()

	//setup
	dirPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer deleteTestRepo(dirPath)
	clonePath, cerr := utils.GitClone(dirPath, "add_tag_npm_analogj_test", "https://github.com/AnalogJ/npm_analogj_test.git")
	require.NoError(t, cerr)
	ferr := utils.GitCheckout(clonePath, "branch_test")
	require.NoError(t, ferr)

	//test
	d1 := []byte("hello\nworld\n")
	werr := ioutil.WriteFile(clonePath+"/tag_testfile.txt", d1, 0644)
	require.NoError(t, werr)
	gcerr := utils.GitCommit(clonePath, "Added New File")
	require.NoError(t, gcerr)
	tid, terr := utils.GitTag(clonePath, "v9.9.9")

	//assert
	require.NoError(t, terr)
	require.NotEmpty(t, tid)
}

func TestGitTag_InvalidDirectory(t *testing.T) {
	t.Parallel()

	//setup
	dirPath := path.Join("this", "path", "does", "not", "exist")

	//test
	tag, ferr := utils.GitTag(dirPath, "version")

	//assert
	require.Error(t, ferr)
	require.Empty(t, tag)
}

func TestGitPush(t *testing.T) {
	t.Skip() //Skipping because access_token not available during remote testing.
	dirPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer deleteTestRepo(dirPath)

	clonePath, cerr := utils.GitClone(dirPath, "push_npm_analogj_test", "https://access_token_here:@github.com/AnalogJ/npm_analogj_test.git")
	require.NoError(t, cerr)

	ferr := utils.GitCheckout(clonePath, "branch_test")
	require.NoError(t, ferr)

	//create a new file
	d1 := []byte("hello\nworld\n")
	werr := ioutil.WriteFile(clonePath+"/push_testfile.txt", d1, 0644)
	require.NoError(t, werr)

	gcerr := utils.GitCommit(clonePath, "Added New File")
	require.NoError(t, gcerr)

	perr := utils.GitPush(clonePath, "branch_test", "branch_test")
	require.NoError(t, perr)

}

func TestGitPush_PullRequest(t *testing.T) {
	t.Skip() //Skipping because access_token not available during remote testing.

	//setup
	dirPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer deleteTestRepo(dirPath)
	clonePath, cerr := utils.GitClone(dirPath, "cookbook_analogj_test", "https://access_token_here:@github.com/AnalogJ/cookbook_analogj_test.git")
	require.NoError(t, cerr)

	//test
	ferr := utils.GitFetch(clonePath, "refs/pull/13/merge", "localBranchName")
	require.NoError(t, ferr)
	d1 := []byte("hello\nworld\n")
	werr := ioutil.WriteFile(clonePath+"/push_testfile.txt", d1, 0644)
	require.NoError(t, werr)
	gcerr := utils.GitCommit(clonePath, "Added New File")
	require.NoError(t, gcerr)
	perr := utils.GitPush(clonePath, "localBranchName", "master")

	//test
	require.NoError(t, perr)

}

func TestGitLatestTaggedCommit(t *testing.T) {
	t.Parallel()

	//setup
	dirPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer deleteTestRepo(dirPath)
	clonePath, cerr := utils.GitClone(dirPath, "cookbook_analogj_test", "https://github.com/AnalogJ/cookbook_analogj_test.git")
	require.NoError(t, cerr)

	//test
	tag, ferr := utils.GitLatestTaggedCommit(clonePath)

	//assert
	require.NoError(t, ferr)
	require.Equal(t, "v0.1.11", tag.TagShortName)

}

func TestGitGenerateChangelog(t *testing.T) {
	t.Parallel()

	//setup
	dirPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer deleteTestRepo(dirPath)
	clonePath, cerr := utils.GitClone(dirPath, "cookbook_analogj_test", "https://github.com/AnalogJ/cookbook_analogj_test.git")
	require.NoError(t, cerr)

	//test
	changelog, ferr := utils.GitGenerateChangelog(clonePath, "43adaa328f74fd44abb33d33d8b149ab3780f209", "f3d573aacc59f2a6e2318dd140f3091c16b4b8fe", "")

	//assert
	require.NoError(t, ferr)
	require.Equal(t,
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

func TestGitGenerateChangelog_InvalidDirectory(t *testing.T) {
	t.Parallel()

	//setup
	dirPath := path.Join("this", "path", "does", "not", "exist")

	//test
	changelog, ferr := utils.GitGenerateChangelog(dirPath, "basheSha", "headSha", "fullName")

	//assert
	require.Error(t, ferr)
	require.Empty(t, changelog)
}

func TestGitGenerateGitIgnore(t *testing.T) {
	t.Parallel()

	//setup
	dirPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)


	//test
	ferr := utils.GitGenerateGitIgnore(dirPath, "Ruby")

	//assert
	require.NoError(t, ferr)
	require.True(t, utils.FileExists(path.Join(dirPath, ".gitignore")), "should be generated")
}


func TestGitGenerateGitIgnore_WithInvalidLanguage(t *testing.T) {
	t.Parallel()

	//setup
	dirPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)


	//test
	ferr := utils.GitGenerateGitIgnore(dirPath, "ThisDoesntExist")

	//assert
	require.Error(t, ferr, "should return an error")
	require.False(t, utils.FileExists(path.Join(dirPath, ".gitignore")), "should not generate gitignore")
}

func deleteTestRepo(testRepoDirectory string) {
	os.RemoveAll(testRepoDirectory)
}
