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
	"encoding/json"
)

type nodeMetadata struct {
	Version string `json:"version"`
	Name    string `json:"name"`
}
type engineNode struct {
	engineBase

	PipelineData *pipeline.Data
	Scm          scm.Interface //Interface
	CurrentMetadata *nodeMetadata
	NextMetadata    *nodeMetadata
}

func (g *engineNode) Init(pipelineData *pipeline.Data, config config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = config
	g.PipelineData = pipelineData
	g.CurrentMetadata = new(nodeMetadata)
	g.NextMetadata = new(nodeMetadata)

	//set command defaults (can be overridden by repo/system configuration)
	g.Config.SetDefault("engine_cmd_lint", "eslint --fix .")
	g.Config.SetDefault("engine_cmd_fmt", "eslint --fix .")
	g.Config.SetDefault("engine_cmd_test", "npm test")
	g.Config.SetDefault("engine_cmd_security_check", "nsp check")
	return nil
}

func (g *engineNode) ValidateTools() error {
	if _, kerr := exec.LookPath("npm"); kerr != nil {
		return errors.EngineValidateToolError("npm binary is missing")
	}

	if _, kerr := exec.LookPath("node"); kerr != nil {
		return errors.EngineValidateToolError("node binary is missing")
	}

	if _, kerr := exec.LookPath("eslint"); kerr != nil && !g.Config.GetBool("engine_disable_lint") {
		return errors.EngineValidateToolError("eslint binary is missing")
	}

	if _, kerr := exec.LookPath("nsp"); kerr != nil && !g.Config.GetBool("engine_disable_security_check") {
		return errors.EngineValidateToolError("nsp binary is missing")
	}

	return nil
}

func (g *engineNode) AssembleStep() error {
	//validate that the npm package.json file exists
	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "package.json")) {
		return errors.EngineBuildPackageInvalid("package.json file is required to process Node package")
	}

	// bump up the package version
	if merr := g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath); merr != nil {
		return merr
	}

	if perr := g.populateNextMetadata(); perr != nil {
		return perr
	}

	if nerr := g.writeNextMetadata(g.PipelineData.GitLocalPath); nerr != nil {
		return nerr
	}

	// check for/create any required missing folders/files
	if derr := os.MkdirAll(path.Join(g.PipelineData.GitLocalPath, "test"), 0644); derr != nil {
		return derr
	}

	gitignorePath := path.Join(g.PipelineData.GitLocalPath, ".gitignore")
	if !utils.FileExists(gitignorePath) {
		if err := utils.GitGenerateGitIgnore(g.PipelineData.GitLocalPath, "Node"); err != nil {
			return err
		}
	}
	return nil
}

func (g *engineNode) DependenciesStep() error {
	// the module has already been downloaded. lets make sure all its dependencies are available.
	if derr := utils.BashCmdExec("npm install", g.PipelineData.GitLocalPath, nil, ""); derr != nil {
		return errors.EngineTestRunnerError("npm install failed. Check module dependencies")
	}

	// create a shrinkwrap file.
	if derr := utils.BashCmdExec("npm shrinkwrap", g.PipelineData.GitLocalPath, nil, ""); derr != nil {
		return errors.EngineTestRunnerError("npm shrinkwrap failed. Check log for exact error")
	}
	return nil
}

func (g *engineNode) CompileStep() error {
	return nil
}

