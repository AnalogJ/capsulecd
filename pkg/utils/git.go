package utils

import (
	"capsulecd/pkg/errors"
	"capsulecd/pkg/pipeline"
	stderrors "errors"
	"fmt"
	git2go "gopkg.in/libgit2/git2go.v25"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Clone a git repo into a local directory.
// Credentials need to be specified by embedding in gitRemote url.
// TODO: this pattern may not work on Bitbucket/GitLab
func GitClone(parentPath string, repositoryName string, gitRemote string) (string, error) {
	absPath, _ := filepath.Abs(path.Join(parentPath, repositoryName))

	if !FileExists(absPath) {
		os.MkdirAll(absPath, os.ModePerm)
	} else {
		return "", errors.ScmFilesystemError(fmt.Sprintf("The local repository path already exists, this should never happen. %s", absPath))
	}

	_, err := git2go.Clone(gitRemote, absPath, new(git2go.CloneOptions))
	return absPath, err
}

func GitFetch(repoPath string, remoteRef string, localBranchName string) error {

	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return oerr
	}

	checkoutOpts := &git2go.CheckoutOpts{
		Strategy: git2go.CheckoutSafe | git2go.CheckoutRecreateMissing | git2go.CheckoutAllowConflicts | git2go.CheckoutUseTheirs,
	}

	remote, lerr := repo.Remotes.Lookup("origin")
	if lerr != nil {
		log.Print("Failed to lookup origin remote")
		return lerr
	}
	time.Sleep(time.Second)

	ferr := remote.Fetch([]string{fmt.Sprintf("%s:%s", remoteRef, localBranchName)}, new(git2go.FetchOptions), "")
	if ferr != nil {
		log.Print("Failed to fetch remote ref into new local branch " + fmt.Sprintf("%s:%s", remoteRef, localBranchName))
		return ferr
	}

	time.Sleep(time.Second)
	//should not raise an error when looking for branch (we just created it above)
	localBranch, berr := repo.LookupBranch(localBranchName, git2go.BranchLocal)
	if berr != nil {
		log.Print("Failed to lookup new local branch " + fmt.Sprintf("%s:%s", remoteRef, localBranchName))
		return berr
	}

	// Getting the tree for the branch
	localCommit, err := repo.LookupCommit(localBranch.Target())
	if err != nil {
		log.Print("Failed to lookup for commit in local branch " + fmt.Sprintf("%s:%s", remoteRef, localBranchName))
		return err
	}
	//defer localCommit.Free()

	tree, err := repo.LookupTree(localCommit.TreeId())
	if err != nil {
		log.Print("Failed to lookup for tree " + fmt.Sprintf("%s:%s", remoteRef, localBranchName))
		return err
	}
	//defer tree.Free()

	// Checkout the tree
	err = repo.CheckoutTree(tree, checkoutOpts)
	if err != nil {
		log.Print("Failed to checkout tree " + fmt.Sprintf("%s:%s", remoteRef, localBranchName))
		return err
	}
	// Setting the Head to point to our branch
	return repo.SetHead("refs/heads/" + localBranchName)
}

func GitCheckout(repoPath string, branchName string) error {
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return oerr
	}

	checkoutOpts := &git2go.CheckoutOpts{
		Strategy: git2go.CheckoutSafe | git2go.CheckoutRecreateMissing | git2go.CheckoutAllowConflicts | git2go.CheckoutUseTheirs,
	}
	//Getting the reference for the remote branch
	// remoteBranch, err := repo.References.Lookup("refs/remotes/origin/" + branchName)
	remoteBranch, err := repo.LookupBranch("origin/"+branchName, git2go.BranchRemote)
	if err != nil {
		log.Print("Failed to find remote branch: " + branchName)
		return err
	}
	//defer remoteBranch.Free()

	// Lookup for commit from remote branch
	commit, err := repo.LookupCommit(remoteBranch.Target())
	if err != nil {
		log.Print("Failed to find remote branch commit: " + branchName)
		return err
	}
	//defer commit.Free()

	localBranch, err := repo.LookupBranch(branchName, git2go.BranchLocal)
	// No local branch, lets create one
	if localBranch == nil || err != nil {
		// Creating local branch
		localBranch, err = repo.CreateBranch(branchName, commit, false)
		if err != nil {
			log.Print("Failed to create local branch: " + branchName)
			return err
		}

		// Setting upstream to origin branch
		err = localBranch.SetUpstream("origin/" + branchName)
		if err != nil {
			log.Print("Failed to create upstream to origin/" + branchName)
			return err
		}
	}
	if localBranch == nil {
		return errors.ScmFilesystemError("Error while locating/creating local branch")
	}
	//defer localBranch.Free()

	// Getting the tree for the branch
	localCommit, err := repo.LookupCommit(localBranch.Target())
	if err != nil {
		log.Print("Failed to lookup for commit in local branch " + branchName)
		return err
	}
	//defer localCommit.Free()

	tree, err := repo.LookupTree(localCommit.TreeId())
	if err != nil {
		log.Print("Failed to lookup for tree " + branchName)
		return err
	}
	//defer tree.Free()

	// Checkout the tree
	err = repo.CheckoutTree(tree, checkoutOpts)
	if err != nil {
		log.Print("Failed to checkout tree " + branchName)
		return err
	}
	// Setting the Head to point to our branch
	return repo.SetHead("refs/heads/" + branchName)
}

