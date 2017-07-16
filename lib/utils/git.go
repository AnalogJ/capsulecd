package utils

import (
	"gopkg.in/libgit2/git2go.v25"
	"path"
	"path/filepath"
	"os"
	"capsulecd/lib/errors"
	"fmt"
	"log"
	"time"
)


func GitClone(parentPath string, repositoryName string, gitRemote string) (string, error) {
	//TODO: credentials may need to be specified
	absPath, aerr := filepath.Abs(path.Join(parentPath, repositoryName))
	if(aerr != nil){
		return "", aerr
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		os.MkdirAll(absPath, os.ModePerm)
	} else {
		return "", errors.ScmFilesystemError(fmt.Sprintf("The local repository path already exists, this should never happen. %s", absPath))
	}

	_, err := git.Clone(gitRemote, absPath, &git.CloneOptions{})
	return absPath, err
}

func GitFetch(repoPath string, remoteRef string, localBranchName string) error {

	repo, oerr := git.OpenRepository(repoPath)
	if(oerr != nil){
		return oerr
	}

	checkoutOpts := &git.CheckoutOpts{
		Strategy: git.CheckoutSafe | git.CheckoutRecreateMissing | git.CheckoutAllowConflicts | git.CheckoutUseTheirs,
	}

	remote, lerr := repo.Remotes.Lookup("origin")
	if(lerr != nil){
		return lerr
	}
	ferr := remote.Fetch([]string{fmt.Sprintf("%s:%s", remoteRef, localBranchName)}, &git.FetchOptions{},"")
	if(ferr != nil){
		return ferr
	}

	localBranch, berr := repo.LookupBranch(localBranchName, git.BranchLocal)
	if(berr != nil){
		return berr
	}

	// Getting the tree for the branch
	localCommit, err := repo.LookupCommit(localBranch.Target())
	if err != nil {
		log.Print("Failed to lookup for commit in local branch " + localBranchName)
		return err
	}
	defer localCommit.Free()

	tree, err := repo.LookupTree(localCommit.TreeId())
	if err != nil {
		log.Print("Failed to lookup for tree " + localBranchName)
		return err
	}
	defer tree.Free()

	// Checkout the tree
	err = repo.CheckoutTree(tree, checkoutOpts)
	if err != nil {
		log.Print("Failed to checkout tree " + localBranchName)
		return err
	}
	// Setting the Head to point to our branch
	return repo.SetHead("refs/heads/" + localBranchName)
}

func GitCheckout(repoPath string, branchName string) error {
	repo, oerr := git.OpenRepository(repoPath)
	if(oerr != nil){
		return oerr
	}

	checkoutOpts := &git.CheckoutOpts{
		Strategy: git.CheckoutSafe | git.CheckoutRecreateMissing | git.CheckoutAllowConflicts | git.CheckoutUseTheirs,
	}
	//Getting the reference for the remote branch
	// remoteBranch, err := repo.References.Lookup("refs/remotes/origin/" + branchName)
	remoteBranch, err := repo.LookupBranch("origin/"+branchName, git.BranchRemote)
	if err != nil {
		log.Print("Failed to find remote branch: " + branchName)
		return err
	}
	defer remoteBranch.Free()

	// Lookup for commit from remote branch
	commit, err := repo.LookupCommit(remoteBranch.Target())
	if err != nil {
		log.Print("Failed to find remote branch commit: " + branchName)
		return err
	}
	defer commit.Free()

	localBranch, err := repo.LookupBranch(branchName, git.BranchLocal)
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
	defer localBranch.Free()

	// Getting the tree for the branch
	localCommit, err := repo.LookupCommit(localBranch.Target())
	if err != nil {
		log.Print("Failed to lookup for commit in local branch " + branchName)
		return err
	}
	defer localCommit.Free()

	tree, err := repo.LookupTree(localCommit.TreeId())
	if err != nil {
		log.Print("Failed to lookup for tree " + branchName)
		return err
	}
	defer tree.Free()

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
	repo, oerr := git.OpenRepository(repoPath)
	if(oerr != nil){
		return oerr
	}

	signature := gitSignature()

	//get repo index.
	idx, ierr := repo.Index()
	if(ierr != nil){return ierr}
	aerr := idx.AddAll([]string{}, git.IndexAddDefault, nil)
	if(aerr != nil){return aerr}
	treeId, wterr := idx.WriteTree()
	if(wterr != nil){return wterr}
	werr := idx.Write()
	if(werr != nil){return werr}

	tree, lerr := repo.LookupTree(treeId)
	if(lerr != nil){return lerr}

	currentBranch, berr := repo.Head()
	if(berr != nil){return berr}

	commitTarget, terr := repo.LookupCommit(currentBranch.Target())
	if terr != nil {return terr}

	_, cerr := repo.CreateCommit("HEAD", signature, signature, message, tree, commitTarget)
	//if(cerr != nil){return cerr}

	return  cerr
}

