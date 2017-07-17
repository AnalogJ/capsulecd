package scm

import (
	"github.com/google/go-github/github"
	"context"
	"golang.org/x/oauth2"
	"capsulecd/lib/config"
	"log"
	"os"
	"io/ioutil"
	"strings"
	"net/http"
	"capsulecd/lib/errors"
	"fmt"
	"net/url"
	"capsulecd/lib/utils"
	"strconv"
	"time"
	"capsulecd/lib/pipeline"
)

type scmGithub struct {
	PipelineData *pipeline.PipelineData
	Client       *github.Client
}

// configure method will generate an authenticated client that can be used to comunicate with Github
// MUST set options.GitParentPath
// MUST set client
func (g *scmGithub) Init(pipelineData *pipeline.PipelineData, client *http.Client) (error) {

	g.PipelineData = pipelineData

	if !config.IsSet("scm_github_access_token") {
		return errors.ScmAuthenticationFailed("Missing github access token")
	}
	if config.IsSet("scm_git_parent_path") {
		g.PipelineData.GitParentPath = config.GetString("scm_git_parent_path")
		os.MkdirAll(g.PipelineData.GitParentPath, os.ModePerm)
	} else {
		dirPath, err := ioutil.TempDir("","")
		if err != nil {
			return err
		}
		g.PipelineData.GitParentPath = dirPath
	}

	if(client != nil){
		//primarily used for testing.
		g.Client = github.NewClient(client)
	} else {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: config.GetString("scm_github_access_token")},
		)
		tc := oauth2.NewClient(ctx, ts)

		//TODO: autopaginate turned on.
		//TODO: add support for alternative api endpoints "scm_github_api_endpoint"
		g.Client = github.NewClient(tc)
	}

	return nil
}


// configure method will retrieve payload data from Scm using authenticated client.
// MUST set options.IsPullRequest
// RETURNS ScmPayload
func (g *scmGithub) RetrievePayload() (*ScmPayload, error) {
	if !config.IsSet("scm_pull_request") {
		log.Print("This is not a pull request. No automatic continuous deployment processing required. Continuous Integration testing will continue.")
		g.PipelineData.IsPullRequest = false

		return &ScmPayload{
			Head: &pipeline.PipelineScmCommitInfo{
				Sha: config.GetString("scm_sha"),
				Ref: config.GetString("scm_branch"),
				Repo: &pipeline.PipelineScmRepoInfo{
					CloneUrl: config.GetString("scm_clone_url"),
					Name: config.GetString("scm_repo_name"),
					FullName: config.GetString("scm_repo_full_name"),
				},
			},
		}, nil
		//make this as similar to a pull request as possible
	} else {
		g.PipelineData.IsPullRequest = true
		ctx := context.Background()
		parts := strings.Split(config.GetString("scm_repo_full_name"), "/")
		pr, _, err := g.Client.PullRequests.Get(ctx, parts[0],parts[1], config.GetInt("scm_pull_request"))

		if(err != nil){
			return nil, errors.ScmAuthenticationFailed(fmt.Sprintf("Could not retrieve pull request from Github: %s", err))
		}

		//validate pullrequest
		if(pr.GetState() != "open"){
			return nil, errors.ScmPayloadUnsupported("Pull request has an invalid action")
		}
		if(pr.Base.Repo.GetDefaultBranch() != pr.Base.GetRef()){
			return nil, errors.ScmPayloadUnsupported(fmt.Sprintf("Pull request is not being created against the default branch of this repository (%s vs %s)", pr.Base.Repo.GetDefaultBranch(), pr.Base.GetRef() ))
		}
		// check the payload push user.

		//TODO: figure out how to do optional authenication. possible options, Source USER, token based auth, no auth when used with capsulecd.com.
		// unless @source_client.collaborator?(payload['base']['repo']['full_name'], payload['user']['login'])
		//
		//   @source_client.add_comment(payload['base']['repo']['full_name'], payload['number'], CapsuleCD::BotUtils.pull_request_comment)
		//   fail CapsuleCD::Error::SourceUnauthorizedUser, 'Pull request was opened by an unauthorized user'
        	// end

		return &ScmPayload{
			Title: pr.GetTitle(),
			PullRequestNumber: strconv.Itoa(pr.GetNumber()),
			Head: &pipeline.PipelineScmCommitInfo{
				Sha: pr.Head.GetSHA(),
				Ref: pr.Head.GetRef(),
				Repo: &pipeline.PipelineScmRepoInfo{
					CloneUrl: pr.Head.Repo.GetCloneURL(),
					Name: pr.Head.Repo.GetName(),
					FullName: pr.Head.Repo.GetFullName(),
				},
			},
			Base: &pipeline.PipelineScmCommitInfo{
				Sha: pr.Base.GetSHA(),
				Ref: pr.Base.GetRef(),
				Repo: &pipeline.PipelineScmRepoInfo{
					CloneUrl: pr.Base.Repo.GetCloneURL(),
					Name: pr.Base.Repo.GetName(),
					FullName: pr.Base.Repo.GetFullName(),
				},
			},
		}, nil
	}
}