//Add all modified files to index, and commit.
func GitCommit(repoPath string, message string) error {
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return oerr
	}

	signature := gitSignature()

	//get repo index.
	idx, ierr := repo.Index()
	if ierr != nil {
		return ierr
	}
	aerr := idx.AddAll([]string{}, git2go.IndexAddDefault, nil)
	if aerr != nil {
		return aerr
	}
	treeId, wterr := idx.WriteTree()
	if wterr != nil {
		return wterr
	}
	werr := idx.Write()
	if werr != nil {
		return werr
	}

	tree, lerr := repo.LookupTree(treeId)
	if lerr != nil {
		return lerr
	}

	currentBranch, berr := repo.Head()
	if berr != nil {
		return berr
	}

	commitTarget, terr := repo.LookupCommit(currentBranch.Target())
	if terr != nil {
		return terr
	}

	_, cerr := repo.CreateCommit("HEAD", signature, signature, message, tree, commitTarget)
	//if(cerr != nil){return cerr}

	return cerr
}

func GitTag(repoPath string, version string) (string, error) {
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return "", oerr
	}
	commitHead, herr := repo.Head()
	if herr != nil {
		return "", herr
	}

	commit, lerr := repo.LookupCommit(commitHead.Target())
	if lerr != nil {
		return "", lerr
	}

	//tagId, terr := repo.Tags.CreateLightweight(version, commit, false)
	tagId, terr := repo.Tags.Create(version, commit, gitSignature(), fmt.Sprintf("(%s) Automated packaging of release by CapsuleCD", version))
	if terr != nil {
		return "", terr
	}

	tagObj, terr := repo.LookupTag(tagId)
	return tagObj.TargetId().String(), terr
}

func GitPush(repoPath string, localBranch string, remoteBranch string, tagName string) error {
	//- https://gist.github.com/danielfbm/37b0ca88b745503557b2b3f16865d8c3
	//- https://stackoverflow.com/questions/37026399/git2go-after-createcommit-all-files-appear-like-being-added-for-deletion
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return oerr
	}

	// Push
	remote, lerr := repo.Remotes.Lookup("origin")
	if lerr != nil {
		return lerr
	}
	//remote.ConnectPush(gitRemoteCallbacks(), &git.ProxyOptions{}, []string{})

	//err = remote.Push([]string{"refs/heads/master"}, nil, signature, message)
	return remote.Push([]string{
		fmt.Sprintf("refs/heads/%s:refs/heads/%s", localBranch, remoteBranch),
		fmt.Sprintf("refs/tags/%s:refs/tags/%s", tagName, tagName),
	}, new(git2go.PushOptions))
}

// Get the nearest tag on branch.
// tag must be nearest, ie. sorted by their distance from the HEAD of the branch, not the date or tagname.
// basically `git describe --tags --abbrev=0`
func GitFindNearestTagName(repoPath string) (string, error) {
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return "", oerr
	}

	descOptions, derr := git2go.DefaultDescribeOptions()
	if derr != nil {
		return "", derr
	}
	descOptions.Strategy = git2go.DescribeTags

	formatOptions, ferr := git2go.DefaultDescribeFormatOptions()
	if ferr != nil {
		return "", ferr
	}
	formatOptions.AbbreviatedSize = 0

	descr, derr := repo.DescribeWorkdir(&descOptions)
	if derr != nil {
		return "", derr
	}

	nearestTag, ferr := descr.Format(&formatOptions)
	if ferr != nil {
		return "", ferr
	}

	return nearestTag, nil
}

