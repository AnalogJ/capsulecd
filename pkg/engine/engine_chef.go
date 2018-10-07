package engine

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/errors"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/scm"
	"capsulecd/pkg/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"capsulecd/pkg/metadata"
)


type engineChef struct {
	engineBase
	Scm             scm.Interface //Interface
	CurrentMetadata *metadata.ChefMetadata
	NextMetadata    *metadata.ChefMetadata
}

func (g *engineChef) Init(pipelineData *pipeline.Data, configData config.Interface, sourceScm scm.Interface) error {
	g.Config = configData
	g.Scm = sourceScm
	g.PipelineData = pipelineData
	g.CurrentMetadata = new(metadata.ChefMetadata)
	g.NextMetadata = new(metadata.ChefMetadata)

	//set command defaults (can be overridden by repo/system configuration)
	g.Config.SetDefault("chef_supermarket_type", "Other")
	g.Config.SetDefault("engine_cmd_compile", "echo 'skipping compile'")
	g.Config.SetDefault("engine_cmd_lint", "foodcritic .")
	g.Config.SetDefault("engine_cmd_fmt", "foodcritic .")
	g.Config.SetDefault("engine_cmd_test", "rake test")
	g.Config.SetDefault("engine_cmd_security_check", "bundle audit check --update")

	return nil
}

func (g *engineChef) GetCurrentMetadata() interface{} {
	return g.CurrentMetadata
}
func (g *engineChef) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *engineChef) ValidateTools() error {
	if _, berr := exec.LookPath("foodcritic"); berr != nil && !g.Config.GetBool("engine_disable_lint") {
		return errors.EngineValidateToolError("foodcritic binary is missing")
	}

	return nil
}

func (g *engineChef) AssembleStep() error {
	//validate that the chef metadata.rb file exists

	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "metadata.rb")) {
		return errors.EngineBuildPackageInvalid("metadata.rb file is required to process Chef cookbook")
	}

	// bump up the chef cookbook version
	if merr := g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath); merr != nil {
		return merr
	}

	if perr := g.populateNextMetadata(); perr != nil {
		return perr
	}

	if nerr := g.writeNextMetadata(g.PipelineData.GitLocalPath); nerr != nil {
		return nerr
	}

	// TODO: check if this cookbook name and version already exist.
	// check for/create any required missing folders/files
	// Berksfile.lock and Gemfile.lock are not required to be committed, but they should be.
	rakefilePath := path.Join(g.PipelineData.GitLocalPath, "Rakefile")
	if !utils.FileExists(rakefilePath) {
		ioutil.WriteFile(rakefilePath, []byte("task :test"), 0644)
	}

	specPath := path.Join(g.PipelineData.GitLocalPath, "spec")
	if !utils.FileExists(specPath) {
		os.MkdirAll(specPath, 0777)
	}

	gitignorePath := path.Join(g.PipelineData.GitLocalPath, ".gitignore")
	if !utils.FileExists(gitignorePath) {
		if err := utils.GitGenerateGitIgnore(g.PipelineData.GitLocalPath, "ChefCookbook"); err != nil {
			log.Print("Generate error")
			return err
		}
	}

	return nil
}


// Use default compile step..
// func (g *engineChef) CompileStep() error {}

// Use default test step.
// func (g *engineChef) TestStep() error {}

func (g *engineChef) PackageStep() error {

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

//private Helpers

func (g *engineChef) retrieveCurrentMetadata(gitLocalPath string) error {
	//dat, err := ioutil.ReadFile(path.Join(gitLocalPath, "metadata.rb"))
	//knife cookbook metadata -o ../ chef-mycookbook -- will generate a metadata.json file.
	if cerr := utils.BashCmdExec(fmt.Sprintf("knife cookbook metadata -o ../ %s", path.Base(gitLocalPath)), gitLocalPath, nil, ""); cerr != nil {
		return cerr
	}
	defer os.Remove(path.Join(gitLocalPath, "metadata.json"))

	//read metadata.json file.
	metadataContent, rerr := ioutil.ReadFile(path.Join(gitLocalPath, "metadata.json"))
	if rerr != nil {
		return rerr
	}

	if uerr := json.Unmarshal(metadataContent, g.CurrentMetadata); uerr != nil {
		return uerr
	}

	return nil
}

func (g *engineChef) populateNextMetadata() error {

	nextVersion, err := g.BumpVersion(g.CurrentMetadata.Version)
	if err != nil {
		return err
	}

	g.NextMetadata.Version = nextVersion
	g.NextMetadata.Name = g.CurrentMetadata.Name
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}

func (g *engineChef) writeNextMetadata(gitLocalPath string) error {
	return utils.BashCmdExec(fmt.Sprintf("knife spork bump %s manual %s -o ../", path.Base(gitLocalPath), g.NextMetadata.Version), gitLocalPath, nil, "")
}
