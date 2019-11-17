package mgr

import (
	"fmt"
	"github.com/analogj/capsulecd/pkg/config"
	"github.com/analogj/capsulecd/pkg/errors"
	"github.com/analogj/capsulecd/pkg/pipeline"
	"github.com/analogj/capsulecd/pkg/utils"
	"net/http"
	"os"
	"path"
	"strings"
)

func DetectGolangMod(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	gomodPath := path.Join(pipelineData.GitLocalPath, "go.mod")
	return utils.FileExists(gomodPath)
}


type mgrGolangMod struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}


func (m *mgrGolangMod) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrGolangMod) MgrValidateTools() error {
	//if _, kerr := exec.LookPath("dep"); kerr != nil {
	//	return errors.EngineValidateToolError("dep binary is missing")
	//}
	return nil
}

func (m *mgrGolangMod) MgrAssembleStep() error {
	if !utils.FileExists(path.Join(m.PipelineData.GitLocalPath, "go.mod")) {
		return errors.EngineBuildPackageInvalid("go.mod file is required to process Golang package")
	}

	return nil
}

func (m *mgrGolangMod) MgrDependenciesStep(currentMetadata interface{}, nextMetadata interface{}) error {
	// the go source has already been downloaded. lets make sure all its dependencies are available.

	currentEnv := os.Environ()
	updatedEnv := []string{
		fmt.Sprintf("GOPATH=%s", m.PipelineData.GolangGoPath),
	}

	for i := range currentEnv {
		if strings.HasPrefix(currentEnv[i], "GOPATH="){
			//skip
			continue
		} else if strings.HasPrefix(currentEnv[i], "PATH=") {
			updatedEnv = append(updatedEnv, fmt.Sprintf("PATH=%s/bin:%s", m.PipelineData.GolangGoPath, currentEnv[i]))
		} else {
			//add all environmental variables that are not GOPATH
			updatedEnv = append(updatedEnv, currentEnv[i])
		}
	}
	if cerr := utils.BashCmdExec("go mod vendor", m.PipelineData.GitLocalPath, updatedEnv, ""); cerr != nil {
		return errors.EngineTestDependenciesError("go mod vendor failed. Check dependencies")
	}

	return nil
}

func (m *mgrGolangMod) MgrPackageStep(currentMetadata interface{}, nextMetadata interface{}) error {
	if !m.Config.GetBool("mgr_keep_lock_file") {
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "go.sum"))
	}
	return nil
}


func (m *mgrGolangMod) MgrDistStep(currentMetadata interface{}, nextMetadata interface{}) error {
	// no real packaging for golang.
	// libraries are stored in version control.
	return nil
}
