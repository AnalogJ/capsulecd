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
)

type scmGithub struct {
	options *ScmOptions
	client  *github.Client
}

// configure method will generate an authenticated client that can be used to comunicate with Github
// MUST set options.GitParentPath
// MUST set client
func (g *scmGithub) Configure() {

	g.options = new(ScmOptions)

	if !config.IsSet("scm_github_access_token") {
		log.Fatal("Missing github access token")
		return
	}
	if config.IsSet("scm_git_parent_path") {
		g.options.GitParentPath = config.GetString("scm_git_parent_path")
		os.MkdirAll(g.options.GitParentPath, os.ModePerm)
	} else {
		dirPath, err := ioutil.TempDir("","")
		if err != nil {
			log.Fatal("Could not create Temp Directory for scm checkout.", err)
			return
		}
		g.options.GitParentPath = dirPath
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.GetString("scm_github_access_token")},
	)
	tc := oauth2.NewClient(ctx, ts)

	//TODO: autopaginate turned on.
	//TODO: add support for alternative api endpoints "scm_github_api_endpoint"
	g.client = github.NewClient(tc)
	return
}


// configure method will retrieve payload data from Scm using authenticated client.
// MUST set options.IsPullRequest
// RETURNS ScmPayload
func (g *scmGithub) RetrievePayload() *ScmPayload {
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
		}
		//make this as similar to a pull request as possible
	} else {
		g.options.IsPullRequest = true
		ctx := context.Background()
		parts := strings.Split(config.GetString("scm_repo_full_name"), "/")
		pr, _, err := g.client.PullRequests.Get(ctx, parts[0],parts[1], config.GetInt("scm_pull_request"))

		if(err != nil){
			log.Fatal("Could not retrieve pull request from Github", err)
			return nil
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
		}
	}
}


// all capsule CD processing will be kicked off via a payload. In Github's case, the payload is the webhook data.
// should check if the pull request opener even has permissions to create a release.
// all sources should process the payload by downloading a git repository that contains the master branch merged with the test branch
// MUST set options.GitLocalPath
// MUST set options.GitLocalBranch
// MUST set options.itHeadInfo
// REQUIRES options.GitParentPath
func (g *scmGithub) ProcessPushPayload() {
	return
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
func (g *scmGithub) ProcessPullRequestPayload() {
	return
}

func (g *scmGithub) Publish() {
	return
}

func (g *scmGithub) Notify() {
	return
}

func (g *scmGithub) Options() *ScmOptions {
	log.Print("ORINT THE PARENT PATH", g.options)
	return g.options
}