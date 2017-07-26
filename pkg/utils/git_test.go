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

//
//func TestGitLatestTaggedCommit(t *testing.T) {
//	t.Parallel()
//
//	//setup
//	dirPath, err := ioutil.TempDir("", "")
//	require.NoError(t, err)
//	defer deleteTestRepo(dirPath)
//	clonePath, cerr := utils.GitClone(dirPath, "tags_analogj_test", "https://github.com/AnalogJ/tags_analogj_test.git")
//	require.NoError(t, cerr)
//
//	//test
//	tag, ferr := utils.GitLatestTaggedCommit(clonePath)
//
//	//assert
//	require.NoError(t, ferr)
//	require.Equal(t, "0.4.0", tag.TagShortName)
//
//}
//
//func TestGitLatestTaggedCommit_InvalidDirectory(t *testing.T) {
//	t.Parallel()
//
//	//setup
//	dirPath := path.Join("this", "path", "does", "not", "exist")
//
//	//test
//	tag, ferr := utils.GitLatestTaggedCommit(dirPath)
//
//	//assert
//	require.Error(t, ferr)
//	require.Empty(t, tag)
//}

func TestGitFindNearestTagName_CheckoutMaster(t *testing.T) {
	t.Parallel()

	//setup
	dirPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer deleteTestRepo(dirPath)
	clonePath, cerr := utils.GitClone(dirPath, "tags_analogj_test", "https://github.com/AnalogJ/tags_analogj_test.git")
	require.NoError(t, cerr)
	cerr = utils.GitCheckout(clonePath, "master")
	require.NoError(t, cerr)

	//test
	tag, ferr := utils.GitFindNearestTagName(clonePath)

	//assert
	require.NoError(t, ferr)
	require.Equal(t, "v0.4.1", tag, "should actually be v0.4.1")

}

func TestGitFindNearestTagName_CheckoutBranch(t *testing.T) {
	t.Parallel()

	//setup
	dirPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer deleteTestRepo(dirPath)
	clonePath, cerr := utils.GitClone(dirPath, "tags_analogj_test", "https://github.com/AnalogJ/tags_analogj_test.git")
	require.NoError(t, cerr)
	cerr = utils.GitCheckout(clonePath, "do_not_merge_2")
	require.NoError(t, cerr)

	//test
	tag, ferr := utils.GitFindNearestTagName(clonePath)

	//assert
	require.NoError(t, ferr)
	require.Equal(t, "v0.4.2-rc2", tag, "should actually be v0.4.2-rc2")
}

func TestGitFindNearestTagName_FetchPullRequest(t *testing.T) {
	t.Parallel()

	//setup
	dirPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer deleteTestRepo(dirPath)
	clonePath, cerr := utils.GitClone(dirPath, "tags_analogj_test2", "https://github.com/AnalogJ/tags_analogj_test2.git")
	require.NoError(t, cerr)
	cerr = utils.GitFetch(clonePath, "refs/pull/1/merge", "tagsAnalogJTest2_pr1")
	require.NoError(t, cerr)

	//test
	tag, ferr := utils.GitFindNearestTagName(clonePath)

	//assert
	require.NoError(t, ferr)
	require.Equal(t, "v2.0.0", tag, "should actually be v2.0.0") //this should be 1.0.0 because it happened before the pr opened.

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
	changelog, ferr := utils.GitGenerateChangelog(clonePath, "43adaa328f74fd44abb33d33d8b149ab3780f209", "f3d573aacc59f2a6e2318dd140f3091c16b4b8fe")

	//assert
	require.NoError(t, ferr)
	require.Equal(t, utils.StripIndent(
		`Timestamp |  SHA | Message | Author
	------------- | ------------- | ------------- | -------------
	2017-07-16T01:41Z | f3d573aa | Added New File | CapsuleCD
	2017-07-16T01:26Z | 842436c9 | Merge 39e720e37a19716c098757cb5c78ea90b18111d3 into 97cae66b077de3798995342da781c270b1786820 | Jason Kulatunga
	2017-07-16T01:26Z | 39e720e3 | Update README.md | Jason Kulatunga
	2016-02-28T06:59Z | 97cae66b | (v0.1.11) Automated packaging of release by CapsuleCD | CapsuleCD
	2016-02-28T06:52Z | d4b8c3b5 | Update README.md | Jason Kulatunga
	2016-02-28T00:01Z | d0d3fc8f | (v0.1.10) Automated packaging of release by CapsuleCD | CapsuleCD
	`), changelog)
}

