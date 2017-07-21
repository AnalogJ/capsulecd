package engine

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/errors"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/scm"
	"capsulecd/pkg/utils"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

type nodeMetadata struct {
	Version string
}
type engineNode struct {
	*engineBase

	PipelineData *pipeline.Data
	Scm          scm.Interface //Interface
}

func (n *engineNode) init(pipelineData *pipeline.Data, config config.Interface, sourceScm scm.Interface) error {
	n.Scm = sourceScm
	n.Config = config
	n.PipelineData = pipelineData

	//set command defaults (can be overridden by repo/system configuration)
	n.Config.SetDefault("engine_cmd_lint", "eslint --fix .")
	n.Config.SetDefault("engine_cmd_fmt", "eslint --fix .")
	n.Config.SetDefault("engine_cmd_test", "npm test")
	n.Config.SetDefault("engine_cmd_security_check", "nsp check")
	return nil
}

func (n *engineNode) ValidateTools() error {
	if _, kerr := exec.LookPath("npm"); kerr != nil {
		return errors.EngineValidateToolError("npm binary is missing")
	}

	if _, kerr := exec.LookPath("node"); kerr != nil {
		return errors.EngineValidateToolError("node binary is missing")
	}

	if _, kerr := exec.LookPath("eslint"); kerr != nil && !n.Config.GetBool("engine_disable_lint") {
		return errors.EngineValidateToolError("eslint binary is missing")
	}

	if _, kerr := exec.LookPath("nsp"); kerr != nil && !n.Config.GetBool("engine_disable_security_check") {
		return errors.EngineValidateToolError("nsp binary is missing")
	}

	return nil
}

func (n *engineNode) AssembleStep() error {
	//validate that the npm package.json file exists
	if !utils.FileExists(path.Join(n.PipelineData.GitLocalPath, "package.json")) {
		return errors.EngineBuildPackageInvalid("package.json file is required to process Node package")
	}

	// no need to bump up the version here. It will automatically be bumped up via the npm version patch command.
	// however we need to read the version from the package.json file and check if a npm module already exists.

	//TODO: check if this module name and version already exist.

	// check for/create any required missing folders/files
	if derr := os.MkdirAll(path.Join(n.PipelineData.GitLocalPath, "test"), 0644); derr != nil {
		return derr
	}

	gitignorePath := path.Join(n.PipelineData.GitLocalPath, ".gitignore")
	if !utils.FileExists(gitignorePath) {
		if err := utils.GitGenerateGitIgnore(n.PipelineData.GitLocalPath, "Node"); err != nil {
			return err
		}
	}
	return nil
}

func (n *engineNode) DependenciesStep() error {
	// the module has already been downloaded. lets make sure all its dependencies are available.
	if derr := utils.BashCmdExec("npm install", n.PipelineData.GitLocalPath, ""); derr != nil {
		return errors.EngineTestRunnerError("npm install failed. Check module dependencies")
	}

	// create a shrinkwrap file.
	if derr := utils.BashCmdExec("npm shrinkwrap", n.PipelineData.GitLocalPath, ""); derr != nil {
		return errors.EngineTestRunnerError("npm shrinkwrap failed. Check log for exact error")
	}
	return nil
}

func (n *engineNode) TestStep() error {

	//skip the lint commands if disabled
	if !n.Config.GetBool("engine_disable_lint") {
		//run test command
		lintCmd := n.Config.GetString("engine_cmd_lint")
		if n.Config.GetBool("engine_enable_code_mutation") {
			lintCmd = n.Config.GetString("engine_cmd_fmt")
		}
		if terr := utils.BashCmdExec(lintCmd, n.PipelineData.GitLocalPath, ""); terr != nil {
			return errors.EngineTestRunnerError(fmt.Sprintf("Lint command (%s) failed. Check log for more details.", lintCmd))
		}
	}

	//skip the test commands if disabled
	if !n.Config.GetBool("engine_disable_test") {
		//run test command
		testCmd := n.Config.GetString("engine_cmd_test")
		if derr := utils.BashCmdExec(testCmd, n.PipelineData.GitLocalPath, ""); derr != nil {
			return errors.EngineTestRunnerError(fmt.Sprintf("Test command (%s) failed. Check log for more details.", testCmd))
		}
	}

	//skip the security test commands if disabled
	if !n.Config.GetBool("engine_disable_security_check") {
		//run security check command
		vulCmd := n.Config.GetString("engine_cmd_security_check")
		if terr := utils.BashCmdExec(vulCmd, n.PipelineData.GitLocalPath, ""); terr != nil {
			return errors.EngineTestRunnerError(fmt.Sprintf("Dependency vulnerability check command (%s) failed. Check log for more details.", vulCmd))
		}
	}

	return nil
}

func (n *engineNode) PackageStep() error {
	if !n.Config.GetBool("engine_package_keep_lock_file") {
		os.Remove(path.Join(n.PipelineData.GitLocalPath, "npm-shrinkwrap.json"))
	}

	if cerr := utils.GitCommit(n.PipelineData.GitLocalPath, "Committing automated changes before packaging."); cerr != nil {
		return cerr
	}

	// run npm publish
	versionCmd := fmt.Sprintf("npm version %s -m '(v%%s) Automated packaging of release by CapsuleCD'",
		n.Config.GetString("engine_version_bump_type"),
	)
	if verr := utils.BashCmdExec(versionCmd, n.PipelineData.GitLocalPath, ""); verr != nil {
		return errors.EngineTestRunnerError("npm version bump failed")
	}

	tagCommit, terr := utils.GitLatestTaggedCommit(n.PipelineData.GitLocalPath)
	if terr != nil {
		return terr
	}

	n.PipelineData.ReleaseCommit = tagCommit.CommitSha
	n.PipelineData.ReleaseVersion = tagCommit.TagShortName
	return nil
}

func (n *engineNode) DistStep() error {
	if !n.Config.IsSet("npm_auth_token") {
		return errors.EngineDistCredentialsMissing("cannot deploy page to npm, credentials missing")
	}

	npmrcFile, _ := ioutil.TempFile("", ".npmrc")
	defer os.Remove(npmrcFile.Name())

	// write the .npmrc config jfile.
	npmrcContent := fmt.Sprintf(
		"//registry.npmjs.org/:_authToken=%s",
		n.Config.GetString("npm_auth_token"),
	)

	if _, werr := npmrcFile.Write([]byte(npmrcContent)); werr != nil {
		return werr
	}

	npmPublishCmd := "npm publish ."
	derr := utils.BashCmdExec(npmPublishCmd, n.PipelineData.GitLocalPath, "")
	if derr != nil {
		return errors.EngineDistPackageError("npm publish failed. Check log for exact error")
	}
	return nil
}