// all capsule CD processing will be kicked off via a payload. In Github's case, the payload is the webhook data.
// should check if the pull request opener even has permissions to create a release.
// all sources should process the payload by downloading a git repository that contains the master branch merged with the test branch
// MUST set options.GitLocalPath
// MUST set options.GitLocalBranch
// MUST set options.GitHeadInfo
// REQUIRES options.GitParentPath
func (g *scmGithub) ProcessPushPayload(payload *ScmPayload) error {
	//set the processed head info
	g.PipelineData.GitHeadInfo = payload.Head
	err := g.PipelineData.GitHeadInfo.Validate()
	if(err != nil){
		return err
	}

	authRemote, aerr := authGitRemote(g.PipelineData.GitHeadInfo.Repo.CloneUrl, config.GetString("scm_github_access_token"))
	if(aerr != nil){
		return aerr
	}
	g.PipelineData.GitRemote = authRemote
	g.PipelineData.GitLocalBranch = g.PipelineData.GitHeadInfo.Ref

	// clone the merged branch
	// https://sethvargo.com/checkout-a-github-pull-request/
	// https://coderwall.com/p/z5rkga/github-checkout-a-pull-request-as-a-branch

	gitLocalPath, cerr := utils.GitClone(g.PipelineData.GitParentPath, g.PipelineData.GitHeadInfo.Repo.Name, g.PipelineData.GitRemote)
	if(cerr != nil){return cerr}
	g.PipelineData.GitLocalPath = gitLocalPath

	return utils.GitCheckout(g.PipelineData.GitLocalPath, g.PipelineData.GitHeadInfo.Ref)
}

// all capsule CD processing will be kicked off via a payload. In Github's case, the payload is the pull request data.
// should check if the pull request opener even has permissions to create a release.
// all sources should process the payload by downloading a git repository that contains the master branch merged with the test branch
// MUST set options.GitLocalPath
// MUST set options.GitLocalBranch
// MUST set options.GitBaseInfo
// MUST set options.GitHeadInfo
// REQUIRES client
// REQUIRES options.GitParentPath
func (g *scmGithub) ProcessPullRequestPayload(payload *ScmPayload) error {
	//set the processed head info
	g.PipelineData.GitHeadInfo = payload.Head
	g.PipelineData.GitBaseInfo = payload.Base
	herr := g.PipelineData.GitHeadInfo.Validate()
	berr := g.PipelineData.GitBaseInfo.Validate()
	if(herr != nil){
		return herr
	} else if(berr != nil){
		return berr
	}


	authRemote, aerr := authGitRemote(g.PipelineData.GitHeadInfo.Repo.CloneUrl, config.GetString("scm_github_access_token"))
	if(aerr != nil){
		return aerr
	}
	g.PipelineData.GitRemote = authRemote

	// clone the merged branch
	// https://sethvargo.com/checkout-a-github-pull-request/
	// https://coderwall.com/p/z5rkga/github-checkout-a-pull-request-as-a-branch

	gitLocalPath, cerr := utils.GitClone(g.PipelineData.GitParentPath, g.PipelineData.GitHeadInfo.Repo.Name, g.PipelineData.GitRemote)
	if(cerr != nil){return cerr}
	g.PipelineData.GitLocalPath = gitLocalPath
	g.PipelineData.GitLocalBranch = fmt.Sprintf("pr_%s",payload.PullRequestNumber)

	ferr := utils.GitFetch(g.PipelineData.GitLocalPath, fmt.Sprintf("refs/pull/%s/merge", payload.PullRequestNumber), g.PipelineData.GitLocalBranch)
	if(ferr != nil){return ferr}

	//return utils.GitCheckout(g.options.GitLocalPath, g.options.GitLocalBranch)
	//
	// show a processing message on the github PR.
	g.Notify(g.PipelineData.GitHeadInfo.Sha, "pending", "Started processing package. Pull request will be merged automatically when complete.")
	return nil;
}

