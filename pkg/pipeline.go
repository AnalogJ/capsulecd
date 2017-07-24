package pkg

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/engine"
	"capsulecd/pkg/errors"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/scm"
	"fmt"
	"log"
	"os"
	"path"
)

type Pipeline struct {
	Data   *pipeline.Data
	Config config.Interface
	Scm    scm.Interface
	Engine engine.Interface
}

func (p *Pipeline) Start(config config.Interface) {
	defer p.Cleanup()

	p.Config = config
	p.Data = new(pipeline.Data)

	// start the source, and whatever work needs to be done there.
	// MUST set options.GitParentPath
	log.Println("pre_scm_init_step")
	p.PreScmInit()
	log.Println("scm_init_step")
	scmImpl, serr := scm.Create(p.Config.GetString("scm"), p.Data, config, nil)
	errors.CheckErr(serr)
	p.Scm = scmImpl
	log.Println("post_scm_init_step")
	p.PostScmInit()

	//Generate a new instance of the engine
	engineImpl, eerr := engine.Create(p.Config.GetString("package_type"), p.Data, config, scmImpl)
	errors.CheckErr(eerr)
	p.Engine = engineImpl

	//retreive payload
	log.Println("pre_scm_retrieve_payload_step")
	p.PreScmRetrievePayload()
	log.Println("scm_retrieve_payload_step")
	payload, perr := scmImpl.RetrievePayload()
	errors.CheckErr(perr)
	log.Println("post_scm_retrieve_payload_step")
	p.PostScmRetrievePayload()

	if p.Data.IsPullRequest {
		log.Println("pre_scm_checkout_pull_request_step")
		p.PreScmCheckoutPullRequestPayload()
		log.Println("scm_checkout_pull_request_step")
		errors.CheckErr(scmImpl.CheckoutPullRequestPayload(payload))
		log.Println("post_scm_checkout_pull_request_step")
		p.PostScmCheckoutPullRequestPayload()
	} else {
		log.Println("pre_scm_checkout_push_payload_step")
		p.PreScmCheckoutPushPayload()
		log.Println("scm_checkout_push_payload_step")
		errors.CheckErr(scmImpl.CheckoutPushPayload(payload))
		log.Println("post_scm_checkout_push_payload_step")
		p.PostScmCheckoutPushPayload()
	}

	// update the config with repo config file options
	p.Config.ReadConfig(path.Join(p.Data.GitLocalPath, "capsule.yml"))

	if p.Config.IsSet("scm_release_assets") {
		//unmarshall config data.
		parsedAssets := new([]pipeline.ScmReleaseAsset)
		err := p.Config.UnmarshalKey("scm_release_assets", parsedAssets)
		errors.CheckErr(err)

		//append the parsed Assets to the current ReleaseAssets storage (incase assets were defined in system yml)
		p.Data.ReleaseAssets = append(p.Data.ReleaseAssets, (*parsedAssets)...)
	}

	//validate that required executables are available for the following build/test/package/etc steps
	p.NotifyStep("validate tools", func() error {
		log.Println("validate_tools_step")
		if verr := engineImpl.ValidateTools(); verr != nil {
			return verr
		}
		return nil
	})

	// now that the payload has been processed we can begin by building the code.
	// this may be creating missing files/default structure, compilation, version bumping, etc.
	p.NotifyStep("assemble", func() error {
		log.Println("pre_assemble_step")
		p.PreAssembleStep()
		log.Println("assemble_step")
		if berr := engineImpl.AssembleStep(); berr != nil {
			return berr
		}
		log.Println("post_assemble_step")
		p.PostAssembleStep()
		return nil
	})

	// this step should download dependencies
	p.NotifyStep("dependencies", func() error {
		log.Println("pre_dependencies_step")
		p.PreDependenciesStep()
		log.Println("dependencies_step")
		if berr := engineImpl.DependenciesStep(); berr != nil {
			return berr
		}
		log.Println("post_dependencies_step")
		p.PostDependenciesStep()
		return nil
	})

	// this step should compile source
	p.NotifyStep("compile", func() error {
		log.Println("pre_compile_step")
		p.PreCompileStep()
		log.Println("compile_step")
		if berr := engineImpl.CompileStep(); berr != nil {
			return berr
		}
		log.Println("post_compile_step")
		p.PostCompileStep()
		return nil
	})

	// run the package test runner(s) (eg. npm test, rake test, kitchen test) and linters/formatters
	p.NotifyStep("test", func() error {
		log.Println("pre_test_step")
		p.PreTestStep()
		log.Println("test_step")
		if terr := engineImpl.TestStep(); terr != nil {
			return terr
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
			if p.Config.GetBool("engine_disable_dist") {
				log.Println("skipping pre_dist_step, dist_step, post_dist_step")
				return nil
			}

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

		p.NotifyStep("scm publish", func() error {
			if p.Config.GetBool("scm_disable_publish") {
				log.Println("skipping pre_scm_publish_step, scm_publish_step, post_scm_publish_step")
				return nil
			}

			log.Println("pre_scm_publish_step")
			p.PreScmPublish()
			log.Println("scm_publish_step")
			if serr := scmImpl.Publish(); serr != nil {
				return serr
			}
			log.Println("post_scm_publish_step")
			p.PostScmPublish()
			return nil
		})

		p.NotifyStep("scm cleanup", func() error {
			log.Println("pre_scm_cleanup_step")
			p.PreScmCleanup()
			log.Println("scm_cleanup_step")
			if serr := scmImpl.Cleanup(); serr != nil {
				//if theres an error, just print it (cleanup will not fail the deployment)
				// it will stop PostSCM cleanup from running.
				log.Print(serr)
				return nil
			}

			log.Println("post_scm_cleanup_step")
			p.PostScmCleanup()
			return nil
		})

		//if there was an error, it should not have gotten to this point. CheckErr panic's
		scmImpl.Notify(
			p.Data.GitHeadInfo.Sha,
			"success",
			"Pull-request was successfully merged, new release created.",
		)
	}
}

// Hook methods
func (p *Pipeline) PreScmInit()                        {}
func (p *Pipeline) PostScmInit()                       {}
func (p *Pipeline) PreScmCheckoutPullRequestPayload()  {}
func (p *Pipeline) PostScmCheckoutPullRequestPayload() {}
func (p *Pipeline) PreScmCheckoutPushPayload()         {}
func (p *Pipeline) PostScmCheckoutPushPayload()        {}
func (p *Pipeline) PreScmPublish()                     {}
func (p *Pipeline) PostScmPublish()                    {}
func (p *Pipeline) PreScmRetrievePayload()             {}
func (p *Pipeline) PostScmRetrievePayload()            {}
func (p *Pipeline) PreScmCleanup()	               {}
func (p *Pipeline) PostScmCleanup()	               {}
func (p *Pipeline) PreAssembleStep()                   {}
func (p *Pipeline) PostAssembleStep()                  {}
func (p *Pipeline) PreDependenciesStep()               {}
func (p *Pipeline) PostDependenciesStep()              {}
func (p *Pipeline) PreCompileStep()                    {}
func (p *Pipeline) PostCompileStep()                   {}
func (p *Pipeline) PreTestStep()                       {}
func (p *Pipeline) PostTestStep()                      {}
func (p *Pipeline) PrePackageStep()                    {}
func (p *Pipeline) PostPackageStep()                   {}
func (p *Pipeline) PreDistStep()                       {}
func (p *Pipeline) PostDistStep()                      {}

func (p *Pipeline) NotifyStep(step string, callback func() error) {
	p.Scm.Notify(p.Data.GitHeadInfo.Sha, "pending", fmt.Sprintf("Started '%s' step. Pull request will be merged automatically when complete.", step))
	cerr := callback()
	if cerr != nil {
		//TODO: remove the temp folder path.
		p.Scm.Notify(p.Data.GitHeadInfo.Sha, "failure", fmt.Sprintf("Error: '%s'", cerr))
		errors.CheckErr(cerr)
	}
}

func (p *Pipeline) Cleanup() {
	if p.Config.GetBool("engine_disable_cleanup"){
		log.Println("Skipping Cleanup...")
		log.Printf("Temporary files at the following locaton should be cleaned manually: '%s'", p.Data.GitParentPath)
	} else if p.Data != nil && p.Data.GitParentPath != "" {
		log.Println("Running Cleanup...")
		os.RemoveAll(p.Data.GitParentPath)
		p.Data.GitParentPath = ""
	}

}

//type NotifyStepCallback func() error
