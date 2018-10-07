package scm

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/pipeline"
	"github.com/analogj/go-bitbucket"
	"net/http"
	"capsulecd/pkg/errors"
	"os"
	"io/ioutil"
	"log"
	"strings"
	"github.com/mitchellh/mapstructure"
	"fmt"
	"time"
	"capsulecd/pkg/utils"
	"strconv"
	"path"
)

type scmBitbucket struct {
	Config       config.Interface
	Client       *bitbucket.Client
	PipelineData *pipeline.Data
}


type scmBitbucketPullrequest struct {
	CreatedOn time.Time `mapstructure:"created_on"`
	PullRequestNumber int `mapstructure:"id"`
	State string `mapstructure:"state"`
	Title string `mapstructure:"title"`
	Base struct {
		Branch struct {
			Name string  `mapstructure:"name"`
		}  `mapstructure:"branch"`
		Commit struct {
			Hash string `mapstructure:"hash"`
		} `mapstructure:"commit"`
		Repository struct {
			FullName string `mapstructure:"full_name"`
			Name string `mapstructure:"name"`
		} `mapstructure:"repository"`
	}  `mapstructure:"destination"`

	Head struct {
		Branch struct {
			Name string  `mapstructure:"name"`
		}  `mapstructure:"branch"`
		Commit struct {
			Hash string `mapstructure:"hash"`
		} `mapstructure:"commit"`
		Repository struct {
			FullName string `mapstructure:"full_name"`
			Name string `mapstructure:"name"`
		} `mapstructure:"repository"`
	}  `mapstructure:"source"`

}



// configure method will generate an authenticated client that can be used to comunicate with Github
// MUST set @git_parent_path
// MUST set @client field
func (b *scmBitbucket) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	b.PipelineData = pipelineData
	b.Config = myconfig

	if !b.Config.IsSet("scm_bitbucket_username") {
		return errors.ScmAuthenticationFailed("Missing bitbucket username")
	}
	if !b.Config.IsSet("scm_bitbucket_password") && !b.Config.IsSet("scm_bitbucket_access_token") {
		return errors.ScmAuthenticationFailed("Bitbucket app password or access token is required")
	}
	if b.Config.IsSet("scm_git_parent_path") {
		b.PipelineData.GitParentPath = b.Config.GetString("scm_git_parent_path")
		os.MkdirAll(b.PipelineData.GitParentPath, os.ModePerm)
	} else {
		dirPath, _ := ioutil.TempDir("", "")
		b.PipelineData.GitParentPath = dirPath
	}

	if b.Config.IsSet("scm_bitbucket_password") {
		b.Client = bitbucket.NewBasicAuth(b.Config.GetString("scm_bitbucket_username"), b.Config.GetString("scm_bitbucket_password"))
	} else {
		b.Client = bitbucket.NewOAuthbearerToken(b.Config.GetString("scm_bitbucket_access_token"))
	}
	if client != nil {
		//primarily used for testing.
		b.Client.HttpClient = client
	}

	return nil
}

