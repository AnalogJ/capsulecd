package pkg

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/engine"
	"capsulecd/pkg/errors"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/scm"
	"capsulecd/pkg/utils"
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

func (p *Pipeline) Start(config config.Interface) error {
	defer p.Cleanup()

	p.Config = config
	p.Data = new(pipeline.Data)

	// start the source, and whatever work needs to be done there.
	// MUST set options.GitParentPath
	p.PreScmInit()
	log.Println("scm_init_step")
	scmImpl, serr := scm.Create(p.Config.GetString("scm"), p.Data, config, nil)
	if serr != nil {
		return serr;
	}
	p.Scm = scmImpl
	p.PostScmInit()

	//Generate a new instance of the engine
	engineImpl, eerr := engine.Create(p.Config.GetString("package_type"), p.Data, config, scmImpl)
	if eerr != nil {
		return eerr;
	}
	p.Engine = engineImpl

	//retreive payload
	p.PreScmRetrievePayload()
	log.Println("scm_retrieve_payload_step")
	payload, perr := scmImpl.RetrievePayload()
	if perr != nil {
		return perr;
	}
	p.PostScmRetrievePayload()

	if p.Data.IsPullRequest {
		p.PreScmCheckoutPullRequestPayload()
		log.Println("scm_checkout_pull_request_step")
		if perr := scmImpl.CheckoutPullRequestPayload(payload); perr != nil {
			return perr
		}
		p.PostScmCheckoutPullRequestPayload()
	} else {
		p.PreScmCheckoutPushPayload()
		log.Println("scm_checkout_push_payload_step")
		if perr := scmImpl.CheckoutPushPayload(payload); perr != nil {
			return perr
		}
		p.PostScmCheckoutPushPayload()
	}

	// update the config with repo config file options
	repoConfig := path.Join(p.Data.GitLocalPath, "capsule.yml")
	if utils.FileExists(repoConfig) {
		if err := p.Config.ReadConfig(repoConfig); err != nil {

		}
	}
	if p.Config.IsSet("scm_release_assets") {
		//unmarshall config data.
		parsedAssets := new([]pipeline.ScmReleaseAsset)
		if err := p.Config.UnmarshalKey("scm_release_assets", parsedAssets); err != nil {
			return err
		}

		//append the parsed Assets to the current ReleaseAssets storage (incase assets were defined in system yml)
		p.Data.ReleaseAssets = append(p.Data.ReleaseAssets, (*parsedAssets)...)
	}

	//validate that required executables are available for the following build/test/package/etc steps
	if err := p.NotifyStep("validate tools", func() error {
		log.Println("validate_tools_step")
		return engineImpl.ValidateTools()
	}); err != nil {
		return err;
	}

	// now that the payload has been processed we can begin by building the code.
	// this may be creating missing files/default structure, compilation, version bumping, etc.
	if err := p.NotifyStep("assemble", func() error {
		p.PreAssembleStep()
		log.Println("assemble_step")
		if berr := engineImpl.AssembleStep(); berr != nil {
			return berr
		}
		p.PostAssembleStep()
		return nil
	}); err != nil {
		return err;
	}

	// this step should download dependencies
	if err := p.NotifyStep("dependencies", func() error {
		p.PreDependenciesStep()
		log.Println("dependencies_step")
		if berr := engineImpl.DependenciesStep(); berr != nil {
			return berr
		}
		p.PostDependenciesStep()
		return nil
	}); err != nil {
		return err;
	}

	// this step should compile source
	if err := p.NotifyStep("compile", func() error {
		p.PreCompileStep()
		log.Println("compile_step")
		if berr := engineImpl.CompileStep(); berr != nil {
			return berr
		}
		p.PostCompileStep()
		return nil
	}); err != nil {
		return err;
	}

	// run the package test runner(s) (eg. npm test, rake test, kitchen test) and linters/formatters
	if err := p.NotifyStep("test", func() error {
		p.PreTestStep()
		log.Println("test_step")
		if terr := engineImpl.TestStep(); terr != nil {
			return terr
		}
		p.PostTestStep()
		return nil
	}); err != nil {
		return err;
	}

	// this step should commit any local changes and create a git tag. Nothing should be pushed to remote repository
	if err := p.NotifyStep("package", func() error {
		p.PrePackageStep()
		log.Println("package_step")
		if perr := engineImpl.PackageStep(); perr != nil {
			return perr
		}
		p.PostPackageStep()
		return nil
	}); err != nil {
		return err;
	}

	if p.Data.IsPullRequest {
		// this step should push the release to the package repository (ie. npm, chef supermarket, rubygems)
		if err := p.NotifyStep("dist", func() error {
			if p.Config.GetBool("engine_disable_dist") {
				log.Println("skipping pre_dist_step, dist_step, post_dist_step")
				return nil
			}

			p.PreDistStep()
			log.Println("dist_step")
			if derr := engineImpl.DistStep(); derr != nil {
				return derr
			}
			p.PostDistStep()
			return nil
		}); err != nil {
			return err;
		}


		if err := p.NotifyStep("scm publish", func() error {
			if p.Config.GetBool("scm_disable_publish") {
				log.Println("skipping pre_scm_publish_step, scm_publish_step, post_scm_publish_step")
				return nil
			}

			p.PreScmPublish()
			log.Println("scm_publish_step")
			if serr := scmImpl.Publish(); serr != nil {
				return serr
			}
			p.PostScmPublish()
			return nil
		}); err != nil {
			return err;
		}

		if err := p.NotifyStep("scm cleanup", func() error {
			p.PreScmCleanup()
			log.Println("scm_cleanup_step")
			if serr := scmImpl.Cleanup(); serr != nil {
				//if theres an error, just print it (cleanup will not fail the deployment)
				// it will stop PostSCM cleanup from running.
				log.Print(serr)
				return nil
			}
			p.PostScmCleanup()
			return nil
		}); err != nil {
			return err;
		}

		//if there was an error, it should not have gotten to this point. CheckErr panic's
		scmImpl.Notify(
			p.Data.GitHeadInfo.Sha,
			"success",
			"Pull-request was successfully merged, new release created.",
		)
	}

	return nil
}

