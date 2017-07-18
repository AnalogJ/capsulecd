package lib

import (
	"capsulecd/lib/config"
	"capsulecd/lib/engine"
	"capsulecd/lib/errors"
	"capsulecd/lib/pipeline"
	"capsulecd/lib/scm"
	"fmt"
	"log"
	"path"
	"os"
)

type Pipeline struct {
	Data   *pipeline.Data
	Scm    scm.Scm
	Engine engine.Engine
}

func (p *Pipeline) Start() {
	defer p.Cleanup()
	//Initialize Configuration not already initialized.
	if !config.IsInitialized() {
		log.Printf("Configuration is not initialized, doing it now.")
		config.Init()
	}

	p.Data = new(pipeline.Data)

	//Generate a new instance of the sourceScm
	scmImpl, serr := scm.Create()
	errors.CheckErr(serr)
	p.Scm = scmImpl

	//Generate a new instance of the engine
	engineImpl, eerr := engine.Create()
	errors.CheckErr(eerr)
	p.Engine = engineImpl
	errors.CheckErr(engineImpl.Init(p.Data, scmImpl))

	// start the source, and whatever work needs to be done there.
	// MUST set options.GitParentPath
	// MUST set options.Client
	log.Println("pre_scm_init_step")
	p.PreScmInit()
	log.Println("scm_init_step")
	errors.CheckErr(scmImpl.Init(p.Data, nil))
	log.Println("post_scm_init_step")
	p.PostScmInit()

	// runner must determine if this is a pull request or a push.
	// if it's a pull request the runner must retrieve the pull request payload and return it
	// if its a push, the runner must retrieve the push payload and return it
	// the variable @runner_is_pullrequest will be true if a pull request was created.
	// MUST set runner_is_pullrequest
	// REQUIRES source_client
	log.Println("pre_scm_retrieve_payload_step")
	p.PreScmRetrievePayload()
	log.Println("scm_retrieve_payload_step")
	payload, perr := scmImpl.RetrievePayload()
	errors.CheckErr(perr)
	log.Println("post_scm_retrieve_payload_step")
	p.PostScmRetrievePayload()

	if p.Data.IsPullRequest {
		// all capsule CD processing will be kicked off via a payload. In this case the payload is the pull request data.
		// should check if the pull request opener even has permissions to create a release.
		// all sources should process the payload by downloading a git repository that contains the master branch merged with the test branch
		// MUST set source_git_local_path
		// MUST set source_git_local_branch
		// MUST set source_git_base_info
		// MUST set source_git_head_info
		// REQUIRES source_client
		log.Println("pre_scm_process_pull_request_step")
		p.PreScmProcessPullRequestPayload()
		log.Println("scm_process_pull_request_step")
		errors.CheckErr(scmImpl.ProcessPullRequestPayload(payload))
		log.Println("post_scm_process_pull_request_step")
		p.PostScmProcessPullRequestPayload()
	} else {
		// start processing the payload, which should result in a local git repository that we
		// can begin to test. Since this is a push, no packaging is required
		// MUST set source_git_local_path
		// MUST set source_git_local_branch
		// MUST set source_git_head_info
		// REQUIRES source_client
		log.Println("pre_scm_process_push_payload_step")
		p.PreScmProcessPushPayload()
		log.Println("scm_process_push_payload_step")
		errors.CheckErr(scmImpl.ProcessPushPayload(payload))
		log.Println("post_scm_process_push_payload_step")
		p.PostScmProcessPushPayload()
	}

	// update the config with repo config file options
	config.ReadConfig(path.Join(p.Data.GitLocalPath, "capsule.yml"))

	//validate that required executables are available for the following build/test/package/etc steps
	p.NotifyStep("validate tools", func() error {
		log.Println("pre_validate_tools_step")
		p.PreValidateTools()
		log.Println("validate_tools_step")
		if verr := engineImpl.ValidateTools(); verr != nil{
			return verr
		}
		log.Println("post_validate_tools_step")
		p.PostValidateTools()
		return nil
	})

	// now that the payload has been processed we can begin by building the code.
	// this may be creating missing files/default structure, compilation, version bumping, etc.
	p.NotifyStep("build", func() error {
		log.Println("pre_build_step")
		p.PreBuildStep()
		log.Println("build_step")
		if berr := engineImpl.BuildStep(); berr!= nil{
			return berr
		}
		log.Println("post_build_step")
		p.PostBuildStep()
		return nil
	})

	// this step should download dependencies, run the package test runner(s) (eg. npm test, rake test, kitchen test)
	// REQUIRES @config.engine_cmd_test
	// REQUIRES @config.engine_disable_test
	p.NotifyStep("test", func() error {
		//skip the test command if disabled
		if config.GetBool("engine_disable_test") {
			return nil
		}

		log.Println("pre_test_step")
		p.PreTestStep()
		log.Println("test_step")
		if terr := engineImpl.TestStep(); terr != nil {
			return terr;
		}
		log.Println("post_test_step")
		p.PostTestStep()
		return nil
	})

	// this step should commit any local changes and create a git tag. Nothing should be pushed to remote repository
	p.NotifyStep("package", func() error {
		log.Println("pre_package_step")
		p.PrePackageStep()
		log.Println("package_step")
		if perr := engineImpl.PackageStep(); perr != nil {
			return perr
		}
		log.Println("post_package_step")
		p.PostPackageStep()
		return nil
	})

	if p.Data.IsPullRequest {
		// this step should push the release to the package repository (ie. npm, chef supermarket, rubygems)
		p.NotifyStep("dist", func() error {
			log.Println("pre_dist_step")
			p.PreDistStep()
			log.Println("dist_step")
			if derr := engineImpl.DistStep(); derr != nil {
				return derr
			}
			log.Println("post_dist_step")
			p.PostDistStep()
			return nil
		})

		// this step should push the merged, tested and version updated code up to the source code repository
		// this step should also do any source specific releases (github release, asset uploading, etc)
		p.NotifyStep("scm publish", func() error {
			log.Println("pre_scm_publish_step")
			p.PreScmPublish()
			log.Println("scm_publish_step")
			if serr := scmImpl.Publish(); serr != nil{
				return serr
			}
			log.Println("post_scm_publish_step")
			p.PostScmPublish()
			return nil
		})
	}
}