func (b *scmBitbucket) RetrievePayload() (*Payload, error) {
	if !b.Config.IsSet("scm_pull_request") {
		log.Print("This is not a pull request. No automatic continuous deployment processing required. Continuous Integration testing will continue.")
		b.PipelineData.IsPullRequest = false

		return &Payload{
			Head: &pipeline.ScmCommitInfo{
				Sha: b.Config.GetString("scm_sha"),
				Ref: b.Config.GetString("scm_branch"),
				Repo: &pipeline.ScmRepoInfo{
					CloneUrl: b.Config.GetString("scm_clone_url"),
					Name:     b.Config.GetString("scm_repo_name"),
					FullName: b.Config.GetString("scm_repo_full_name"),
				}},
		}, nil
		//make this as similar to a pull request as possible
	} else {
		b.PipelineData.IsPullRequest = true
		parts := strings.Split(b.Config.GetString("scm_repo_full_name"), "/")
		prDataMap, err := b.Client.Repositories.PullRequests.Get(&bitbucket.PullRequestsOptions{
			ID: b.Config.GetString("scm_pull_request"),
			Owner: parts[0],
			RepoSlug: parts[1],
		})
		if err != nil {
			return nil, errors.ScmAuthenticationFailed("Could not retrieve pull request from Bitbucket")
		}


		prData  := new(scmBitbucketPullrequest)
		mapstructure.Decode(prDataMap, prData)

		//validate pullrequest
		if strings.ToLower(prData.State) != "open" {
			return nil, errors.ScmPayloadUnsupported("Pull request has an invalid action")
		}
		//TODO: see if we can determien the "main branch" using the Bitbucket API.
		//if pr.Base.Repo.GetDefaultBranch() != pr.Base.GetRef() {
		//	return nil, errors.ScmPayloadUnsupported(fmt.Sprintf("Pull request is not being created against the default branch of this repository (%s vs %s)", pr.Base.Repo.GetDefaultBranch(), pr.Base.GetRef()))
		//}

		//TODO: figure out how to do optional authenication. possible options, Source USER, token based auth, no auth when used with capsulecd.com.
		// unless @source_client.collaborator?(payload['base']['repo']['full_name'], payload['user']['login'])
		//
		//   @source_client.add_comment(payload['base']['repo']['full_name'], payload['number'], CapsuleCD::BotUtils.pull_request_comment)
		//   fail CapsuleCD::Error::SourceUnauthorizedUser, 'Pull request was opened by an unauthorized user'
		// end

		return &Payload{
			Title:             prData.Title,
			PullRequestNumber: strconv.Itoa(prData.PullRequestNumber),
			Head: &pipeline.ScmCommitInfo{
				Sha: prData.Head.Commit.Hash,
				Ref: prData.Head.Branch.Name,
				Repo: &pipeline.ScmRepoInfo{
					CloneUrl: fmt.Sprintf("https://bitbucket.org/%s.git", prData.Head.Repository.FullName),
					Name:     prData.Head.Repository.Name,
					FullName: prData.Head.Repository.FullName,
				},
			},
			Base: &pipeline.ScmCommitInfo{
				Sha: prData.Base.Commit.Hash,
				Ref: prData.Base.Branch.Name,
				Repo: &pipeline.ScmRepoInfo{
					CloneUrl: fmt.Sprintf("https://bitbucket.org/%s.git", prData.Base.Repository.FullName),
					Name:     prData.Base.Repository.Name,
					FullName: prData.Base.Repository.FullName,
				},
			},
		}, nil
	}
	return nil, nil
}

func (b *scmBitbucket) CheckoutPushPayload(payload *Payload) error {
	//set the processed head info
	b.PipelineData.GitHeadInfo = payload.Head
	if err := b.PipelineData.GitHeadInfo.Validate(); err != nil {
		return err
	}

	var cloneCred string
	if b.Config.IsSet("scm_bitbucket_password") {
		cloneCred = b.Config.GetString("scm_bitbucket_password")
	} else {
		cloneCred = b.Config.GetString("scm_bitbucket_access_token")
	}

	authRemote, aerr := authGitRemote(
		b.PipelineData.GitHeadInfo.Repo.CloneUrl,
		b.Config.GetString("scm_bitbucket_username"),
		cloneCred,
	)
	if aerr != nil {
		return aerr
	}
	b.PipelineData.GitRemote = authRemote
	b.PipelineData.GitLocalBranch = b.PipelineData.GitHeadInfo.Ref

	// clone the merged branch
	// https://sethvargo.com/checkout-a-github-pull-request/
	// https://coderwall.com/p/z5rkga/github-checkout-a-pull-request-as-a-branch

	gitLocalPath, cerr := utils.GitClone(b.PipelineData.GitParentPath, b.PipelineData.GitHeadInfo.Repo.Name, b.PipelineData.GitRemote)
	if cerr != nil {
		return cerr
	}
	b.PipelineData.GitLocalPath = gitLocalPath

	if cerr := utils.GitCheckout(b.PipelineData.GitLocalPath, b.PipelineData.GitHeadInfo.Ref); cerr != nil {
		return cerr
	}

	//retrieve and store the nearestTag to this commit.
	nearestTag, err := utils.GitFindNearestTagName(gitLocalPath)
	if err != nil {
		return nil // we dont care about failures finding the nearest tag, we'll just have an empty changelog.
	}

	tagDetails, err := utils.GitGetTagDetails(gitLocalPath, nearestTag)
	if err != nil {
		return nil // we dont care about failures finding the nearest tag, we'll just have an empty changelog.
	}
	b.PipelineData.GitNearestTag = tagDetails

	return nil
}

