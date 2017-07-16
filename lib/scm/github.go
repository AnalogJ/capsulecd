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
)

type scmGithub struct {
	options *ScmOptions
	client  *github.Client
}

// configure method will generate an authenticated client that can be used to comunicate with Github
// MUST set options.GitParentPath
// MUST set client
func (g *scmGithub) Configure(client *http.Client) (error) {

	g.options = new(ScmOptions)

	if !config.IsSet("scm_github_access_token") {
		return errors.ScmAuthenticationFailed("Missing github access token")
	}
	if config.IsSet("scm_git_parent_path") {
		g.options.GitParentPath = config.GetString("scm_git_parent_path")
		os.MkdirAll(g.options.GitParentPath, os.ModePerm)
	} else {
		dirPath, err := ioutil.TempDir("","")
		if err != nil {
			return err
		}
		g.options.GitParentPath = dirPath
	}

	if(client != nil){
		//primarily used for testing.
		g.client = github.NewClient(client)
	} else {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: config.GetString("scm_github_access_token")},
		)
		tc := oauth2.NewClient(ctx, ts)

		//TODO: autopaginate turned on.
		//TODO: add support for alternative api endpoints "scm_github_api_endpoint"
		g.client = github.NewClient(tc)
	}

	return nil
}


// configure method will retrieve payload data from Scm using authenticated client.
// MUST set options.IsPullRequest
// RETURNS ScmPayload
func (g *scmGithub) RetrievePayload() (*ScmPayload, error) {
	if !config.IsSet("scm_pull_request") {
		log.Print("This is not a pull request. No automatic continuous deployment processing required. Continuous Integration testing will continue.")
		g.options.IsPullRequest = false

		return &ScmPayload{
			Head: &ScmCommitInfo{
				Sha: config.GetString("scm_sha"),
				Ref: config.GetString("scm_branch"),
				Repo: &ScmRepoInfo{
					CloneUrl: config.GetString("scm_clone_url"),
					Name: config.GetString("scm_repo_name"),
					FullName: config.GetString("scm_repo_full_name"),
				},
			},
		}, nil
		//make this as similar to a pull request as possible
	} else {
		g.options.IsPullRequest = true
		ctx := context.Background()
		parts := strings.Split(config.GetString("scm_repo_full_name"), "/")
		pr, _, err := g.client.PullRequests.Get(ctx, parts[0],parts[1], config.GetInt("scm_pull_request"))

		if(err != nil){
			return nil, errors.ScmAuthenticationFailed(fmt.Sprintf("Could not retrieve pull request from Github: %s", err))
		}

		return &ScmPayload{
			Title: pr.GetTitle(),
			Head: &ScmCommitInfo{
				Sha: pr.Head.GetSHA(),
				Ref: pr.Head.GetRef(),
				Repo: &ScmRepoInfo{
					CloneUrl: pr.Head.Repo.GetCloneURL(),
					Name: pr.Head.Repo.GetName(),
					FullName: pr.Head.Repo.GetFullName(),
				},
			},
			Base: &ScmCommitInfo{
				Sha: pr.Base.GetSHA(),
				Ref: pr.Base.GetRef(),
				Repo: &ScmRepoInfo{
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
	g.options.GitHeadInfo = payload.Head
	err := g.options.GitHeadInfo.Validate()
	if(err != nil){
		return err
	}

	if(config.IsSet("scm_github_access_token")){
		// set the remote url, with embedded token
		u, err := url.Parse(g.options.GitHeadInfo.Repo.CloneUrl)
		if err != nil {
			return err
		}

		u.User = url.UserPassword(config.GetString("scm_github_access_token"), "")
		g.options.GitRemote  = u.String()
	} else {
		g.options.GitRemote = g.options.GitHeadInfo.Repo.CloneUrl
	}

	g.options.GitLocalBranch = g.options.GitHeadInfo.Ref

	// clone the merged branch
	// https://sethvargo.com/checkout-a-github-pull-request/
	// https://coderwall.com/p/z5rkga/github-checkout-a-pull-request-as-a-branch

	gitLocalPath, cerr := utils.GitClone(g.options.GitParentPath, g.options.GitHeadInfo.Repo.Name, g.options.GitRemote)
	if(cerr != nil){return cerr}
	g.options.GitLocalPath = gitLocalPath

	return utils.GitCheckout(g.options.GitLocalPath, g.options.GitHeadInfo.Ref)
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
	return nil
}

func (g *scmGithub) Publish() error {
	return nil
}

func (g *scmGithub) Notify() error {
	return nil
}

func (g *scmGithub) Options() *ScmOptions {
	log.Print("ORINT THE PARENT PATH", g.options)
	return g.options
}