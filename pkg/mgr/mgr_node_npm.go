package mgr

import (
	"capsulecd/pkg/pipeline"
	"net/http"
	"path"
	"os/exec"
	"capsulecd/pkg/errors"
	"os"
	"capsulecd/pkg/config"
	"capsulecd/pkg/utils"
	"io/ioutil"
	"fmt"
)

func DetectNodeNpm(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	npmPath := path.Join(pipelineData.GitLocalPath, "package.json")
	return utils.FileExists(npmPath)
}


type mgrNodeNpm struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}


func (m *mgrNodeNpm) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrNodeNpm) MgrValidateTools() error {
	if _, kerr := exec.LookPath("npm"); kerr != nil {
		return errors.EngineValidateToolError("npm binary is missing")
	}
	return nil
}

func (m *mgrNodeNpm) MgrAssembleStep() error {
	//validate that the npm package.json file exists
	if !utils.FileExists(path.Join(m.PipelineData.GitLocalPath, "package.json")) {
		return errors.EngineBuildPackageInvalid("package.json file is required to process Node package")
	}

	return nil
}

func (m *mgrNodeNpm) MgrDependenciesStep(currentMetadata interface{}, nextMetadata interface{}) error {
	// the module has already been downloaded. lets make sure all its dependencies are available.
	if derr := utils.BashCmdExec("npm install", m.PipelineData.GitLocalPath, nil, ""); derr != nil {
		return errors.EngineTestDependenciesError("npm install failed. Check module dependencies")
	}

	// create a shrinkwrap file.
	if derr := utils.BashCmdExec("npm shrinkwrap", m.PipelineData.GitLocalPath, nil, ""); derr != nil {
		return errors.EngineTestDependenciesError("npm shrinkwrap failed. Check log for exact error")
	}
	return nil
}

func (m *mgrNodeNpm) MgrPackageStep(currentMetadata interface{}, nextMetadata interface{}) error {
	if !m.Config.GetBool("mgr_keep_lock_file") {
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "npm-shrinkwrap.json"))
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "package-lock.json"))
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "yarn.lock"))
	}
	return nil
}


func (m *mgrNodeNpm) MgrDistStep(currentMetadata interface{}, nextMetadata interface{}) error {
	if !m.Config.IsSet("npm_auth_token") {
		return errors.MgrDistCredentialsMissing("cannot deploy page to npm, credentials missing")
	}

	npmrcFile, _ := ioutil.TempFile("", ".npmrc")
	defer os.Remove(npmrcFile.Name())

	// write the .npmrc config jfile.
	npmrcContent := fmt.Sprintf(
		"//registry.npmjs.org/:_authToken=%s",
		m.Config.GetString("npm_auth_token"),
	)

	if _, werr := npmrcFile.Write([]byte(npmrcContent)); werr != nil {
		return werr
	}

	npmPublishCmd := fmt.Sprintf("npm --userconfig %s publish .", npmrcFile.Name())
	derr := utils.BashCmdExec(npmPublishCmd, m.PipelineData.GitLocalPath, nil, "")
	if derr != nil {
		return errors.MgrDistPackageError("npm publish failed. Check log for exact error")
	}
	return nil
}