func (b *scmBitbucket) CheckoutPullRequestPayload(payload *Payload) error {
	//set the processed head info
	b.PipelineData.GitHeadInfo = payload.Head
	b.PipelineData.GitBaseInfo = payload.Base

	herr := b.PipelineData.GitHeadInfo.Validate()
	berr := b.PipelineData.GitBaseInfo.Validate()
	if herr != nil {
		return herr
	} else if berr != nil {
		return berr
	}

	var cloneCred string
	if b.Config.IsSet("scm_bitbucket_password") {
		cloneCred = b.Config.GetString("scm_bitbucket_password")
	} else {
		cloneCred = b.Config.GetString("scm_bitbucket_access_token")
	}

	authBaseRemoteUrl, aberr := authGitRemote(
		b.PipelineData.GitBaseInfo.Repo.CloneUrl,
		b.Config.GetString("scm_bitbucket_username"),
		cloneCred,
	)
	if aberr != nil {
		return aberr
	}
	b.PipelineData.GitRemote = authBaseRemoteUrl


	authHeadRemoteUrl, aherr := authGitRemote(
		b.PipelineData.GitBaseInfo.Repo.CloneUrl,
		b.Config.GetString("scm_bitbucket_username"),
		cloneCred,
	)
	if aherr != nil {
		return aherr
	}


	// clone the merged branch
	// https://sethvargo.com/checkout-a-github-pull-request/
	// https://coderwall.com/p/z5rkga/github-checkout-a-pull-request-as-a-branch
	// https://help.github.com/articles/checking-out-pull-requests-locally/

	gitLocalPath, cerr := utils.GitClone(b.PipelineData.GitParentPath, b.PipelineData.GitBaseInfo.Repo.Name, b.PipelineData.GitRemote)
	if cerr != nil {
		return cerr
	}
	b.PipelineData.GitLocalPath = gitLocalPath
	b.PipelineData.GitLocalBranch = fmt.Sprintf("pr_%s", payload.PullRequestNumber)

	ferr := utils.GitMergeRemoteBranch(b.PipelineData.GitLocalPath, b.PipelineData.GitLocalBranch, b.PipelineData.GitBaseInfo.Ref, authHeadRemoteUrl, b.PipelineData.GitHeadInfo.Ref)
	if ferr != nil {
		return ferr
	}

	// show a processing message on the github PR.
	b.Notify(b.PipelineData.GitHeadInfo.Sha, "pending", "Started processing package. Pull request will be merged automatically when complete.")

	//retrieve and store the nearestTag to this commit.
	nearestTag, err := utils.GitFindNearestTagName(gitLocalPath)
	if err != nil {
		return nil // we dont care about failures finding the nearest tag, we'll just have an empty changelog.
	}

	tagDetails, err := utils.GitGetTagDetails(gitLocalPath, nearestTag)
	if err != nil {
		return nil // we dont care about failures finding the nearest tag, we'll just have an empty changelog.
	}
	b.PipelineData.GitNearestTag = tagDetails

	return nil
}

func (b *scmBitbucket) Publish() error {

	// push the version bumped metadata file + newly created files to
	perr := utils.GitPush(b.PipelineData.GitLocalPath, b.PipelineData.GitLocalBranch, b.PipelineData.GitBaseInfo.Ref, fmt.Sprintf("v%s", b.PipelineData.ReleaseVersion))
	if perr != nil {
		return perr
	}
	//sleep because bitbucket needs time to process the new tag.
	time.Sleep(5 * time.Second)


	//TODO: Bitbucket does not seem to support Github style releases.

	//// calculate the release sha
	//releaseSha := utils.LeftPad2Len(b.PipelineData.ReleaseCommit, "0", 40)
	//
	////get the release changelog
	//// logic is complicated.
	//// If this is a push we can only do a tag-tag Changelog
	//// If this is a pull request we can do either
	//// if disable_nearest_tag_changelog is true, we must attempt
	//var releaseBody string = ""
	//if b.PipelineData.GitNearestTag != nil && !b.Config.GetBool("scm_disable_nearest_tag_changelog") {
	//	releaseBody, _ = utils.GitGenerateChangelog(
	//		b.PipelineData.GitLocalPath,
	//		b.PipelineData.GitNearestTag.TagShortName,
	//		b.PipelineData.GitLocalBranch,
	//	)
	//}
	////fallback to using diff if pullrequest.
	//if b.PipelineData.IsPullRequest && releaseBody == "" {
	//	releaseBody, _ = utils.GitGenerateChangelog(
	//		b.PipelineData.GitLocalPath,
	//		b.PipelineData.GitBaseInfo.Sha,
	//		b.PipelineData.GitHeadInfo.Sha,
	//	)
	//}

	////create release.
	//ctx := context.Background()
	//parts := strings.Split(b.Config.GetString("scm_repo_full_name"), "/")
	//version := fmt.Sprintf("v%s", b.PipelineData.ReleaseVersion)
	//
	//log.Printf("Creating new release for `%s/%s` with version: `%s` on commit: `%s`. Commit message: `%s`", parts[0], parts[1], version, releaseSha, releaseBody)
	//
	//releaseData, _, rerr := g.Client.Repositories.CreateRelease(
	//	ctx,
	//	parts[0],
	//	parts[1],
	//	&github.RepositoryRelease{
	//		TargetCommitish: &releaseSha,
	//		Body:            &releaseBody,
	//		TagName:         &version,
	//		Name:            &version,
	//	},
	//)
	//if rerr != nil {
	//	return rerr
	//}

	//TODO add publish assets support
	if perr := b.PublishAssets(nil); perr != nil {
		log.Print("An error occured while publishing assets:")
		log.Print(perr)
		log.Print("Continuing...")
	}

	return nil
}