func GitTag(repoPath string, version string) (string, error) {
	repo, oerr := git.OpenRepository(repoPath)
	if(oerr != nil){return "", oerr}
	commitHead, herr := repo.Head()
	if(herr != nil){return "", herr}

	commit, lerr := repo.LookupCommit(commitHead.Target())
	if(lerr != nil){return "", lerr}

	tagId, terr := repo.Tags.CreateLightweight(version, commit, false) //TODO: this should be an annotated tag.
	return tagId.String(), terr
}

func GitPush(repoPath string, localBranch string, remoteBranch string) error {
	//- https://gist.github.com/danielfbm/37b0ca88b745503557b2b3f16865d8c3
	//- https://stackoverflow.com/questions/37026399/git2go-after-createcommit-all-files-appear-like-being-added-for-deletion
	repo, oerr := git.OpenRepository(repoPath)
	if(oerr != nil){return oerr}

	// Push
	remote, lerr := repo.Remotes.Lookup("origin")
	if(lerr != nil){return lerr}
	//remote.ConnectPush(gitRemoteCallbacks(), &git.ProxyOptions{}, []string{})


	//err = remote.Push([]string{"refs/heads/master"}, nil, signature, message)
	return remote.Push([]string{fmt.Sprintf("refs/heads/%s:refs/heads/%s", localBranch, remoteBranch)}, &git.PushOptions{})
}

func GitLatestTaggedCommit(repoPath string) (string, error) {
	repo, oerr := git.OpenRepository(repoPath)
	if(oerr != nil){return "", oerr}


	_, terr := repo.Tags.List()
	if(terr != nil){return "", terr}

	//return fmt.Sprintf("PRINTING TAGS: %v", tags[0]), nil


	repo.Tags.Foreach(func(name string, id *git.Oid) error {
		log.Printf("TAG name: %s ", name)
		tag, lerr := repo.LookupTag(id)

		//handle lightweight(non-annotated) tags.
		if(lerr != nil){
			//this is a lightweight tag

			commitRef, rerr := repo.LookupCommit(id)
			if(rerr != nil){return nil}

			commitDetails, derr := commitRef.Describe(&git.DescribeOptions{})
			if(derr != nil){return nil}

			commitDetails.Format()
			log.Printf(commitDetails)

			log.Printf("Light-weight tag lookup: %s, commit ID: %s",commitRef.Id().String(), id.String())


		} else {
			log.Printf( "Tag ID: %s, Commit ID: %s, DATE: %s", id.String(), tag.Id().String(), tag.Tagger().When.String())
		}



		//tag, err := r.repository.LookupTag(id)
		//if err != nil {
		//	return err
		//}
		//log.Info(tag.Tagger().When)
		return nil
	})

	return "", nil
}

func GitGenerateChangelog(repoPath string, baseSha string, headSha string, fullName string) (string, error) {
	return "", nil
}

func GitGenerateGitIgnore(repoPath string, ignoreTypes string) (string, error) {
return "", nil
}


//private methods

func gitSignature() *git.Signature {
	return &git.Signature{
		Name: "CapsuleCD",
		Email: "CapsuleCD@users.noreply.github.com",
		When: time.Now(),
	}
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
//	ret, cred := git.NewCredUserpassPlaintext("d1fb4e41af2af60fd255a1106c24df2a0da3b6cf", "") //TODO: remote cred.
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