func GitGenerateChangelog(repoPath string, baseSha string, headSha string) (string, error) {
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return "", oerr
	}

	markdown := StripIndent(`Timestamp |  SHA | Message | Author
	------------- | ------------- | ------------- | -------------
	`)

	revWalk, werr := repo.Walk()
	if werr != nil {
		return "", werr
	}

	rerr := revWalk.PushRange(fmt.Sprintf("%s..%s", baseSha, headSha))
	if rerr != nil {
		return "", rerr
	}

	revWalk.Iterate(func(commit *git2go.Commit) bool {
		markdown += fmt.Sprintf("%s | %.8s | %s | %s\n", //TODO: this should have a link for the SHA.
			commit.Author().When.UTC().Format("2006-01-02T15:04Z"),
			commit.Id().String(),
			cleanCommitMessage(commit.Message()),
			commit.Author().Name,
		)
		return true
	})
	//for {
	//	err := revWalk.Next()
	//	if err != nil {
	//		break
	//	}
	//
	//	log.Info(gi.String())
	//}

	return markdown, nil
}

func GitGenerateGitIgnore(repoPath string, ignoreType string) error {
	//https://github.com/GlenDC/go-gitignore/blob/master/gitignore/provider/github.go

	gitIgnoreBytes, err := getGitIgnore(ignoreType)
	if err != nil {
		return err
	}

	gitIgnorePath := filepath.Join(repoPath, ".gitignore")
	return ioutil.WriteFile(gitIgnorePath, gitIgnoreBytes, 0644)
}

func GitGetTagDetails(repoPath string, tagName string) (*pipeline.GitTagDetails, error) {
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return nil, oerr
	}

	id, aerr := repo.References.Dwim(tagName)
	if aerr != nil {
		return nil, aerr
	}
	tag, lerr := repo.LookupTag(id.Target()) //assume its an annotated tag.

	var currentTag *pipeline.GitTagDetails
	if lerr != nil {
		//this is a lightweight tag, not an annotated tag.
		commitRef, rerr := repo.LookupCommit(id.Target())
		if rerr != nil {
			return nil, rerr
		}

		author := commitRef.Author()

		log.Printf("Light-weight tag (%s) Commit ID: %s, DATE: %s", tagName, commitRef.Id().String(), author.When.String())

		currentTag = &pipeline.GitTagDetails{
			TagShortName: tagName,
			CommitSha:    commitRef.Id().String(),
			CommitDate:   author.When,
		}

	} else {

		log.Printf("Annotated tag (%s) Tag ID: %s, Commit ID: %s, DATE: %s", tagName, tag.Id().String(), tag.TargetId().String(), tag.Tagger().When.String())

		currentTag = &pipeline.GitTagDetails{
			TagShortName: tagName,
			CommitSha:    tag.TargetId().String(),
			CommitDate:   tag.Tagger().When,
		}
	}
	return currentTag, nil

}

//private methods

func gitSignature() *git2go.Signature {
	return &git2go.Signature{
		Name:  "CapsuleCD",
		Email: "CapsuleCD@users.noreply.github.com",
		When:  time.Now(),
	}
}

func cleanCommitMessage(commitMessage string) string {
	commitMessage = strings.TrimSpace(commitMessage)
	if commitMessage == "" {
		return "--"
	}

	commitMessage = strings.Replace(commitMessage, "|", "/", -1)
	commitMessage = strings.Replace(commitMessage, "\n", " ", -1)

	return commitMessage
}

func getGitIgnore(languageName string) ([]byte, error) {
	gitURL := fmt.Sprintf("https://raw.githubusercontent.com/github/gitignore/master/%s.gitignore", languageName)

	resp, err := http.Get(gitURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, stderrors.New(fmt.Sprintf("Could not find .gitignore for '%s'", languageName))
	}

	return ioutil.ReadAll(resp.Body)
}

//func gitRemoteCallbacks() *git.RemoteCallbacks {
//	return  &git.RemoteCallbacks{
//		CredentialsCallback: credentialsCallback,
//		CertificateCheckCallback: certificateCheckCallback,
//	}
//}
//
//func credentialsCallback(url string, username_from_url string, allowed_types git.CredType) (git.ErrorCode, *git.Cred) {
//	log.Printf("This is the CRED URL FOR PUSH: %s %s",url, username_from_url)
//	ret, cred := git.NewCredUserpassPlaintext("placeholder", "") //TODO: remote cred.
//
//	log.Printf("THIS IS THE CRED RESPONS: %s %s", ret, cred)
//	return git.ErrorCode(ret), &cred
//}
//
//func certificateCheckCallback(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
//	if hostname != "github.com" {
//		return git.ErrUser
//	}
//	return git.ErrOk
//}
