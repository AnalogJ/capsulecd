package scm

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/errors"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/utils"
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type scmGithub struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *github.Client
}

func (g *scmGithub) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	g.PipelineData = pipelineData
	g.Config = myconfig

	if !g.Config.IsSet("scm_github_access_token") {
		return errors.ScmAuthenticationFailed("Missing github access token")
	}
	if g.Config.IsSet("scm_git_parent_path") {
		g.PipelineData.GitParentPath = g.Config.GetString("scm_git_parent_path")
		os.MkdirAll(g.PipelineData.GitParentPath, os.ModePerm)
	} else {
		dirPath, _ := ioutil.TempDir("", "")
		g.PipelineData.GitParentPath = dirPath
	}

	if client != nil {
		//primarily used for testing.
		g.Client = github.NewClient(client)
	} else {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: g.Config.GetString("scm_github_access_token")},
		)
		tc := oauth2.NewClient(ctx, ts)

		//TODO: autopaginate turned on.
		//TODO: add support for alternative api endpoints "scm_github_api_endpoint"
		g.Client = github.NewClient(tc)
	}

	if g.Config.IsSet("scm_github_api_endpoint") {

		apiUrl, aerr := url.Parse(g.Config.GetString("scm_github_api_endpoint"))
		if aerr != nil {
			return aerr
		}
		g.Client.BaseURL = apiUrl
	}

	return nil
}

func (g *scmGithub) RetrievePayload() (*Payload, error) {
	if !g.Config.IsSet("scm_pull_request") {
		log.Print("This is not a pull request. No automatic continuous deployment processing required. Continuous Integration testing will continue.")
		g.PipelineData.IsPullRequest = false

		return &Payload{
			Head: &pipeline.ScmCommitInfo{
				Sha: g.Config.GetString("scm_sha"),
				Ref: g.Config.GetString("scm_branch"),
				Repo: &pipeline.ScmRepoInfo{
					CloneUrl: g.Config.GetString("scm_clone_url"),
					Name:     g.Config.GetString("scm_repo_name"),
					FullName: g.Config.GetString("scm_repo_full_name"),
				}},
		}, nil
		//make this as similar to a pull request as possible
	} else {
		g.PipelineData.IsPullRequest = true
		ctx := context.Background()
		parts := strings.Split(g.Config.GetString("scm_repo_full_name"), "/")
		pr, _, err := g.Client.PullRequests.Get(ctx, parts[0], parts[1], g.Config.GetInt("scm_pull_request"))

		if err != nil {
			return nil, errors.ScmAuthenticationFailed(fmt.Sprintf("Could not retrieve pull request from Github: %s", err))
		}

		//validate pullrequest
		if pr.GetState() != "open" {
			return nil, errors.ScmPayloadUnsupported("Pull request has an invalid action")
		}
		if pr.Base.Repo.GetDefaultBranch() != pr.Base.GetRef() {
			return nil, errors.ScmPayloadUnsupported(fmt.Sprintf("Pull request is not being created against the default branch of this repository (%s vs %s)", pr.Base.Repo.GetDefaultBranch(), pr.Base.GetRef()))
		}
		// check the payload push user.

		//TODO: figure out how to do optional authenication. possible options, Source USER, token based auth, no auth when used with capsulecd.com.
		// unless @source_client.collaborator?(payload['base']['repo']['full_name'], payload['user']['login'])
		//
		//   @source_client.add_comment(payload['base']['repo']['full_name'], payload['number'], CapsuleCD::BotUtils.pull_request_comment)
		//   fail CapsuleCD::Error::SourceUnauthorizedUser, 'Pull request was opened by an unauthorized user'
		// end

		return &Payload{
			Title:             pr.GetTitle(),
			PullRequestNumber: strconv.Itoa(pr.GetNumber()),
			Head: &pipeline.ScmCommitInfo{
				Sha: pr.Head.GetSHA(),
				Ref: pr.Head.GetRef(),
				Repo: &pipeline.ScmRepoInfo{
					CloneUrl: pr.Head.Repo.GetCloneURL(),
					Name:     pr.Head.Repo.GetName(),
					FullName: pr.Head.Repo.GetFullName(),
				},
			},
			Base: &pipeline.ScmCommitInfo{
				Sha: pr.Base.GetSHA(),
				Ref: pr.Base.GetRef(),
				Repo: &pipeline.ScmRepoInfo{
					CloneUrl: pr.Base.Repo.GetCloneURL(),
					Name:     pr.Base.Repo.GetName(),
					FullName: pr.Base.Repo.GetFullName(),
				},
			},
		}, nil
	}
}

