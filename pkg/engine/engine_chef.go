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
	"os"
	"os/exec"
	"path"
	"log"
)

type chefMetadata struct {
	Version string `json:"version"`
	Name    string `json:"name"`
}
type engineChef struct {
	engineBase
	PipelineData    *pipeline.Data
	Scm             scm.Interface //Interface
	CurrentMetadata *chefMetadata
	NextMetadata    *chefMetadata
}

func (g *engineChef) Init(pipelineData *pipeline.Data, configData config.Interface, sourceScm scm.Interface) error {
	g.Config = configData
	g.Scm = sourceScm
	g.PipelineData = pipelineData
	g.CurrentMetadata = new(chefMetadata)
	g.NextMetadata = new(chefMetadata)


	//set command defaults (can be overridden by repo/system configuration)
	g.Config.SetDefault("chef_supermarket_type", "Other")
	g.Config.SetDefault("engine_cmd_lint", "foodcritic .")
	g.Config.SetDefault("engine_cmd_test", "rake test")
	g.Config.SetDefault("engine_cmd_security_check", "bundle audit check --update")

	return nil
}

func (g *engineChef) ValidateTools() error {
	//a chefdk like environment needs to be available for this Engine
	if _, kerr := exec.LookPath("knife"); kerr != nil {
		return errors.EngineValidateToolError("knife binary is missing")
	}

	if _, berr := exec.LookPath("berks"); berr != nil {
		return errors.EngineValidateToolError("berkshelf binary is missing")
	}

	//TODO: figure out how to validate that "bundle audit" command exists.
	if _, berr := exec.LookPath("bundle"); berr != nil {
		return errors.EngineValidateToolError("bundler binary is missing")
	}

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
	// Berksfile.lock and Gemfile.lock are not required to be commited, but they should be.
	rakefilePath := path.Join(g.PipelineData.GitLocalPath, "Rakefile")
	if !utils.FileExists(rakefilePath) {
		ioutil.WriteFile(rakefilePath, []byte("task :test"), 0644)
	}
	berksfilePath := path.Join(g.PipelineData.GitLocalPath, "Berksfile")
	if !utils.FileExists(berksfilePath) {
		ioutil.WriteFile(berksfilePath, []byte(`source "https://supermarket.chef.io"
		metadata
		`), 0644)
	}
	gemfilePath := path.Join(g.PipelineData.GitLocalPath, "Gemfile")
	if !utils.FileExists(gemfilePath) {
		ioutil.WriteFile(gemfilePath, []byte("source \"https://rubygems.org\""), 0644)
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

func (g *engineChef) DependenciesStep() error {
	// the cookbook has already been downloaded. lets make sure all its dependencies are available.
	if cerr := utils.CmdExec("berks", []string{"install"}, g.PipelineData.GitLocalPath, ""); cerr != nil {
		return errors.EngineTestDependenciesError("berks install failed. Check cookbook dependencies")
	}

	//download all its gem dependencies
	if berr := utils.CmdExec("bundle", []string{"install"}, g.PipelineData.GitLocalPath, ""); berr != nil {
		return errors.EngineTestDependenciesError("bundle install failed. Check Gem dependencies")
	}
	return nil
}

func (g *engineChef) TestStep() error {

	//skip the lint commands if disabled
	if !g.Config.GetBool("engine_disable_lint") {
		//run test command
		lintCmd := g.Config.GetString("engine_cmd_lint")
		if terr := utils.BashCmdExec(lintCmd, g.PipelineData.GitLocalPath, ""); terr != nil {
			return errors.EngineTestRunnerError(fmt.Sprintf("Lint command (%s) failed. Check log for more details.", lintCmd))
		}
	}

	//skip the test commands if disabled
	if !g.Config.GetBool("engine_disable_test") {
		//run test command
		testCmd := g.Config.GetString("engine_cmd_test")
		if terr := utils.BashCmdExec(testCmd, g.PipelineData.GitLocalPath, ""); terr != nil {
			return errors.EngineTestRunnerError(fmt.Sprintf("Test command (%s) failed. Check log for more details.", testCmd))
		}
	}

	//skip the security test commands if disabled
	if !g.Config.GetBool("engine_disable_security_check") {
		//run security check command
		vulCmd := g.Config.GetString("engine_cmd_security_check")
		if terr := utils.BashCmdExec(vulCmd, g.PipelineData.GitLocalPath, ""); terr != nil {
			return errors.EngineTestRunnerError(fmt.Sprintf("Dependency vulnerability check command (%s) failed. Check log for more details.", vulCmd))
		}
	}
	return nil
}

func (g *engineChef) PackageStep() error {
	if !g.Config.GetBool("engine_package_keep_lock_file") {
		os.Remove(path.Join(g.PipelineData.GitLocalPath, "Berksfile.lock"))
		os.Remove(path.Join(g.PipelineData.GitLocalPath, "Gemfile.lock"))
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

func (g *engineChef) DistStep() error {
	if !g.Config.IsSet("chef_supermarket_username") || !g.Config.IsSet("chef_supermarket_key") {
		return errors.EngineDistCredentialsMissing("Cannot deploy cookbook to supermarket, credentials missing")
	}

	// knife is really sensitive to folder names. The cookbook name MUST match the folder name otherwise knife throws up
	// when doing a knife cookbook share. So we're going to make a new tmp directory, create a subdirectory with the EXACT
	// cookbook name, and then copy the cookbook contents into it. Yeah yeah, its pretty nasty, but blame Chef.
	tmpParentPath, terr := ioutil.TempDir("", "")
	if terr != nil {
		return terr
	}
	defer os.RemoveAll(tmpParentPath)

	tmpLocalPath := path.Join(tmpParentPath, g.NextMetadata.Name)
	if cerr := utils.CopyDir(g.PipelineData.GitLocalPath, tmpLocalPath); cerr != nil {
		return cerr
	}

	pemFile, _ := ioutil.TempFile("", "client.pem")
	defer os.Remove(pemFile.Name())
	knifeFile, _ := ioutil.TempFile("", "knife.rb")
	defer os.Remove(knifeFile.Name())

	// write the knife.rb config jfile.
	knifeContent := fmt.Sprintf(
		`node_name "%s" # Replace with the login name you use to login to the Supermarket.
    		client_key "%s" # Define the path to wherever your client.pem file lives.  This is the key you generated when you signed up for a Chef account.
        	cookbook_path [ '%s' ] # Directory where the cookbook you're uploading resides.
		`,
		g.Config.GetString("chef_supermarket_username"),
		pemFile.Name(),
		tmpParentPath,
	)

	_, kerr := knifeFile.Write([]byte(knifeContent))
	if kerr != nil {
		return kerr
	}

	chefKey, berr := g.Config.GetBase64Decoded("chef_supermarket_key")
	if berr != nil {
		return berr
	}
	_, perr := pemFile.Write([]byte(chefKey))
	if perr != nil {
		return perr
	}

	cookbookDistCmd := fmt.Sprintf("knife cookbook site share %s %s -c %s",
		g.NextMetadata.Name,
		g.Config.GetString("chef_supermarket_type"),
		knifeFile.Name(),
	)

	if derr := utils.BashCmdExec(cookbookDistCmd, "", ""); derr != nil {
		return errors.EngineDistPackageError("knife cookbook upload to supermarket failed")
	}
	return nil
}

//private Helpers

func (g *engineChef) retrieveCurrentMetadata(gitLocalPath string) error {
	//dat, err := ioutil.ReadFile(path.Join(gitLocalPath, "metadata.rb"))
	//knife cookbook metadata -o ../ chef-mycookbook -- will generate a metadata.json file.
	if cerr := utils.CmdExec("knife", []string{"cookbook", "metadata", "-o", "../", path.Base(gitLocalPath)}, gitLocalPath, ""); cerr != nil {
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
	return nil
}

func (g *engineChef) writeNextMetadata(gitLocalPath string) error {
	return utils.CmdExec("knife", []string{"spork", "bump", path.Base(gitLocalPath), "manual", g.NextMetadata.Version, "-o", "../"}, gitLocalPath, "")
}