func (g *engineNode) TestStep() error {

	//skip the lint commands if disabled
	if !g.Config.GetBool("engine_disable_lint") {
		//run test command
		lintCmd := g.Config.GetString("engine_cmd_lint")
		if g.Config.GetBool("engine_enable_code_mutation") {
			lintCmd = g.Config.GetString("engine_cmd_fmt")
		}
		if terr := utils.BashCmdExec(lintCmd, g.PipelineData.GitLocalPath, nil, ""); terr != nil {
			return errors.EngineTestRunnerError(fmt.Sprintf("Lint command (%s) failed. Check log for more details.", lintCmd))
		}
	}

	//skip the test commands if disabled
	if !g.Config.GetBool("engine_disable_test") {
		//run test command
		testCmd := g.Config.GetString("engine_cmd_test")
		if derr := utils.BashCmdExec(testCmd, g.PipelineData.GitLocalPath, nil, ""); derr != nil {
			return errors.EngineTestRunnerError(fmt.Sprintf("Test command (%s) failed. Check log for more details.", testCmd))
		}
	}

	//skip the security test commands if disabled
	if !g.Config.GetBool("engine_disable_security_check") {
		//run security check command
		vulCmd := g.Config.GetString("engine_cmd_security_check")
		if terr := utils.BashCmdExec(vulCmd, g.PipelineData.GitLocalPath, nil, ""); terr != nil {
			return errors.EngineTestRunnerError(fmt.Sprintf("Dependency vulnerability check command (%s) failed. Check log for more details.", vulCmd))
		}
	}

	return nil
}

func (g *engineNode) PackageStep() error {
	if !g.Config.GetBool("engine_package_keep_lock_file") {
		os.Remove(path.Join(g.PipelineData.GitLocalPath, "npm-shrinkwrap.json"))
	}

	if cerr := utils.GitCommit(g.PipelineData.GitLocalPath, fmt.Sprintf("(v%s) Automated packaging of release by CapsuleCD", g.NextMetadata.Version)); cerr != nil {
		return cerr
	}

	tagCommit, terr := utils.GitTag(g.PipelineData.GitLocalPath, fmt.Sprintf("v%s", g.NextMetadata.Version))
	if terr != nil {
		return terr
	}

	g.PipelineData.ReleaseCommit = tagCommit
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}

func (g *engineNode) DistStep() error {
	if !g.Config.IsSet("npm_auth_token") {
		return errors.EngineDistCredentialsMissing("cannot deploy page to npm, credentials missing")
	}

	npmrcFile, _ := ioutil.TempFile("", ".npmrc")
	defer os.Remove(npmrcFile.Name())

	// write the .npmrc config jfile.
	npmrcContent := fmt.Sprintf(
		"//registry.npmjs.org/:_authToken=%s",
		g.Config.GetString("npm_auth_token"),
	)

	if _, werr := npmrcFile.Write([]byte(npmrcContent)); werr != nil {
		return werr
	}

	npmPublishCmd := fmt.Sprintf("npm --userconfig %s publish .", npmrcFile.Name())
	derr := utils.BashCmdExec(npmPublishCmd, g.PipelineData.GitLocalPath, nil, "")
	if derr != nil {
		return errors.EngineDistPackageError("npm publish failed. Check log for exact error")
	}
	return nil
}


//private Helpers

func (g *engineNode) retrieveCurrentMetadata(gitLocalPath string) error {
	//read package.json file.
	packageContent, rerr := ioutil.ReadFile(path.Join(gitLocalPath, "package.json"))
	if rerr != nil {
		return rerr
	}

	if uerr := json.Unmarshal(packageContent, g.CurrentMetadata); uerr != nil {
		return uerr
	}

	return nil
}

func (g *engineNode) populateNextMetadata() error {

	nextVersion, err := g.BumpVersion(g.CurrentMetadata.Version)
	if err != nil {
		return err
	}

	g.NextMetadata.Version = nextVersion
	g.NextMetadata.Name = g.CurrentMetadata.Name
	return nil
}

func (g *engineNode) writeNextMetadata(gitLocalPath string) error {
	// The version will be bumped up via the npm version command.
	// --no-git-tag-version ensures that we dont create a git commit (which npm will do by default).
	versionCmd := fmt.Sprintf("npm --no-git-tag-version version %s",
		g.NextMetadata.Version,
	)
	if verr := utils.BashCmdExec(versionCmd, g.PipelineData.GitLocalPath, nil, ""); verr != nil {
		return errors.EngineTestRunnerError("npm version bump failed")
	}
	return nil
}