func (g *scmGithub) CheckoutPushPayload(payload *Payload) error {
	//set the processed head info
	g.PipelineData.GitHeadInfo = payload.Head
	if err := g.PipelineData.GitHeadInfo.Validate(); err != nil {
		return err
	}

	authRemote, aerr := authGitRemote(g.PipelineData.GitHeadInfo.Repo.CloneUrl, g.Config.GetString("scm_github_access_token"))
	if aerr != nil {
		return aerr
	}
	g.PipelineData.GitRemote = authRemote
	g.PipelineData.GitLocalBranch = g.PipelineData.GitHeadInfo.Ref

	// clone the merged branch
	// https://sethvargo.com/checkout-a-github-pull-request/
	// https://coderwall.com/p/z5rkga/github-checkout-a-pull-request-as-a-branch

	gitLocalPath, cerr := utils.GitClone(g.PipelineData.GitParentPath, g.PipelineData.GitHeadInfo.Repo.Name, g.PipelineData.GitRemote)
	if cerr != nil {
		return cerr
	}
	g.PipelineData.GitLocalPath = gitLocalPath

	if cerr := utils.GitCheckout(g.PipelineData.GitLocalPath, g.PipelineData.GitHeadInfo.Ref); cerr != nil {
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
	g.PipelineData.GitNearestTag = tagDetails

	return nil
}

func (g *scmGithub) CheckoutPullRequestPayload(payload *Payload) error {
	//set the processed head info
	g.PipelineData.GitHeadInfo = payload.Head
	g.PipelineData.GitBaseInfo = payload.Base
	herr := g.PipelineData.GitHeadInfo.Validate()
	berr := g.PipelineData.GitBaseInfo.Validate()
	if herr != nil {
		return herr
	} else if berr != nil {
		return berr
	}

	authRemote, aerr := authGitRemote(g.PipelineData.GitHeadInfo.Repo.CloneUrl, g.Config.GetString("scm_github_access_token"))
	if aerr != nil {
		return aerr
	}
	g.PipelineData.GitRemote = authRemote

	// clone the merged branch
	// https://sethvargo.com/checkout-a-github-pull-request/
	// https://coderwall.com/p/z5rkga/github-checkout-a-pull-request-as-a-branch
	// https://help.github.com/articles/checking-out-pull-requests-locally/

	gitLocalPath, cerr := utils.GitClone(g.PipelineData.GitParentPath, g.PipelineData.GitHeadInfo.Repo.Name, g.PipelineData.GitRemote)
	if cerr != nil {
		return cerr
	}
	g.PipelineData.GitLocalPath = gitLocalPath
	g.PipelineData.GitLocalBranch = fmt.Sprintf("pr_%s", payload.PullRequestNumber)

	ferr := utils.GitFetch(g.PipelineData.GitLocalPath, fmt.Sprintf("refs/pull/%s/merge", payload.PullRequestNumber), g.PipelineData.GitLocalBranch)
	if ferr != nil {
		return ferr
	}

	// show a processing message on the github PR.
	g.Notify(g.PipelineData.GitHeadInfo.Sha, "pending", "Started processing package. Pull request will be merged automatically when complete.")

	//retrieve and store the nearestTag to this commit.
	nearestTag, err := utils.GitFindNearestTagName(gitLocalPath)
	if err != nil {
		return nil // we dont care about failures finding the nearest tag, we'll just have an empty changelog.
	}

	tagDetails, err := utils.GitGetTagDetails(gitLocalPath, nearestTag)
	if err != nil {
		return nil // we dont care about failures finding the nearest tag, we'll just have an empty changelog.
	}
	g.PipelineData.GitNearestTag = tagDetails

	return nil
}

func (g *scmGithub) Publish() error {

	// push the version bumped metadata file + newly created files to
	perr := utils.GitPush(g.PipelineData.GitLocalPath, g.PipelineData.GitLocalBranch, g.PipelineData.GitBaseInfo.Ref, fmt.Sprintf("v%s", g.PipelineData.ReleaseVersion))
	if perr != nil {
		return perr
	}
	//sleep because github needs time to process the new tag.
	time.Sleep(5 * time.Second)

	// calculate the release sha
	releaseSha := utils.LeftPad2Len(g.PipelineData.ReleaseCommit, "0", 40)

	//get the release changelog
	// logic is complicated.
	// If this is a push we can only do a tag-tag Changelog
	// If this is a pull request we can do either
	// if disable_nearest_tag_changelog is true, we must attempt
	var releaseBody string = ""
	if g.PipelineData.GitNearestTag != nil && !g.Config.GetBool("scm_disable_nearest_tag_changelog") {
		releaseBody, _ = utils.GitGenerateChangelog(
			g.PipelineData.GitLocalPath,
			g.PipelineData.GitNearestTag.TagShortName,
			g.PipelineData.GitLocalBranch,
		)
	}
	//fallback to using diff if pullrequest.
	if g.PipelineData.IsPullRequest && releaseBody == "" {
		releaseBody, _ = utils.GitGenerateChangelog(
			g.PipelineData.GitLocalPath,
			g.PipelineData.GitBaseInfo.Sha,
			g.PipelineData.GitHeadInfo.Sha,
		)
	}

	//create release.
	ctx := context.Background()
	parts := strings.Split(g.Config.GetString("scm_repo_full_name"), "/")
	version := fmt.Sprintf("v%s", g.PipelineData.ReleaseVersion)
	releaseData, _, rerr := g.Client.Repositories.CreateRelease(
		ctx,
		parts[0],
		parts[1],
		&github.RepositoryRelease{
			TargetCommitish: &releaseSha,
			Body:            &releaseBody,
			TagName:         &version,
			Name:            &version,
		},
	)
	if rerr != nil {
		return rerr
	}

	if perr := g.PublishAssets(releaseData.GetID()); perr != nil {
		log.Print("An error occured while publishing assets:")
		log.Print(perr)
		log.Print("Continuing...")
	}

	return nil
}

func (g *scmGithub) PublishAssets(releaseData interface{}) error {
	//releaseData should be an ID (int)
	releaseId, ok := releaseData.(int)
	if !ok {
		return fmt.Errorf("Invalid releaseID, cannot upload assets")
	}

	ctx := context.Background()
	parts := strings.Split(g.Config.GetString("scm_repo_full_name"), "/")

	for _, assetData := range g.PipelineData.ReleaseAssets {
		publishAsset(
			g.Client,
			ctx,
			parts[0],
			parts[1],
			assetData.ArtifactName,
			path.Join(g.PipelineData.GitLocalPath, assetData.LocalPath),
			releaseId,
			5)
	}
	return nil
}

func (g *scmGithub) Cleanup() error {

	if !g.Config.GetBool("scm_enable_branch_cleanup") { //Default is false, so this will just return without doing anything.
		// - exit if "scm_enable_branch_cleanup" is not true
		return errors.ScmCleanupFailed("scm_enable_branch_cleanup is false. Skipping cleanup")
	} else if !g.PipelineData.IsPullRequest {
		return errors.ScmCleanupFailed("scm cleanup unnecessary for push's. Skipping cleanup")
	} else if g.PipelineData.GitHeadInfo.Repo.FullName != g.PipelineData.GitBaseInfo.Repo.FullName {
		// exit if the HEAD PR branch is not in the same organization and repository as the BASE
		return errors.ScmCleanupFailed("HEAD PR branch is not in the same organization & repo as the BASE. Skipping cleanup")
	}

	ctx := context.Background()
	parts := strings.Split(g.PipelineData.GitBaseInfo.Repo.FullName, "/")

	repoData, _, err := g.Client.Repositories.Get(ctx, parts[0], parts[1])
	if err != nil {
		return err
	}

	if g.PipelineData.GitHeadInfo.Ref == repoData.GetDefaultBranch() || g.PipelineData.GitHeadInfo.Ref == "master" {
		//exit if the HEAD branch is the repo default branch
		//exit if the HEAD branch is master
		return errors.ScmCleanupFailed("HEAD PR branch is default repo branch, or master. Skipping cleanup")
	}

	_, drerr := g.Client.Git.DeleteRef(ctx, parts[0], parts[1], fmt.Sprintf("heads/%s", g.PipelineData.GitHeadInfo.Ref))
	if drerr != nil {
		return drerr
	}

	return nil
}

func (g *scmGithub) Notify(ref string, state string, message string) error {
	targetURL := "https://www.capsulecd.com"
	contextApp := "CapsuleCD"

	ctx := context.Background()
	parts := strings.Split(g.Config.GetString("scm_repo_full_name"), "/")
	_, _, serr := g.Client.Repositories.CreateStatus(ctx, parts[0], parts[1], ref, &github.RepoStatus{
		State:       &state,
		TargetURL:   &targetURL,
		Description: &message,
		Context:     &contextApp,
	})
	return serr
}

//private

func publishAsset(client *github.Client, ctx context.Context, repoOwner string, repoName string, assetName, filePath string, releaseID, retries int) error {

	log.Printf("Attempt (%d) to upload release asset %s from %s", retries, assetName, filePath)
	f, err := os.Open(filePath)
	if err != nil {
		log.Print(err)
		return err
	}

	_, _, err = client.Repositories.UploadReleaseAsset(ctx, repoOwner, repoName, releaseID, &github.UploadOptions{
		Name: assetName,
	}, f)

	if err != nil && retries > 0 {
		fmt.Println("artifact upload errored out, retrying in one second. Err:", err)
		time.Sleep(time.Second)
		err = publishAsset(client, ctx, repoOwner, repoName, assetName, filePath, releaseID, retries-1)
	}

	return err
}

func authGitRemote(cloneUrl string, accessToken string) (string, error) {
	if accessToken != "" {
		// set the remote url, with embedded token
		u, err := url.Parse(cloneUrl)
		if err != nil {
			return "", err
		}
		u.User = url.UserPassword(accessToken, "")
		return u.String(), nil
	} else {
		return cloneUrl, nil
	}
}