// REQUIRES client
// REQUIRES options.ScmReleaseCommit
// REQUIRES options.GitLocalPath
// REQUIRES options.GitLocalBranch
// REQUIRES options.GitBaseInfo
// REQUIRES options.GitHeadInfo
// REQUIRES options.ReleaseArtifacts
// REQUIRES options.GitParentPath
func (g *scmGithub) Publish() error {

	// set the pull request status (we do this before the merge, because we cant update status on a merged
	//PR anways. If the push fails, the status will be set to error correctly.
	return g.Notify(
		g.PipelineData.GitBaseInfo.Repo.FullName,
		"success",
		"Pull-request was successfully merged, new release created.",
	)

	// push the version bumped metadata file + newly created files to
	perr := utils.GitPush(g.PipelineData.GitLocalPath, g.PipelineData.GitLocalBranch, g.PipelineData.GitBaseInfo.Ref)
	if(perr != nil){ return perr }
	//sleep because github needs time to process the new tag.
	time.Sleep(5 * time.Second)

	// calculate teh relaese sha
	releaseSha := utils.LeftPad2Len(g.PipelineData.ReleaseCommit, "0", 40)

	//get the release changelog
	releaseBody, clerr := utils.GitGenerateChangelog(
		g.PipelineData.GitLocalPath,
		g.PipelineData.GitBaseInfo.Sha,
		g.PipelineData.GitHeadInfo.Sha,
		g.PipelineData.GitBaseInfo.Repo.FullName,
	)
	if(clerr != nil){
		return clerr
	}

	//create release.
	ctx := context.Background()
	parts := strings.Split(config.GetString("scm_repo_full_name"), "/")
	version := fmt.Sprintf("v%s", g.PipelineData.ReleaseVersion)
	g.Client.Repositories.CreateRelease(
		ctx,
		parts[0],
		parts[1],
		&github.RepositoryRelease{
			TargetCommitish: &releaseSha,
			Body: &releaseBody,
			TagName: &version,
			Name: &version,
		},
	)

	//TODO: upload artifacts
	//@source_release_artifacts.each do |release_artifact|
	//	@source_client.upload_asset(release[:url], release_artifact[:path], name: release_artifact[:name])
	//end
	//
	os.RemoveAll(g.PipelineData.GitParentPath)
	return nil

}

// requires @source_client
// requires @source_git_parent_path
// requires @source_git_base_info
// requires @source_git_head_info
// requires @config.engine_disable_cleanup
func (g *scmGithub) Notify(ref string, state string, message string) error {

	targetURL := "https://www.capsulecd.com"
	contextApp := "CapsuleCD"

	ctx := context.Background()
	parts := strings.Split(config.GetString("scm_repo_full_name"), "/")
	_, _, serr := g.Client.Repositories.CreateStatus(ctx, parts[0], parts[1], ref, &github.RepoStatus{
		State: &state,
		TargetURL: &targetURL,
		Description: &message,
		Context: &contextApp,
	})
	return serr
}

//private

func authGitRemote(cloneUrl string, accessToken string) (string, error) {
	if(accessToken != ""){
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
