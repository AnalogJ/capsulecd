package lib

import (
	"capsulecd/lib/config"
	"capsulecd/lib/scm"
	"capsulecd/lib/engine"
	"capsulecd/lib/errors"
	"path"
	"fmt"
)

type Pipeline struct {
	SourceScm *scm.Scm
	Engine *engine.Engine
}

func (p *Pipeline) Start(){
	//Initialize Configuration not already initialized.
	if(!config.IsInitialized()){
		config.Init()
	}

	//Generate a new instance of the sourceScm
	scmImpl, serr := scm.Create()
	errors.CheckErr(serr)
	p.SourceScm = &scmImpl

	//Generate a new instance of the engine
	engineImpl, eerr := engine.Create()
	errors.CheckErr(eerr)
	p.Engine = &engineImpl
	engineImpl.Init(&scmImpl)

	p.PreValidateTools()
	engineImpl.ValidateTools()
	p.PostValidateTools()

	// start the source, and whatever work needs to be done there.
	// MUST set options.GitParentPath
	// MUST set options.Client
	p.PreScmInit()
	scmImpl.Init(nil)
	p.PostScmInit()

	// runner must determine if this is a pull request or a push.
	// if it's a pull request the runner must retrieve the pull request payload and return it
	// if its a push, the runner must retrieve the push payload and return it
	// the variable @runner_is_pullrequest will be true if a pull request was created.
	// MUST set runner_is_pullrequest
	// REQUIRES source_client
	p.PreScmRetrievePayload()
	payload, _ := scmImpl.RetrievePayload()
	p.PostScmRetrievePayload()

	if scmImpl.Options().IsPullRequest {
		// all capsule CD processing will be kicked off via a payload. In this case the payload is the pull request data.
		// should check if the pull request opener even has permissions to create a release.
		// all sources should process the payload by downloading a git repository that contains the master branch merged with the test branch
		// MUST set source_git_local_path
		// MUST set source_git_local_branch
		// MUST set source_git_base_info
		// MUST set source_git_head_info
		// REQUIRES source_client
		p.PreScmProcessPullRequestPayload()
		scmImpl.ProcessPullRequestPayload(payload)
		p.PostScmProcessPullRequestPayload()
	} else {
		// start processing the payload, which should result in a local git repository that we
		// can begin to test. Since this is a push, no packaging is required
		// MUST set source_git_local_path
		// MUST set source_git_local_branch
		// MUST set source_git_head_info
		// REQUIRES source_client
		p.PreScmProcessPushPayload()
		scmImpl.ProcessPushPayload(payload)
		p.PostScmProcessPushPayload()
	}

	// update the config with repo config file options
	config.ReadConfig(path.Join(scmImpl.Options().GitLocalPath, "capsule.yml"))

	// now that the payload has been processed we can begin by building the code.
	// this may be creating missing files/default structure, compilation, version bumping, etc.
	p.NotifyStep("build", func() error {
		p.PreBuildStep()
		engineImpl.BuildStep()
		p.PostBuildStep()
		return nil
	})

	// this step should download dependencies, run the package test runner(s) (eg. npm test, rake test, kitchen test)
	// REQUIRES @config.engine_cmd_test
	// REQUIRES @config.engine_disable_test
	p.NotifyStep("test", func() error {
		p.PreTestStep()
		engineImpl.TestStep()
		p.PostTestStep()
		return nil
	})

	// this step should commit any local changes and create a git tag. Nothing should be pushed to remote repository
	p.NotifyStep("package", func() error {
		p.PrePackageStep()
		engineImpl.PackageStep()
		p.PostPackageStep()
		return nil
	})

	if scmImpl.Options().IsPullRequest {
		// this step should push the release to the package repository (ie. npm, chef supermarket, rubygems)
		p.NotifyStep("release", func() error {
			p.PreReleaseStep()
			engineImpl.ReleaseStep()
			p.PostReleaseStep()
			return nil
		})

		// this step should push the merged, tested and version updated code up to the source code repository
		// this step should also do any source specific releases (github release, asset uploading, etc)
		p.NotifyStep("scm publish", func() error {
			p.PreScmRelease()
			scmImpl.Publish()
			p.PostScmRelease()
			return nil
		})
	}
}

// Hook methods
func (p *Pipeline) PreValidateTools(){}
func (p *Pipeline) PostValidateTools(){}
func (p *Pipeline) PreScmInit(){}
func (p *Pipeline) PostScmInit(){}
func (p *Pipeline) PreScmProcessPullRequestPayload(){}
func (p *Pipeline) PostScmProcessPullRequestPayload(){}
func (p *Pipeline) PreScmProcessPushPayload(){}
func (p *Pipeline) PostScmProcessPushPayload(){}
func (p *Pipeline) PreScmRelease(){}
func (p *Pipeline) PostScmRelease(){}
func (p *Pipeline) PreScmRetrievePayload(){}
func (p *Pipeline) PostScmRetrievePayload(){}
func (p *Pipeline) PreBuildStep(){}
func (p *Pipeline) PostBuildStep(){}
func (p *Pipeline) PreTestStep(){}
func (p *Pipeline) PostTestStep(){}
func (p *Pipeline) PrePackageStep(){}
func (p *Pipeline) PostPackageStep(){}
func (p *Pipeline) PreReleaseStep(){}
func (p *Pipeline) PostReleaseStep(){}

func (p *Pipeline) NotifyStep(step string, callback func() error){
	(*p.SourceScm).Notify((*p.SourceScm).Options().GitHeadInfo.Sha, "pending", fmt.Sprintf("Started '%s' step. Pull request will be merged automatically when complete.", step))
	cerr := callback()
	if(cerr != nil){
		//TODO: remove the temp folder path.
		(*p.SourceScm).Notify((*p.SourceScm).Options().GitHeadInfo.Sha, "failure", fmt.Sprintf("Error: '%s'", cerr))
	}
}

//type NotifyStepCallback func() error

