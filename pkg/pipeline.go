package pkg

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/engine"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/scm"
	"capsulecd/pkg/utils"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"capsulecd/pkg/mgr"
)

type Pipeline struct {
	Data   *pipeline.Data
	Config config.Interface
	Scm    scm.Interface
	Engine engine.Interface
	PackageManager mgr.Interface
}

func (p *Pipeline) Start(config config.Interface) error {
	// Initialize Pipeline.
	p.Config = config
	p.Data = new(pipeline.Data)

	defer p.Cleanup()
	if err := p.PipelineInitStep(); err != nil {
		return err
	}

	payload, err := p.ScmRetrievePayloadStep()
	if err != nil {
		return err
	}

	if p.Data.IsPullRequest {
		if perr := p.ScmCheckoutPullRequestStep(payload); perr != nil {
			return perr
		}
	} else {
		if perr := p.ScmCheckoutPushPayloadStep(payload); perr != nil {
			return perr
		}
	}

	if err := p.StepExecNotify("parse_repo_config", p.ParseRepoConfig); err != nil {
		return err
	}

	if err := p.StepExecNotify("mgr_init_step", p.MgrInitStep); err != nil {
		return err
	}


	if err := p.StepExecNotify("validate_tools", p.ValidateTools); err != nil {
		return err
	}

	if err := p.StepExecNotify("mgr_validate_tools", p.MgrValidateTools); err != nil {
		return err
	}

	if err := p.StepExecNotify("assemble_step", p.AssembleStep); err != nil { //this step includes Mgr work.
		return err
	}

	if err := p.StepExecNotify("mgr_dependencies_step", p.MgrDependenciesStep); err != nil {
		return err
	}

	if err := p.StepExecNotify("compile_step", p.CompileStep); err != nil {
		return err
	}

	if err := p.StepExecNotify("test_step", p.TestStep); err != nil {
		return err
	}

	if err := p.StepExecNotify("package_step", p.PackageStep); err != nil { //this step includes Mgr work
		return err
	}

	if err := p.StepExecNotify("mgr_dist_step", p.MgrDistStep); err != nil {
		return err
	}

	if err := p.StepExecNotify("scm_publish_step", p.ScmPublishStep); err != nil {
		return err
	}

	if err := p.StepExecNotify("scm_cleanup_step", p.ScmCleanupStep); err != nil {
		return err
	}

	//if there was an error, it should not have gotten to this point. CheckErr panic's
	p.Scm.Notify(
		p.Data.GitHeadInfo.Sha,
		"success",
		"Pull-request was successfully merged, new release created.",
	)

	return nil
}

func (p *Pipeline) PipelineInitStep() error {

	// PRE HOOK
	if err := p.RunHook("pipeline_init_step.pre"); err != nil {
		return err
	}

	if p.Config.IsSet("pipeline_init_step.override") {
		log.Println("Cannot override the pipeline_init_step, ignoring.")
	}

	// start the source, and whatever work needs to be done there.
	// MUST set options.GitParentPath
	log.Println("pipeline_init_step")
	scmImpl, serr := scm.Create(p.Config.GetString("scm"), p.Data, p.Config, nil)
	if serr != nil {
		return serr
	}
	p.Scm = scmImpl

	//Generate a new instance of the engine
	engineImpl, eerr := engine.Create(p.Config.GetString("package_type"), p.Data, p.Config, p.Scm)
	if eerr != nil {
		return eerr
	}
	p.Engine = engineImpl

	// POST HOOK
	if err := p.RunHook("pipeline_init_step.post"); err != nil {
		return err
	}
	return nil
}

func (p *Pipeline) ScmRetrievePayloadStep() (*scm.Payload, error) {

	// PRE HOOK
	if err := p.RunHook("scm_retrieve_payload_step.pre"); err != nil {
		return nil, err
	}

	if p.Config.IsSet("scm_retrieve_payload_step.override") {
		log.Println("Cannot override the scm_retrieve_payload_step, ignoring.")
	}

	log.Println("scm_retrieve_payload_step")
	payload, perr := p.Scm.RetrievePayload()
	if perr != nil {
		return nil, perr
	}

	// POST HOOK
	if err := p.RunHook("scm_retrieve_payload_step.post"); err != nil {
		return nil, err
	}
	return payload, nil
}

func (p *Pipeline) ScmCheckoutPullRequestStep(payload *scm.Payload) error {

	// PRE HOOK
	if err := p.RunHook("scm_checkout_pull_request_step.pre"); err != nil {
		return err
	}

	if p.Config.IsSet("scm_checkout_pull_request_step.override") {
		if err := p.RunHook("scm_checkout_pull_request_step.override"); err != nil {
			return err
		}
	} else {
		log.Println("scm_checkout_pull_request_step")
		if err := p.Scm.CheckoutPullRequestPayload(payload); err != nil {
			return err
		}
	}

	// POST HOOK
	if err := p.RunHook("scm_checkout_pull_request_step.post"); err != nil {
		return err
	}
	return nil
}

