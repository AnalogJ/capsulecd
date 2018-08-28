package mgr

import (
	"capsulecd/pkg/pipeline"
	"net/http"
	"os/exec"
	"capsulecd/pkg/errors"
	"path"
	"os"
	"io/ioutil"
	"fmt"
	"capsulecd/pkg/config"
	"capsulecd/pkg/utils"
)

func DetectPythonPip(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	//theres no way to automatically determine if a project was created via Yarn (vs NPM)
	return false
}


type mgrPythonPip struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}


func (m *mgrPythonPip) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrPythonPip) MgrValidateTools() error {
	if _, berr := exec.LookPath("twine"); berr != nil {
		return errors.EngineValidateToolError("twine binary is missing")
	}
	if _, berr := exec.LookPath("pip"); berr != nil {
		return errors.EngineValidateToolError("pip binary is missing")
	}
	return nil
}

func (m *mgrPythonPip) MgrAssembleStep() error {
	// check for/create any required missing folders/files
	if !utils.FileExists(path.Join(m.PipelineData.GitLocalPath, "requirements.txt")) {
		ioutil.WriteFile(path.Join(m.PipelineData.GitLocalPath, "requirements.txt"),
			[]byte(""),
			0644,
		)
	}

	return nil
}

func (m *mgrPythonPip) MgrDependenciesStep(currentMetadata interface{}, nextMetadata interface{}) error {
	return nil //dependencies are installed as part of Tox.
}

func (m *mgrPythonPip) MgrPackageStep(currentMetadata interface{}, nextMetadata interface{}) error {
	if !m.Config.GetBool("mgr_keep_lock_file") {
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "requirements.txt"))
	}
	return nil
}


func (m *mgrPythonPip) MgrDistStep(currentMetadata interface{}, nextMetadata interface{}) error {
	if !m.Config.IsSet("pypi_username") || !m.Config.IsSet("pypi_password") {
		return errors.MgrDistCredentialsMissing("Cannot deploy python package to pypi/warehouse, credentials missing")
	}

	pypircFile, _ := ioutil.TempFile("", ".pypirc")
	defer os.Remove(pypircFile.Name())

	// write the .pypirc config jfile.
	pypircContent := fmt.Sprintf(utils.StripIndent(
		`[distutils]
		index-servers=pypi

		[pypi]
		repository = %s
		username = %s
		password = %s
		`),
		m.Config.GetString("pypi_repository"),
		m.Config.GetString("pypi_username"),
		m.Config.GetString("pypi_password"),
	)

	if _, perr := pypircFile.Write([]byte(pypircContent)); perr != nil {
		return perr
	}

	pythonDistCmd := "python setup.py sdist"
	if derr := utils.BashCmdExec(pythonDistCmd, m.PipelineData.GitLocalPath, nil, ""); derr != nil {
		return errors.MgrDistPackageError("python setup.py sdist failed")
	}

	// using twine instead of setup.py (it supports HTTPS.)https://python-packaging-user-guide.readthedocs.org/en/latest/distributing/#uploading-your-project-to-pypi
	pypiUploadCmd := fmt.Sprintf("twine upload --config-file %s  dist/*",
		pypircFile.Name(),
	)

	if uerr := utils.BashCmdExec(pypiUploadCmd, m.PipelineData.GitLocalPath, nil, ""); uerr != nil {
		return errors.MgrDistPackageError("twine package upload failed. Check log for exact error")
	}
	return nil
}