// Hook methods
func (p *Pipeline) PreValidateTools()                 {}
func (p *Pipeline) PostValidateTools()                {}
func (p *Pipeline) PreScmInit()                       {}
func (p *Pipeline) PostScmInit()                      {}
func (p *Pipeline) PreScmProcessPullRequestPayload()  {}
func (p *Pipeline) PostScmProcessPullRequestPayload() {}
func (p *Pipeline) PreScmProcessPushPayload()         {}
func (p *Pipeline) PostScmProcessPushPayload()        {}
func (p *Pipeline) PreScmPublish()                    {}
func (p *Pipeline) PostScmPublish()                   {}
func (p *Pipeline) PreScmRetrievePayload()            {}
func (p *Pipeline) PostScmRetrievePayload()           {}
func (p *Pipeline) PreBuildStep()                     {}
func (p *Pipeline) PostBuildStep()                    {}
func (p *Pipeline) PreTestStep()                      {}
func (p *Pipeline) PostTestStep()                     {}
func (p *Pipeline) PrePackageStep()                   {}
func (p *Pipeline) PostPackageStep()                  {}
func (p *Pipeline) PreDistStep()                      {}
func (p *Pipeline) PostDistStep()                     {}

func (p *Pipeline) NotifyStep(step string, callback func() error) {
	p.Scm.Notify(p.Data.GitHeadInfo.Sha, "pending", fmt.Sprintf("Started '%s' step. Pull request will be merged automatically when complete.", step))
	cerr := callback()
	if cerr != nil {
		//TODO: remove the temp folder path.
		p.Scm.Notify(p.Data.GitHeadInfo.Sha, "failure", fmt.Sprintf("Error: '%s'", cerr))
		p.Cleanup()
		errors.CheckErr(cerr)
	}
}

func (p *Pipeline) Cleanup(){
	log.Println("Running Cleanup...")
	if p.Data != nil && p.Data.GitParentPath != "" {
		os.RemoveAll(p.Data.GitParentPath)
	}
}

//type NotifyStepCallback func() error
