package mgr

import (
	"capsulecd/pkg/pipeline"
	"net/http"
	"path"
	"os/exec"
	"capsulecd/pkg/errors"
	"os"
	"capsulecd/pkg/utils"
	"capsulecd/pkg/config"
)

func DetectGolangGlide(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	glideyamlPath := path.Join(pipelineData.GitLocalPath, "glide.yaml")
	return utils.FileExists(glideyamlPath)
}


type mgrGolangGlide struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}


func (m *mgrGolangGlide) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrGolangGlide) MgrValidateTools() error {
	if _, kerr := exec.LookPath("glide"); kerr != nil {
		return errors.EngineValidateToolError("glide binary is missing")
	}
	return nil
}

func (m *mgrGolangGlide) MgrAssembleStep() error {
	if !utils.FileExists(path.Join(m.PipelineData.GitLocalPath, "glide.yaml")) {
		return errors.EngineBuildPackageInvalid("glide.yaml file is required to process Golang/Glide package")
	}
	return nil
}

func (m *mgrGolangGlide) MgrDependenciesStep(currentMetadata interface{}, nextMetadata interface{}) error {
	// the go source has already been downloaded. lets make sure all its dependencies are available.
	if cerr := utils.BashCmdExec("glide install", m.PipelineData.GitLocalPath, nil, ""); cerr != nil {
		return errors.EngineTestDependenciesError("glide install failed. Check glide dependencies")
	}

	return nil
}

func (m *mgrGolangGlide) MgrPackageStep(currentMetadata interface{}, nextMetadata interface{}) error {
	if !m.Config.GetBool("mgr_keep_lock_file") {
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "glide.lock"))
	}
	return nil
}


func (m *mgrGolangGlide) MgrDistStep(currentMetadata interface{}, nextMetadata interface{}) error {
	// no real packaging for golang.
	// libraries are stored in version control.
	return nil
}