package mgr

import (
	"github.com/analogj/capsulecd/pkg/pipeline"
	"net/http"
	"path"
	"os/exec"
	"github.com/analogj/capsulecd/pkg/errors"
	"os"
	"io/ioutil"
	"fmt"
	"github.com/analogj/capsulecd/pkg/config"
	"github.com/analogj/capsulecd/pkg/utils"
)

func DetectNodeYarn(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	//theres no way to automatically determine if a project was created via Yarn (vs NPM)
	return false
}


type mgrNodeYarn struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}


func (m *mgrNodeYarn) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrNodeYarn) MgrValidateTools() error {
	if _, kerr := exec.LookPath("yarn"); kerr != nil {
		return errors.EngineValidateToolError("yarn binary is missing")
	}
	return nil
}

func (m *mgrNodeYarn) MgrAssembleStep() error {
	//validate that the npm package.json file exists
	if !utils.FileExists(path.Join(m.PipelineData.GitLocalPath, "package.json")) {
		return errors.EngineBuildPackageInvalid("package.json file is required to process Node package")
	}

	return nil
}

func (m *mgrNodeYarn) MgrDependenciesStep(currentMetadata interface{}, nextMetadata interface{}) error {
	// the module has already been downloaded. lets make sure all its dependencies are available.
	if derr := utils.BashCmdExec("yarn install --non-interactive", m.PipelineData.GitLocalPath, nil, ""); derr != nil {
		return errors.EngineTestDependenciesError("yarn install failed. Check module dependencies")
	}

	return nil
}

func (m *mgrNodeYarn) MgrPackageStep(currentMetadata interface{}, nextMetadata interface{}) error {
	if !m.Config.GetBool("mgr_keep_lock_file") {
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "npm-shrinkwrap.json"))
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "package-lock.json"))
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "yarn.lock"))
	}
	return nil
}


func (m *mgrNodeYarn) MgrDistStep(currentMetadata interface{}, nextMetadata interface{}) error {
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

	//TODO: is it worth using the Yarn publish command as well?
	npmPublishCmd := fmt.Sprintf("npm --userconfig %s publish .", npmrcFile.Name())
	derr := utils.BashCmdExec(npmPublishCmd, m.PipelineData.GitLocalPath, nil, "")
	if derr != nil {
		return errors.MgrDistPackageError("npm publish failed. Check log for exact error")
	}
	return nil
}