func (p *Pipeline) ScmCheckoutPushPayloadStep(payload *scm.Payload) error {

	// PRE HOOK
	if err := p.RunHook("scm_checkout_push_payload_step.pre"); err != nil {
		return err
	}

	if p.Config.IsSet("scm_checkout_push_payload_step.override") {
		if err := p.RunHook("scm_checkout_push_payload_step.override"); err != nil {
			return err
		}
	} else {
		log.Println("scm_checkout_push_payload_step")
		if err := p.Scm.CheckoutPushPayload(payload); err != nil {
			return err
		}
	}

	// POST HOOK
	if err := p.RunHook("scm_checkout_push_payload_step.post"); err != nil {
		return err
	}
	return nil
}

func (p *Pipeline) ParseRepoConfig() error {
	log.Println("parse_repo_config")
	// update the config with repo config file options
	repoConfig := path.Join(p.Data.GitLocalPath, "capsule.yml")
	if utils.FileExists(repoConfig) {
		if err := p.Config.ReadConfig(repoConfig); err != nil {
			return errors.New("An error occured while parsing repository capsule.yml file")
		}
	} else {
		log.Println("No repo capsule.yml file found, using existing config.")
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
	return nil
}

func (p *Pipeline) MgrInitStep() error {
	log.Println("mgr_init_step")
	if p.Config.IsSet("mgr_type") {
		mgr, merr := mgr.Create(p.Config.GetString("mgr_type"), p.Data, p.Config, nil)
		if merr != nil {
			return merr
		}
		p.PackageManager = mgr
	} else {
		mgr, merr := mgr.Detect(p.Config.GetString("package_type"), p.Data, p.Config, nil)
		if merr != nil {
			return merr
		}
		p.PackageManager = mgr
	}
	return nil
}

// validate that required executables are available for the following build/test/package/etc steps
func (p *Pipeline) ValidateTools() error {
	log.Println("validate_tools")
	return p.Engine.ValidateTools()
}

func (p *Pipeline) MgrValidateTools() error {
	log.Println("mgr_validate_tools")
	return p.PackageManager.MgrValidateTools()
}


// now that the payload has been processed we can begin by building the code.
// this may be creating missing files/default structure, compilation, version bumping, etc.
func (p *Pipeline) AssembleStep() error {
	// PRE HOOK
	if err := p.RunHook("assemble_step.pre"); err != nil {
		return err
	}

	if p.Config.IsSet("assemble_step.override") {
		log.Println("Cannot override the assemble_step, ignoring.")

	}
	log.Println("assemble_step")
	if err := p.Engine.AssembleStep(); err != nil {
		return err
	}
	log.Println("mgr_assemble_step")
	if err := p.PackageManager.MgrAssembleStep(); err != nil {
		return err
	}
	// POST HOOK
	if err := p.RunHook("assemble_step.post"); err != nil {
		return err
	}
	return nil
}

// this step should download dependencies
func (p *Pipeline) MgrDependenciesStep() error {
	// PRE HOOK
	if err := p.RunHook("mgr_dependencies_step.pre"); err != nil {
		return err
	}

	if p.Config.IsSet("mgr_dependencies_step.override") {
		if err := p.RunHook("mgr_dependencies_step.override"); err != nil {
			return err
		}
	} else {
		log.Println("mgr_dependencies_step")
		if err := p.PackageManager.MgrDependenciesStep(p.Engine.GetCurrentMetadata(), p.Engine.GetNextMetadata()); err != nil {
			return err
		}
	}

	// POST HOOK
	if err := p.RunHook("mgr_dependencies_step.post"); err != nil {
		return err
	}
	return nil
}

// this step should compile source
func (p *Pipeline) CompileStep() error {
	if p.Config.GetBool("engine_disable_compile") {
		log.Println("skipping compile_step.pre, compile_step, compile_step.post")
		return nil
	}

	// PRE HOOK
	if err := p.RunHook("compile_step.pre"); err != nil {
		return err
	}

	if p.Config.IsSet("compile_step.override") {
		if err := p.RunHook("compile_step.override"); err != nil {
			return err
		}
	} else {
		log.Println("compile_step")
		if err := p.Engine.CompileStep(); err != nil {
			return err
		}
	}

	// POST HOOK
	if err := p.RunHook("compile_step.post"); err != nil {
		return err
	}
	return nil
}

// run the package test runner(s) (eg. npm test, rake test, kitchen test) and linters/formatters
func (p *Pipeline) TestStep() error {
	if p.Config.GetBool("engine_disable_test") {
		log.Println("skipping test_step.pre, test_step, test_step.post")
		return nil
	}

	// PRE HOOK
	if err := p.RunHook("test_step.pre"); err != nil {
		return err
	}

	if p.Config.IsSet("test_step.override") {
		if err := p.RunHook("test_step.override"); err != nil {
			return err
		}
	} else {
		log.Println("test_step")
		if err := p.Engine.TestStep(); err != nil {
			return err
		}
	}

	// POST HOOK
	if err := p.RunHook("test_step.post"); err != nil {
		return err
	}
	return nil
}

// this step should commit any local changes and create a git tag. It should also generate the releaser artifacts. Nothing should be pushed to remote repository
func (p *Pipeline) PackageStep() error {
	// PRE HOOK
	if err := p.RunHook("package_step.pre"); err != nil {
		return err
	}

	if p.Config.IsSet("package_step.override") {
		log.Println("Cannot override the package_step, ignoring.")
	}
	log.Println("mgr_package_step")
	if err := p.PackageManager.MgrPackageStep(p.Engine.GetCurrentMetadata(), p.Engine.GetNextMetadata()); err != nil {
		return err
	}
	log.Println("package_step")
	if err := p.Engine.PackageStep(); err != nil {
		return err
	}

	// POST HOOK
	if err := p.RunHook("package_step.post"); err != nil {
		return err
	}
	return nil
}

// this step should push the release to the package repository (ie. npm, chef supermarket, rubygems)
func (p *Pipeline) MgrDistStep() error {
	if p.Config.GetBool("mgr_disable_dist") {
		log.Println("skipping mgr_dist_step.pre, mgr_dist_step, mgr_dist_step.post")
		return nil
	}

	// PRE HOOK
	if err := p.RunHook("mgr_dist_step.pre"); err != nil {
		return err
	}

	if p.Config.IsSet("mgr_dist_step.override") {
		if err := p.RunHook("mgr_dist_step.override"); err != nil {
			return err
		}
	} else {
		log.Println("mgr_dist_step")
		if err := p.PackageManager.MgrDistStep(p.Engine.GetCurrentMetadata(), p.Engine.GetNextMetadata()); err != nil {
			return err
		}
	}

	// POST HOOK
	if err := p.RunHook("mgr_dist_step.post"); err != nil {
		return err
	}
	return nil
}

func (p *Pipeline) ScmPublishStep() error {
	if p.Config.GetBool("scm_disable_publish") {
		log.Println("skipping scm_publish_step.pre, scm_publish_step, scm_publish_step.post")
		return nil
	}

	// PRE HOOK
	if err := p.RunHook("scm_publish_step.pre"); err != nil {
		return err
	}

	if p.Config.IsSet("scm_publish_step.override") {
		if err := p.RunHook("scm_publish_step.override"); err != nil {
			return err
		}
	} else {
		log.Println("scm_publish_step")
		if err := p.Scm.Publish(); err != nil {
			return err
		}
	}

	// POST HOOK
	if err := p.RunHook("scm_publish_step.post"); err != nil {
		return err
	}
	return nil
}

func (p *Pipeline) ScmCleanupStep() error {
	if p.Config.GetBool("scm_disable_cleanup") {
		log.Println("skipping scm_cleanup_step.pre, scm_cleanup_step, scm_cleanup_step.post")
		return nil
	}

	// PRE HOOK
	if err := p.RunHook("scm_cleanup_step.pre"); err != nil {
		return err
	}

	if p.Config.IsSet("scm_cleanup_step.override") {
		if err := p.RunHook("scm_cleanup_step.override"); err != nil {
			return err
		}
	} else {
		log.Println("scm_cleanup_step")
		if err := p.Scm.Cleanup(); err != nil {
			return err
		}
	}

	// POST HOOK
	if err := p.RunHook("scm_cleanup_step.post"); err != nil {
		return err
	}
	return nil
}

// Helpers

func (p *Pipeline) StepExecNotify(step string, callback func() error) error {
	p.Scm.Notify(p.Data.GitHeadInfo.Sha, "pending", fmt.Sprintf("Started '%s' step. Pull request will be merged automatically when complete.", step))
	if cerr := callback(); cerr != nil {
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

func (p *Pipeline) RunHook(hookKey string) error {
	log.Println(hookKey)

	hookSteps := p.Config.GetStringSlice(hookKey)
	if hookSteps == nil {
		return nil
	}
	for i, cmd := range hookSteps {
		cmdPopulated, aerr := utils.PopulateTemplate(cmd, p.Data)
		if aerr != nil {
			return aerr
		}

		if err := utils.BashCmdExec(cmdPopulated, p.Data.GitLocalPath, nil, fmt.Sprintf("%s.%s", hookKey, i)); err != nil {
			return err
		}
	}
	return nil
}
