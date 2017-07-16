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

	if _, berr := exec.LookPath("berkshelf"); berr != nil {
		return errors.EngineValidateToolError("berkshelf binary is missing")
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
		ioutil.WriteFile(berksfilePath, []byte("site :opscode"), 0644)
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
	return nil
}

func (g *engineChef) PackageStep() (error) {
	return nil
}

func (g *engineChef) ReleaseStep() (error) {
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