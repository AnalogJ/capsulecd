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
	"log"
)

type chefMetadata struct {
	version string
	name string
}
type engineChef struct {
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

	if _, err := os.Stat(path.Join((*g.Scm).Options().GitLocalPath, "metadata.rb")); os.IsNotExist(err) {
		return errors.EngineBuildPackageInvalid("metadata.rb file is required to process Chef cookbook")
	}

	// bump up the chef cookbook version
	merr := g.retrieveCurrentMetadata((*g.Scm).Options().GitLocalPath)
	if(merr != nil){ return merr }

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