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
	"path/filepath"
	"capsulecd/lib/pipeline"
)

type chefMetadata struct {
	Version string `json:"version"`
	Name string `json:"name"`
}
type engineChef struct {
	*EngineBase

	PipelineData *pipeline.PipelineData
	Scm scm.Scm //Interface
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

func (g *engineChef) Init(pipelineData *pipeline.PipelineData, sourceScm scm.Scm) (error) {
	g.Scm = sourceScm
	g.PipelineData = pipelineData
	g.CurrentMetadata = new(chefMetadata)
	g.NextMetadata = new(chefMetadata)
	return nil
}

func (g *engineChef) BuildStep() (error) {
	//validate that the chef metadata.rb file exists

	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "metadata.rb")) {
		return errors.EngineBuildPackageInvalid("metadata.rb file is required to process Chef cookbook")
	}

	// bump up the chef cookbook version
	merr := g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath)
	if(merr != nil){ return merr }

	perr := g.populateNextMetadata()
	if(perr != nil){ return perr }

	nerr := g.writeNextMetadata(g.PipelineData.GitLocalPath)
	if(nerr != nil){ return nerr }

	// TODO: check if this cookbook name and version already exist.
	// check for/create any required missing folders/files
	// Berksfile.lock and Gemfile.lock are not required to be commited, but they should be.
	rakefilePath := path.Join(g.PipelineData.GitLocalPath, "Rakefile")
	if !utils.FileExists(rakefilePath){
		ioutil.WriteFile(rakefilePath, []byte("task :test"), 0644)
	}
	berksfilePath := path.Join(g.PipelineData.GitLocalPath, "Berksfile")
	if !utils.FileExists(berksfilePath){
		ioutil.WriteFile(berksfilePath, []byte("site \"https://supermarket.chef.io\""), 0644)
	}
	gemfilePath := path.Join(g.PipelineData.GitLocalPath, "Gemfile")
	if !utils.FileExists(gemfilePath){
		ioutil.WriteFile(gemfilePath, []byte("source \"https://rubygems.org\""), 0644)
	}
	specPath := path.Join(g.PipelineData.GitLocalPath, "spec")
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
	cerr := utils.CmdExec("berks", []string{"install"}, g.PipelineData.GitLocalPath, "")
	if(cerr != nil) {return errors.EngineTestDependenciesError("berks install failed. Check cookbook dependencies")}

	//download all its gem dependencies
	berr := utils.CmdExec("bundle", []string{"install"}, g.PipelineData.GitLocalPath, "")
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
	terr := utils.BashCmdExec(testCmd, g.PipelineData.GitLocalPath, "")
	if(terr != nil){return errors.EngineTestRunnerError(fmt.Sprintf("Test command (%s) failed. Check log for more details.", testCmd))}
	return nil
}

func (g *engineChef) PackageStep() (error) {
	// commit changes to the cookbook. (test run occurs before this, and it should clean up any instrumentation files, created,
	// as they will be included in the commmit and any release artifacts)

	cerr := utils.GitCommit(g.PipelineData.GitLocalPath,fmt.Sprintf("(v%s) Automated packaging of release by CapsuleCD", g.NextMetadata.Version))
	if(cerr != nil){return cerr}
	tagCommit, terr := utils.GitTag(g.PipelineData.GitLocalPath, fmt.Sprintf("v%s", g.NextMetadata.Version))
	if(terr != nil){return terr}

	g.PipelineData.ReleaseCommit = tagCommit
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}

func (g *engineChef) DistStep() (error) {
	pemPath, _ := filepath.Abs("~/client.pem")
	knifePath, _ := filepath.Abs("~/knife.rb")

	if(!config.IsSet("chef_supermarket_username") || !config.IsSet("chef_supermarket_key")){
		return errors.EngineDistCredentialsMissing("Cannot deploy cookbook to supermarket, credentials missing")
	}

	// knife is really sensitive to folder names. The cookbook name MUST match the folder name otherwise knife throws up
	// when doing a knife cookbook share. So we're going to make a new tmp directory, create a subdirectory with the EXACT
	// cookbook name, and then copy the cookbook contents into it. Yeah yeah, its pretty nasty, but blame Chef.
	tmpParentPath, terr := ioutil.TempDir("","")
	if(terr != nil){return terr}
	defer os.RemoveAll(tmpParentPath)

	tmpLocalPath := path.Join(tmpParentPath, g.NextMetadata.Name)
	cerr := utils.CopyDir(g.PipelineData.GitLocalPath, tmpLocalPath)
	if(cerr != nil){return cerr}

	// write the knife.rb config jfile.
	knifeContent := fmt.Sprintf(
		`node_name "%s" # Replace with the login name you use to login to the Supermarket.
    		client_key "%s" # Define the path to wherever your client.pem file lives.  This is the key you generated when you signed up for a Chef account.
        	cookbook_path [ '%s' ] # Directory where the cookbook you're uploading resides.
		`,
		config.GetString("chef_supermarket_username"),
		pemPath,
		tmpParentPath,
	)

	kerr := ioutil.WriteFile(knifePath, []byte(knifeContent), 0644)
	if(kerr != nil){return kerr}

	perr := ioutil.WriteFile(pemPath, []byte(config.GetBase64Decoded("chef_supermarket_key")), 0644)
	if(perr != nil){return perr}

	cookbookDistCmd := fmt.Sprintf("knife cookbook site share %s %s -c %s",
		g.NextMetadata.Name,
		config.GetString("chef_supermarket_type"),
		knifePath,
	)

	derr := utils.BashCmdExec(cookbookDistCmd,"","")
	if(derr != nil){
		return errors.EngineDistPackageError("knife cookbook upload to supermarket failed")
	}
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