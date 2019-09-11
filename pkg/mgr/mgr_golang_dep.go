package mgr

import (
	"fmt"
	"github.com/analogj/capsulecd/pkg/config"
	"github.com/analogj/capsulecd/pkg/errors"
	"github.com/analogj/capsulecd/pkg/pipeline"
	"github.com/analogj/capsulecd/pkg/utils"
	"net/http"
	"os"
	"os/exec"
	"path"
)

func DetectGolangDep(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	gopkgPath := path.Join(pipelineData.GitLocalPath, "Gopkg.toml")
	return utils.FileExists(gopkgPath)
}


type mgrGolangDep struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}


func (m *mgrGolangDep) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrGolangDep) MgrValidateTools() error {
	if _, kerr := exec.LookPath("dep"); kerr != nil {
		return errors.EngineValidateToolError("dep binary is missing")
	}
	return nil
}

func (m *mgrGolangDep) MgrAssembleStep() error {
	if !utils.FileExists(path.Join(m.PipelineData.GitLocalPath, "Gopkg.toml")) {
		return errors.EngineBuildPackageInvalid("Gopkg.toml file is required to process Golang/Dep package")
	}

	return nil
}

func (m *mgrGolangDep) MgrDependenciesStep(currentMetadata interface{}, nextMetadata interface{}) error {
	// the go source has already been downloaded. lets make sure all its dependencies are available.

	updatedEnv := []string{
		fmt.Sprintf("GOPATH=%s:/go", m.PipelineData.GolangGoPath),
		fmt.Sprintf("PATH=%s/bin:/go/bin", m.PipelineData.GolangGoPath),

	}
	if cerr := utils.BashCmdExec("dep ensure -v", m.PipelineData.GitLocalPath, updatedEnv, ""); cerr != nil {
		return errors.EngineTestDependenciesError("dep ensure failed. Check dep dependencies")
	}

	return nil
}

func (m *mgrGolangDep) MgrPackageStep(currentMetadata interface{}, nextMetadata interface{}) error {
	if !m.Config.GetBool("mgr_keep_lock_file") {
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "Gopkg.lock"))
	}
	return nil
}


func (m *mgrGolangDep) MgrDistStep(currentMetadata interface{}, nextMetadata interface{}) error {
	// no real packaging for golang.
	// libraries are stored in version control.
	return nil
}