// Hook methods
func (p *Pipeline) PreScmInit()                       { p.RunHook("scm_init_step.pre") }
func (p *Pipeline) PostScmInit()                      { p.RunHook("scm_init_step.post") }
func (p *Pipeline) PreScmRetrievePayload()            { p.RunHook("scm_retrieve_payload_step.pre") }
func (p *Pipeline) PostScmRetrievePayload()           { p.RunHook("scm_retrieve_payload_step.post") }
func (p *Pipeline) PreScmCheckoutPullRequestPayload() { p.RunHook("scm_checkout_pull_request_step.pre") }
func (p *Pipeline) PostScmCheckoutPullRequestPayload() {
	p.RunHook("scm_checkout_pull_request_step.post")
}
func (p *Pipeline) PreScmCheckoutPushPayload()  { p.RunHook("scm_checkout_push_payload_step.pre") }
func (p *Pipeline) PostScmCheckoutPushPayload() { p.RunHook("scm_checkout_push_payload_step.post") }
func (p *Pipeline) PreScmPublish()              { p.RunHook("scm_publish_step.pre") }
func (p *Pipeline) PostScmPublish()             { p.RunHook("scm_publish_step.post") }
func (p *Pipeline) PreScmCleanup()              { p.RunHook("scm_cleanup_step.pre") }
func (p *Pipeline) PostScmCleanup()             { p.RunHook("scm_cleanup_step.post") }
func (p *Pipeline) PreAssembleStep()            { p.RunHook("assemble_step.pre") }
func (p *Pipeline) PostAssembleStep()           { p.RunHook("assemble_step.post") }
func (p *Pipeline) PreDependenciesStep()        { p.RunHook("dependencies_step.pre") }
func (p *Pipeline) PostDependenciesStep()       { p.RunHook("dependencies_step.post") }
func (p *Pipeline) PreCompileStep()             { p.RunHook("compile_step.pre") }
func (p *Pipeline) PostCompileStep()            { p.RunHook("compile_step.post") }
func (p *Pipeline) PreTestStep()                { p.RunHook("test_step.pre") }
func (p *Pipeline) PostTestStep()               { p.RunHook("test_step.post") }
func (p *Pipeline) PrePackageStep()             { p.RunHook("package_step.pre") }
func (p *Pipeline) PostPackageStep()            { p.RunHook("package_step.post") }
func (p *Pipeline) PreDistStep()                { p.RunHook("dist_step.pre") }
func (p *Pipeline) PostDistStep()               { p.RunHook("dist_step.post") }

func (p *Pipeline) NotifyStep(step string, callback func() error) error {
	p.Scm.Notify(p.Data.GitHeadInfo.Sha, "pending", fmt.Sprintf("Started '%s' step. Pull request will be merged automatically when complete.", step))
	if cerr := callback(); cerr != nil {
		//TODO: remove the temp folder path.
		p.Scm.Notify(p.Data.GitHeadInfo.Sha, "failure", fmt.Sprintf("Error: '%s'", cerr))
		return cerr
	}
	return nil
}

func (p *Pipeline) Cleanup() {
	if p.Config.GetBool("engine_disable_cleanup") {
		log.Println("Skipping Cleanup...")
		log.Printf("Temporary files at the following locaton should be cleaned manually: '%s'", p.Data.GitParentPath)
	} else if p.Data != nil && p.Data.GitParentPath != "" {
		log.Println("Running Cleanup...")
		os.RemoveAll(p.Data.GitParentPath)
		p.Data.GitParentPath = ""
	}
}

func (p *Pipeline) RunHook(hookKey string) {
	log.Println(hookKey)

	hookSteps := p.Config.GetStringSlice(hookKey)
	if hookSteps == nil {
		return
	}
	for i := range hookSteps {
		utils.BashCmdExec(hookSteps[i], p.Data.GitLocalPath, nil, fmt.Sprintf("%s.%s", hookKey, i))
	}
}

//type NotifyStepCallback func() error
