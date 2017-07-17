package engine

import (
	"capsulecd/lib/scm"
	"os"
	"path"
	"capsulecd/lib/errors"
	"io/ioutil"
	"os/exec"
	"capsulecd/lib/utils"
	"encoding/json"
	"capsulecd/lib/config"
	"fmt"
)

type chefMetadata struct {
	Version string `json:"version"`
	Name string `json:"name"`
}
type engineChef struct {
	*EngineBase

	Scm *scm.Scm
	CurrentMetadata *chefMetadata
	NextMetadata *chefMetadata
}

func (g *engineChef) ValidateTools() (error) {
	//a chefdk like environment needs to be available for this Engine
	if _, kerr := exec.LookPath("knife"); kerr != nil {
		return errors.EngineValidateToolError("knife binary is missing")
	}

	if _, berr := exec.LookPath("berks"); berr != nil {
		return errors.EngineValidateToolError("berkshelf binary is missing")
	}

	if _, berr := exec.LookPath("bundle"); berr != nil {
		return errors.EngineValidateToolError("bundler binary is missing")
	}

	return nil
}

func (g *engineChef) Init(sourceScm *scm.Scm) (error) {
	g.Scm = sourceScm
	g.CurrentMetadata = new(chefMetadata)
	g.NextMetadata = new(chefMetadata)
	return nil
}

func (g *engineChef) BuildStep() (error) {
	//validate that the chef metadata.rb file exists

	if !utils.FileExists(path.Join((*g.Scm).Options().GitLocalPath, "metadata.rb")) {
		return errors.EngineBuildPackageInvalid("metadata.rb file is required to process Chef cookbook")
	}

	// bump up the chef cookbook version
	merr := g.retrieveCurrentMetadata((*g.Scm).Options().GitLocalPath)
	if(merr != nil){ return merr }

	perr := g.populateNextMetadata()
	if(perr != nil){ return perr }

	nerr := g.writeNextMetadata((*g.Scm).Options().GitLocalPath)
	if(nerr != nil){ return nerr }

	// TODO: check if this cookbook name and version already exist.
	// check for/create any required missing folders/files
	// Berksfile.lock and Gemfile.lock are not required to be commited, but they should be.
	rakefilePath := path.Join((*g.Scm).Options().GitLocalPath, "Rakefile")
	if !utils.FileExists(rakefilePath){
		ioutil.WriteFile(rakefilePath, []byte("task :test"), 0644)
	}
	berksfilePath := path.Join((*g.Scm).Options().GitLocalPath, "Berksfile")
	if !utils.FileExists(berksfilePath){
		ioutil.WriteFile(berksfilePath, []byte("site \"https://supermarket.chef.io\""), 0644)
	}
	gemfilePath := path.Join((*g.Scm).Options().GitLocalPath, "Gemfile")
	if !utils.FileExists(gemfilePath){
		ioutil.WriteFile(gemfilePath, []byte("source \"https://rubygems.org\""), 0644)
	}
	specPath := path.Join((*g.Scm).Options().GitLocalPath, "spec")
	if !utils.FileExists(specPath){
		os.MkdirAll(specPath, 0777)
	}

	//unless File.exist?(@source_git_local_path + '/.gitignore')
	//TODO: CapsuleCD::GitUtils.create_gitignore(@source_git_local_path, ['ChefCookbook'])
	//end
	return nil
}

func (g *engineChef) TestStep() (error) {
	// the cookbook has already been downloaded. lets make sure all its dependencies are available.
	cerr := utils.CmdExec("berks", []string{"install"}, (*g.Scm).Options().GitLocalPath, "")
	if(cerr != nil) {return errors.EngineTestDependenciesError("berks install failed. Check cookbook dependencies")}

	//download all its gem dependencies
	berr := utils.CmdExec("bundle", []string{"install"}, (*g.Scm).Options().GitLocalPath, "")
	if(berr != nil) {return errors.EngineTestDependenciesError("bundle install failed. Check Gem dependencies")}

	//skip the test command if disabled
	if(config.GetBool("engine_disable_test")){
		return nil
	}

	//run test command
	var testCmd string
	if config.IsSet("engine_cmd_test") {
		testCmd = config.GetString("engine_cmd_test")
	} else{
		testCmd = "rake test"
	}
	terr := utils.BashCmdExec(testCmd, (*g.Scm).Options().GitLocalPath, "")
	if(terr != nil){return errors.EngineTestRunnerError(fmt.Sprintf("Test command (%s) failed. Check log for more details.", testCmd))}
	return nil
}

func (g *engineChef) PackageStep() (error) {
	// commit changes to the cookbook. (test run occurs before this, and it should clean up any instrumentation files, created,
	// as they will be included in the commmit and any release artifacts)

	cerr := utils.GitCommit((*g.Scm).Options().GitLocalPath,fmt.Sprintf("(v%s) Automated packaging of release by CapsuleCD", g.NextMetadata.Version))
	if(cerr != nil){return cerr}
	_, terr := utils.GitTag((*g.Scm).Options().GitLocalPath, fmt.Sprintf("v%s", g.NextMetadata.Version))
	if(terr != nil){return terr}

	//TODO: do something with the tag output.
	return nil
}

func (g *engineChef) DistStep() (error) {
	return nil
}

//private Helpers

func (g *engineChef) retrieveCurrentMetadata(gitLocalPath string) (error) {
	//dat, err := ioutil.ReadFile(path.Join(gitLocalPath, "metadata.rb"))
	//knife cookbook metadata -o ../ chef-mycookbook -- will generate a metadata.json file.
	cerr := utils.CmdExec("knife", []string{"cookbook", "metadata", "-o", "../", path.Base(gitLocalPath)}, gitLocalPath, "")
	if(cerr != nil){ return cerr }
	defer os.Remove(path.Join(gitLocalPath, "metadata.json"))

	//read metadata.json file.
	metadataContent, rerr := ioutil.ReadFile(path.Join(gitLocalPath, "metadata.json"))
	if(rerr != nil) { return rerr }

	uerr := json.Unmarshal(metadataContent, g.CurrentMetadata )
	if(uerr != nil) { return uerr }

	return nil
}

func (g *engineChef) populateNextMetadata() error {

	nextVersion, err := g.BumpVersion(g.CurrentMetadata.Version)
	if err != nil {return err}

	g.NextMetadata.Version = nextVersion
	g.NextMetadata.Name = g.CurrentMetadata.Name
	return nil
}

func (g *engineChef) writeNextMetadata(gitLocalPath string) (error) {
	return utils.CmdExec("knife", []string{"spork", "bump", path.Base(gitLocalPath), "manual", g.NextMetadata.Version, "-o", "../"}, gitLocalPath, "")
}