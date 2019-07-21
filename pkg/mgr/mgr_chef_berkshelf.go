package mgr

import (
	"github.com/analogj/capsulecd/pkg/pipeline"
	"net/http"
	"github.com/analogj/capsulecd/pkg/config"
	"os/exec"
	"github.com/analogj/capsulecd/pkg/errors"
	"path"
	"io/ioutil"
	"github.com/analogj/capsulecd/pkg/utils"
	"os"
	"fmt"
	"github.com/analogj/capsulecd/pkg/metadata"
)

func DetectChefBerkshelf(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	berksfilePath := path.Join(pipelineData.GitLocalPath, "Berksfile")
	return utils.FileExists(berksfilePath)
}


type mgrChefBerkshelf struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}


func (m *mgrChefBerkshelf) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrChefBerkshelf) MgrValidateTools() error {
	//a chef/berkshelf like environment needs to be available for this Engine
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
	return nil
}

func (m *mgrChefBerkshelf) MgrAssembleStep() error {

	berksfilePath := path.Join(m.PipelineData.GitLocalPath, "Berksfile")
	if !utils.FileExists(berksfilePath) {
		ioutil.WriteFile(berksfilePath, []byte(utils.StripIndent(
			`source "https://supermarket.chef.io"
		metadata
		`)), 0644)
	}
	gemfilePath := path.Join(m.PipelineData.GitLocalPath, "Gemfile")
	if !utils.FileExists(gemfilePath) {
		ioutil.WriteFile(gemfilePath, []byte(`source "https://rubygems.org"`), 0644)
	}
	return nil
}

func (m *mgrChefBerkshelf) MgrDependenciesStep(currentMetadata interface{}, nextMetadata interface{}) error {
	// the cookbook has already been downloaded. lets make sure all its dependencies are available.
	if cerr := utils.BashCmdExec("berks install", m.PipelineData.GitLocalPath, nil, ""); cerr != nil {
		return errors.EngineTestDependenciesError("berks install failed. Check cookbook dependencies")
	}

	//download all its gem dependencies
	if berr := utils.BashCmdExec("bundle install", m.PipelineData.GitLocalPath, nil, ""); berr != nil {
		return errors.EngineTestDependenciesError("bundle install failed. Check Gem dependencies")
	}
	return nil
}

func (m *mgrChefBerkshelf) MgrPackageStep(currentMetadata interface{}, nextMetadata interface{}) error {
	if !m.Config.GetBool("mgr_keep_lock_file") {
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "Berksfile.lock"))
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "Gemfile.lock"))
	}
	return nil
}


func (m *mgrChefBerkshelf) MgrDistStep(currentMetadata interface{}, nextMetadata interface{}) error {
	if !m.Config.IsSet("chef_supermarket_username") || !m.Config.IsSet("chef_supermarket_key") {
		return errors.MgrDistCredentialsMissing("Cannot deploy cookbook to supermarket, credentials missing")
	}

	// knife is really sensitive to folder names. The cookbook name MUST match the folder name otherwise knife throws up
	// when doing a knife cookbook share. So we're going to make a new tmp directory, create a subdirectory with the EXACT
	// cookbook name, and then copy the cookbook contents into it. Yeah yeah, its pretty nasty, but blame Chef.
	tmpParentPath, terr := ioutil.TempDir("", "")
	if terr != nil {
		return terr
	}
	defer os.RemoveAll(tmpParentPath)

	tmpLocalPath := path.Join(tmpParentPath, nextMetadata.(*metadata.ChefMetadata).Name)
	if cerr := utils.CopyDir(m.PipelineData.GitLocalPath, tmpLocalPath); cerr != nil {
		return cerr
	}

	pemFile, _ := ioutil.TempFile("", "client.pem")
	defer os.Remove(pemFile.Name())
	knifeFile, _ := ioutil.TempFile("", "knife.rb")
	defer os.Remove(knifeFile.Name())

	// write the knife.rb config jfile.
	knifeContent := fmt.Sprintf(utils.StripIndent(
		`node_name "%s" # Replace with the login name you use to login to the Supermarket.
    		client_key "%s" # Define the path to wherever your client.pem file lives.  This is the key you generated when you signed up for a Chef account.
        	cookbook_path [ '%s' ] # Directory where the cookbook you're uploading resides.
		`),
		m.Config.GetString("chef_supermarket_username"),
		pemFile.Name(),
		tmpParentPath,
	)

	_, kerr := knifeFile.Write([]byte(knifeContent))
	if kerr != nil {
		return kerr
	}

	chefKey, berr := m.Config.GetBase64Decoded("chef_supermarket_key")
	if berr != nil {
		return berr
	}
	_, perr := pemFile.Write([]byte(chefKey))
	if perr != nil {
		return perr
	}

	cookbookDistCmd := fmt.Sprintf("knife cookbook site share %s %s -c %s",
		nextMetadata.(*metadata.ChefMetadata).Name,
		m.Config.GetString("chef_supermarket_type"),
		knifeFile.Name(),
	)

	if derr := utils.BashCmdExec(cookbookDistCmd, "", nil, ""); derr != nil {
		return errors.MgrDistPackageError("knife cookbook upload to supermarket failed")
	}
	return nil
}
