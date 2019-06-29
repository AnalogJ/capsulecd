package engine

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/errors"
	"capsulecd/pkg/metadata"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/scm"
	"capsulecd/pkg/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

type engineNode struct {
	engineBase

	Scm             scm.Interface //Interface
	CurrentMetadata *metadata.NodeMetadata
	NextMetadata    *metadata.NodeMetadata
}

func (g *engineNode) Init(pipelineData *pipeline.Data, config config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = config
	g.PipelineData = pipelineData
	g.CurrentMetadata = new(metadata.NodeMetadata)
	g.NextMetadata = new(metadata.NodeMetadata)

	//set command defaults (can be overridden by repo/system configuration)
	g.Config.SetDefault("engine_cmd_compile", "echo 'skipping compile'")
	g.Config.SetDefault("engine_cmd_lint", "eslint --fix .")
	g.Config.SetDefault("engine_cmd_fmt", "eslint --fix .")
	g.Config.SetDefault("engine_cmd_test", "npm test")
	g.Config.SetDefault("engine_cmd_security_check", "nsp check")
	return nil
}

func (g *engineNode) GetCurrentMetadata() interface{} {
	return g.CurrentMetadata
}
func (g *engineNode) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *engineNode) ValidateTools() error {

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

// use default Compile step
//func (g *engineNode) CompileStep() error { }

// use default Test step
//func (g *engineNode) TestStep() error { }

func (g *engineNode) PackageStep() error {

	if cerr := utils.GitCommit(g.PipelineData.GitLocalPath, fmt.Sprintf("(v%s) %s", g.NextMetadata.Version, g.Config.GetString("engine_version_bump_msg"))); cerr != nil {
		return cerr
	}

	tagCommit, terr := utils.GitTag(g.PipelineData.GitLocalPath, fmt.Sprintf("v%s", g.NextMetadata.Version), g.Config.GetString("engine_version_bump_msg"))
	if terr != nil {
		return terr
	}

	g.PipelineData.ReleaseCommit = tagCommit
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
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
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
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