func (b *scmBitbucket) PublishAssets(releaseData interface{}) error {

	parts := strings.Split(b.Config.GetString("scm_repo_full_name"), "/")

	for _, assetData := range b.PipelineData.ReleaseAssets {
		// handle templated destination artifact names
		artifactNamePopulated, aerr := utils.PopulateTemplate(assetData.ArtifactName, b.PipelineData)
		if aerr != nil {
			return aerr
		}

		localPathPopulated, lerr := utils.PopulateTemplate(assetData.LocalPath, b.PipelineData)
		if lerr != nil {
			return lerr
		}

		b.publishAsset(
			b.Client,
			parts[0],
			parts[1],
			artifactNamePopulated,
			path.Join(b.PipelineData.GitLocalPath, localPathPopulated),
			5)
	}
	return nil
}

func (b *scmBitbucket) Cleanup() error {
	if !b.Config.GetBool("scm_enable_branch_cleanup") { //Default is false, so this will just return without doing anything.
		// - exit if "scm_enable_branch_cleanup" is not true
		return errors.ScmCleanupFailed("scm_enable_branch_cleanup is false. Skipping cleanup")
	} else if !b.PipelineData.IsPullRequest {
		return errors.ScmCleanupFailed("scm cleanup unnecessary for push's. Skipping cleanup")
	} else if b.PipelineData.GitHeadInfo.Repo.FullName != b.PipelineData.GitBaseInfo.Repo.FullName {
		// exit if the HEAD PR branch is not in the same organization and repository as the BASE
		return errors.ScmCleanupFailed("HEAD PR branch is not in the same organization & repo as the BASE. Skipping cleanup")
	}

	//TODO: deleting a branch is only supported on BB Server not BB Cloud

	//parts := strings.Split(b.PipelineData.GitBaseInfo.Repo.FullName, "/")
	//
	//repoData, _, err := b.Client.Repositories.Get(ctx, parts[0], parts[1])
	//if err != nil {
	//	return err
	//}
	//
	//if b.PipelineData.GitHeadInfo.Ref == repoData.GetDefaultBranch() || b.PipelineData.GitHeadInfo.Ref == "master" {
	//	//exit if the HEAD branch is the repo default branch
	//	//exit if the HEAD branch is master
	//	return errors.ScmCleanupFailed("HEAD PR branch is default repo branch, or master. Skipping cleanup")
	//}
	//
	//_, drerr := b.Client.Git.DeleteRef(ctx, parts[0], parts[1], fmt.Sprintf("heads/%s", b.PipelineData.GitHeadInfo.Ref))
	//if drerr != nil {
	//	return drerr
	//}

	return nil
}

func (b *scmBitbucket) Notify(ref string, state /*pending, failure, success*/ string, message string) error {
	//https://developer.atlassian.com/bitbucket/api/2/reference/resource/repositories/%7Busername%7D/%7Brepo_slug%7D/commit/%7Bnode%7D/statuses/build

	targetURL := "https://www.capsulecd.com"
	contextApp := "CapsuleCD"

	parts := strings.Split(b.Config.GetString("scm_repo_full_name"), "/")

	co := bitbucket.CommitsOptions{
		Owner: parts[0],
		RepoSlug: parts[1],
		Revision: ref,
	}

	cso := bitbucket.CommitStatusOptions{
		Key:		 "build",
		State:       b.convertNotifyState(state),
		Url:         targetURL,
		Name:        contextApp,
		Description: message,
	}

	_, err := b.Client.Repositories.Commits.CreateCommitStatus(&co, &cso)
	return err
}


func (b *scmBitbucket) convertNotifyState(state string) string {
	switch state {
	case "pending":
		return "INPROGRESS"
	case "failure":
		return "FAILED"
	case "success":
		return "SUCCESSFUL"
	default:
		return "INPROGRESS"
	}
}


//private

func (b *scmBitbucket) publishAsset(client *bitbucket.Client, repoOwner string, repoName string, assetName, filePath string, retries int) error {

	log.Printf("Attempt (%d) to upload release asset %s from %s", retries, assetName, filePath)

	dl := bitbucket.DownloadsOptions{
		Owner: repoOwner,
		RepoSlug: repoName,
		FilePath: filePath,
		FileName: assetName,
	}

	 _, err := client.Repositories.Downloads.Create(&dl)

	if err != nil && retries > 0 {
		fmt.Println("artifact upload errored out, retrying in one second. Err:", err)
		time.Sleep(time.Second)
		err = b.publishAsset(client, repoOwner, repoName, assetName, filePath, retries-1)
	}

	return err
}