func TestGitGenerateChangelog_InvalidDirectory(t *testing.T) {
	t.Parallel()

	//setup
	dirPath := path.Join("this", "path", "does", "not", "exist")

	//test
	changelog, ferr := utils.GitGenerateChangelog(dirPath, "basheSha", "headSha")

	//assert
	require.Error(t, ferr)
	require.Empty(t, changelog)
}

/*

Support the following case:

1.0.0	       2.0.0
t--o--o--o--o--t--o--o----+   	master
      |			  ^
      |			  | pr
      +--o--o-----o--o----+  	feature-branch


*/

func TestGitGenerateChangelog_TagSincePROpened(t *testing.T) {
	t.Parallel()

	//setup
	dirPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer deleteTestRepo(dirPath)
	clonePath, cerr := utils.GitClone(dirPath, "tags_analogj_test2", "https://github.com/AnalogJ/tags_analogj_test2.git")
	require.NoError(t, cerr)
	cerr = utils.GitFetch(clonePath, "refs/pull/1/merge", "tagsAnalogJTest2_pr1")
	require.NoError(t, cerr)
	//headSha, err := utils.GitHead(clonePath)
	//require.NoError(t, err)

	//test
	changelog, ferr := utils.GitGenerateChangelog(clonePath, "v2.0.0", "tagsAnalogJTest2_pr1")

	//assert
	require.NoError(t, ferr)
	require.Equal(t, utils.StripIndent(
		`Timestamp |  SHA | Message | Author
		------------- | ------------- | ------------- | -------------
		2017-07-24T21:08Z | 4abaa1bd | Merge 9ec0a955d2219f278e569ba5349dcc18a76044a4 into 04e64639394a31231b107ac923d594ea1e3cd257 | Jason Kulatunga
		2017-07-24T21:07Z | 9ec0a955 | test commit | Jason Kulatunga
		2017-07-24T21:06Z | 23e532e8 | test commit | Jason Kulatunga
		2017-07-24T21:06Z | 04e64639 | commit after new branch | Jason Kulatunga
		2017-07-24T21:06Z | 5f5ed1a1 | commit after new branch | Jason Kulatunga
		2017-07-24T21:03Z | 96a7b0d3 | commit before master tag | Jason Kulatunga
		2017-07-24T21:03Z | 67145244 | commit before master tag | Jason Kulatunga
		`), changelog)
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

func TestGitGetTagDetails(t *testing.T) {
	//setup
	dirPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer deleteTestRepo(dirPath)
	clonePath, cerr := utils.GitClone(dirPath, "tags_analogj_test", "https://github.com/AnalogJ/tags_analogj_test.git")
	require.NoError(t, cerr)
	cerr = utils.GitFetch(clonePath, "refs/pull/1/merge", "tagsAnalogJTest_pr1")
	require.NoError(t, cerr)

	//test
	tag1, err1 := utils.GitGetTagDetails(clonePath, "v0.4.1")
	require.NoError(t, err1)

	tag2, err2 := utils.GitGetTagDetails(clonePath, "0.4.0")
	require.NoError(t, err2)

	//assert
	require.Equal(t, "v0.4.1", tag1.TagShortName, "should have correct lightweight tag name")
	require.Equal(t, "ff6fdb84a33b665cf41651eb51d1b86cbf5d3653", tag1.CommitSha, "should have correct lightweight tag sha")

	require.Equal(t, "0.4.0", tag2.TagShortName, "should have correct annotated tag name")
	require.Equal(t, "825711cfb2fb9d44615b415cf723a8590890402f", tag2.CommitSha, "should have correct annotated tag sha")
}

func deleteTestRepo(testRepoDirectory string) {
	os.RemoveAll(testRepoDirectory)
}
