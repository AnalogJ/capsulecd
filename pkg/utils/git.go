package utils

import (
	"github.com/analogj/capsulecd/pkg/errors"
	"github.com/analogj/capsulecd/pkg/pipeline"
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

// https://stackoverflow.com/questions/13638235/git-checkout-remote-reference
// https://gist.github.com/danielfbm/ba4ae91efa96bb4771351bdbd2c8b06f
// https://github.com/libgit2/git2go/issues/126
// https://www.atlassian.com/git/articles/pull-request-proficiency-fetching-abilities-unlocked
// https://www.atlassian.com/blog/archives/how-to-fetch-pull-requests
// https://stackoverflow.com/questions/48806891/bitbucket-does-not-update-refspec-for-pr-causing-jenkins-to-build-old-commits
func GitFetchPullRequest(repoPath string, pullRequestNumber string, localBranchName string, srcPatternTmpl string, destPatternTmpl string) error {

	//defaults for Templates if they are not specified.
	if len(srcPatternTmpl) == 0 {
		srcPatternTmpl = "refs/pull/%s/merge" //this default template is for Github
	}

	if len(destPatternTmpl) == 0 {
		destPatternTmpl = "refs/remotes/origin/pr/%s/merge"
	}

	//populate the templates
	srcPattern := fmt.Sprintf(srcPatternTmpl, pullRequestNumber)
	destPattern := fmt.Sprintf(destPatternTmpl, pullRequestNumber)
	refspec := fmt.Sprintf("+%s:%s", srcPattern, destPattern)

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

	// fetch the pull request merge and head references into this repo.
	ferr := remote.Fetch([]string{refspec}, new(git2go.FetchOptions), "")
	if ferr != nil {
		log.Print("Failed to fetch PR reference from remote")
		return ferr
	}

	// Get a reference to the PR merge branch in this repo
	prRef, err := repo.References.Lookup(destPattern)
	if err != nil {
		log.Print("Failed to find PR reference locally: " + destPattern)
		return err
	}

	// Lookup commmit for PR branch
	prCommit, err := repo.LookupCommit(prRef.Target())
	if err != nil {
		log.Print(fmt.Sprintf("Failed to find PR head commit: %s", prRef.Target()))
		return err
	}
	defer prCommit.Free()

	prLocalBranch, err := repo.LookupBranch(localBranchName, git2go.BranchLocal)
	// No local branch, lets create one
	if prLocalBranch == nil || err != nil {
		// Creating local branch
		prLocalBranch, err = repo.CreateBranch(localBranchName, prCommit, false)
		if err != nil {
			log.Print("Failed to create local branch: " + localBranchName)
			return err
		}
	}
	if prLocalBranch == nil {
		return errors.ScmFilesystemError("Error while locating/creating local branch")
	}
	defer prLocalBranch.Free()

	// Getting the tree for the branch
	localCommit, err := repo.LookupCommit(prLocalBranch.Target())
	if err != nil {
		log.Print("Failed to lookup for commit in local branch " + localBranchName)
		return err
	}
	//defer localCommit.Free()

	tree, err := repo.LookupTree(localCommit.TreeId())
	if err != nil {
		log.Print("Failed to lookup for tree " + localBranchName)
		return err
	}
	//defer tree.Free()

	// Checkout the tree
	err = repo.CheckoutTree(tree, checkoutOpts)
	if err != nil {
		log.Print("Failed to checkout tree " + localBranchName)
		return err
	}
	// Setting the Head to point to our branch
	return repo.SetHead("refs/heads/" + localBranchName)
}

// https://github.com/welaw/welaw/blob/100be9cf9a4c6d26f8126678c05072ff725202dd/pkg/easyrepo/merge.go#L11
// https://gist.github.com/danielfbm/37b0ca88b745503557b2b3f16865d8c3
// https://gist.github.com/danielfbm/ba4ae91efa96bb4771351bdbd2c8b06f
// https://github.com/Devying/git2go-example/blob/master/fetch1.go
//https://github.com/jandre/passward/blob/e37bce388cf6417d7123c802add1937574c2b30e/passward/git.go#L186-L206
// https://github.com/electricbookworks/electric-book-gui/blob/4d9ad588dbdf7a94345ef10a1bb6944bc2a2f69a/src/go/src/ebw/git/RepoConflict.go
func GitMergeRemoteBranch(repoPath string, localBranchName string, baseBranchName string, remoteUrl string, remoteBranchName string) error {

	checkoutOpts := &git2go.CheckoutOpts{
		Strategy: git2go.CheckoutSafe | git2go.CheckoutRecreateMissing | git2go.CheckoutAllowConflicts | git2go.CheckoutUseTheirs,
	}

	//get current checked out repository.
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return oerr
	}

	// Lookup commmit for base branch
	baseBranch, err := repo.LookupBranch(baseBranchName, git2go.BranchLocal)
	if err != nil {
		log.Print("Failed to find local base branch: " + baseBranchName)
		return err
	}

	baseCommit, err := repo.LookupCommit(baseBranch.Target())
	if err != nil {
		log.Print(fmt.Sprintf("Failed to find head commit for base branch: %s", baseBranchName))
		return err
	}
	defer baseCommit.Free()

	// Check if there's a local branch with the pr_* name already.
	prLocalBranch, err := repo.LookupBranch(localBranchName, git2go.BranchLocal)
	// No local branch, lets create one
	if prLocalBranch == nil || err != nil {
		// Creating local pr branch from the base branch commit.
		prLocalBranch, err = repo.CreateBranch(localBranchName, baseCommit, false)
		if err != nil {
			log.Print("Failed to create local branch: " + localBranchName)
			return err
		}
	}

	// Getting the tree for the branch
	prLocalCommit, err := repo.LookupCommit(prLocalBranch.Target())
	if err != nil {
		log.Print("Failed to lookup for commit in local branch " + localBranchName)
		return err
	}
	//defer localCommit.Free()

	tree, err := repo.LookupTree(prLocalCommit.TreeId())
	if err != nil {
		log.Print("Failed to lookup for tree " + localBranchName)
		return err
	}
	//defer tree.Free()

	// Checkout the tree
	err = repo.CheckoutTree(tree, checkoutOpts)
	if err != nil {
		log.Print("Failed to checkout tree " + localBranchName)
		return err
	}
	// Setting the Head to point to our branch
	herr := repo.SetHead("refs/heads/" + localBranchName)
	if herr != nil {
		return herr
	}

	//add a new remote for the PR head.
	prRemoteAlias := "pr_origin"
	prRemote, rerr := repo.Remotes.Create(prRemoteAlias, remoteUrl)
	if rerr != nil {
		return rerr
	}

	//fetch the commits for the remoteBranchName
	rferr := prRemote.Fetch([]string{"refs/heads/" + remoteBranchName}, new(git2go.FetchOptions), "")
	if rferr != nil {
		return rferr
	}

	remoteBranch, errRef := repo.References.Lookup(fmt.Sprintf("refs/remotes/%s/%s", prRemoteAlias, remoteBranchName))
	if errRef != nil {
		return errRef
	}
	remoteBranchID := remoteBranch.Target()

	//Assuming we are already checkout as the destination branch
	remotePrAnnCommit, err := repo.AnnotatedCommitFromRef(remoteBranch)
	if err != nil {
		log.Print("Failed get annotated commit from remote ")
		return err
	}
	defer remotePrAnnCommit.Free()

	//Getting repo HEAD
	head, err := repo.Head()
	if err != nil {
		log.Print("Failed get head ")
		return err
	}

	// Do merge analysis
	mergeHeads := make([]*git2go.AnnotatedCommit, 1)
	mergeHeads[0] = remotePrAnnCommit
	analysis, _, err := repo.MergeAnalysis(mergeHeads)

	if analysis&git2go.MergeAnalysisNone != 0 || analysis&git2go.MergeAnalysisUpToDate != 0 {
		log.Print("Found nothing to merge. This should not happen for valid PR's")
		return errors.ScmMergeNothingToMergeError("Found nothing to merge. This should not happen for valid PR's")
	} else if analysis&git2go.MergeAnalysisNormal != 0 {
		// Just merge changes

		//Options for merge
		mergeOpts, err := git2go.DefaultMergeOptions()
		if err != nil {
			return err
		}
		mergeOpts.FileFavor = git2go.MergeFileFavorNormal
		mergeOpts.TreeFlags = git2go.MergeTreeFailOnConflict

		//Options for checkout
		mergeCheckoutOpts := &git2go.CheckoutOpts{
			Strategy: git2go.CheckoutSafe | git2go.CheckoutRecreateMissing | git2go.CheckoutUseTheirs,
		}

		//Merge action
		if err = repo.Merge(mergeHeads, &mergeOpts, mergeCheckoutOpts); err != nil {
			log.Print("Failed to merge heads")
			// Check for conflicts
			index, err := repo.Index()
			if err != nil {
				log.Print("Failed to get repo index")
				return err
			}
			if index.HasConflicts() {
				log.Printf("Conflicts encountered. Please resolve them. %v", err)
				return errors.ScmMergeConflictError("Merge resulted in conflicts. Please solve the conflicts before merging.")
			}
			return err
		}

		//Getting repo Index
		index, err := repo.Index()
		if err != nil {
			log.Print("Failed to get repo index")
			return err
		}
		defer index.Free()

		//Checking for conflicts
		if index.HasConflicts() {
			return errors.ScmMergeConflictError("Merge resulted in conflicts. Please solve the conflicts before merging.")
		}

		// Make the merge commit
		sig := gitSignature()

		// Get Write Tree
		treeId, err := index.WriteTree()
		if err != nil {
			return err
		}

		tree, err := repo.LookupTree(treeId)
		if err != nil {
			return err
		}

		localCommit, err := repo.LookupCommit(head.Target())
		if err != nil {
			return err
		}

		remoteCommit, err := repo.LookupCommit(remoteBranchID)
		if err != nil {
			return err
		}

		repo.CreateCommit("HEAD", sig, sig, "", tree, localCommit, remoteCommit)
		// Clean up
		repo.StateCleanup()
	} else if analysis&git2go.MergeAnalysisFastForward != 0 {
		// Fast-forward changes
		// Get remote tree

		remoteTree, err := repo.LookupTree(remoteBranchID)
		if err != nil {
			return err
		}

		// Checkout
		if err := repo.CheckoutTree(remoteTree, nil); err != nil {
			return err
		}

		// Point branch to the object
		prLocalBranch.SetTarget(remoteBranchID, "")
		if _, err := head.SetTarget(remoteBranchID, ""); err != nil {
			return err
		}

	} else {
		log.Printf("Unexpected merge analysis result %d", analysis)
		return errors.ScmMergeAnalysisUnknownError(fmt.Sprintf("Unexpected merge analysis result: %d", analysis))
	}
	return nil

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

func GitTag(repoPath string, version string, message string) (string, error) {
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
	tagId, terr := repo.Tags.Create(version, commit, gitSignature(), fmt.Sprintf("(%s) %s", version, message))
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
