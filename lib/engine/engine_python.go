package engine

import (
	"capsulecd/lib/config"
	"capsulecd/lib/errors"
	"capsulecd/lib/pipeline"
	"capsulecd/lib/scm"
	"capsulecd/lib/utils"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

type pythonMetadata struct {
	Version string
}
type enginePython struct {
	*EngineBase

	PipelineData    *pipeline.Data
	Scm             scm.Scm //Interface
	CurrentMetadata *pythonMetadata
	NextMetadata    *pythonMetadata
}

func (g *enginePython) ValidateTools() error {
	if _, kerr := exec.LookPath("tox"); kerr != nil {
		return errors.EngineValidateToolError("tox binary is missing")
	}

	if _, berr := exec.LookPath("python"); berr != nil {
		return errors.EngineValidateToolError("python binary is missing")
	}

	if _, berr := exec.LookPath("twine"); berr != nil {
		return errors.EngineValidateToolError("twine binary is missing")
	}

	return nil
}

func (g *enginePython) Init(pipelineData *pipeline.Data, sourceScm scm.Scm) error {
	g.Scm = sourceScm
	g.PipelineData = pipelineData
	g.CurrentMetadata = new(pythonMetadata)
	g.NextMetadata = new(pythonMetadata)
	return nil
}

func (g *enginePython) BuildStep() error {
	//validate that the python setup.py file exists
	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "setup.py")) {
		return errors.EngineBuildPackageInvalid("setup.py file is required to process Python package")
	}

	// check for/create required VERSION file
	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "VERSION")) {
		ioutil.WriteFile(path.Join(g.PipelineData.GitLocalPath, "VERSION"),
			[]byte("0.0.0"),
			0644,
		)
	}

	// bump up the version here.
	// since there's no standardized way to bump up the version in the setup.py file, we're going to assume that the version
	// is specified in a VERSION file in the root of the source repository
	// this is option #4 in the python packaging guide:
	// https://packaging.python.org/en/latest/single_source_version/#single-sourcing-the-version
	//
	// additional packaging structures, like those listed below, may also be supported in the future.
	// http://stackoverflow.com/a/7071358/1157633

	if merr := g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath); merr != nil {
		return merr
	}

	if perr := g.populateNextMetadata(); perr != nil {
		return perr
	}

	if nerr := g.writeNextMetadata(g.PipelineData.GitLocalPath); nerr != nil {
		return nerr
	}

	// make sure the package testing manager is available.
	// there is a standardized way to test packages (python setup.py tests), however for automation tox is preferred
	// because of virtualenv and its support for multiple interpreters.
	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "tox.ini")) {
		toxIniContent := `# Tox (http://tox.testrun.org/) is a tool for running tests
			# in multiple virtualenvs. This configuration file will run the
			# test suite on all supported python versions. To use it, "pip install tox"
			# and then run "tox" from this directory.
			[tox]
			envlist = py27
			usedevelop = True
			[testenv]
			commands = python setup.py test
			deps =
			  -rrequirements.txt
			TOX
			`

		ioutil.WriteFile(path.Join(g.PipelineData.GitLocalPath, "tox.ini"),
			[]byte(toxIniContent),
			0644,
		)
	}

	// check for/create any required missing folders/files
	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "requirements.txt")) {
		ioutil.WriteFile(path.Join(g.PipelineData.GitLocalPath, "requirements.txt"),
			[]byte(""),
			0644,
		)
	}

	os.MkdirAll(path.Join(g.PipelineData.GitLocalPath, "tests"), 0644)

	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "tests", "__init__.py")) {
		ioutil.WriteFile(path.Join(g.PipelineData.GitLocalPath, "tests", "__init__.py"),
			[]byte(""),
			0644,
		)
	}

	//TODO: add gitignore content.
	//if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, ".gitignore")) {
	//	ioutil.WriteFile(path.Join(g.PipelineData.GitLocalPath, ".gitignore"),
	//		[]byte(""),
	//		0644,
	//	)
	//}
	return nil
}

func (g *enginePython) TestStep() error {
	//run test command
	var testCmd string
	if config.IsSet("engine_cmd_test") {
		testCmd = config.GetString("engine_cmd_test")
	} else {
		testCmd = "tox"
	}
	//running tox will install all dependencies in a virtual env, and then run unit tests.
	terr := utils.BashCmdExec(testCmd, g.PipelineData.GitLocalPath, "")
	if terr != nil {
		return errors.EngineTestRunnerError(fmt.Sprintf("Test command (%s) failed. Check log for more details.", testCmd))
	}
	return nil
}

func (g *enginePython) PackageStep() error {
	// commit changes to the cookbook. (test run occurs before this, and it should clean up any instrumentation files, created,
	// as they will be included in the commmit and any release artifacts)

	if cerr := utils.GitCommit(g.PipelineData.GitLocalPath, fmt.Sprintf("(v%s) Automated packaging of release by CapsuleCD", g.NextMetadata.Version)); cerr != nil {
		return cerr
	}
	tagCommit, terr := utils.GitTag(g.PipelineData.GitLocalPath, fmt.Sprintf("v%s", g.NextMetadata.Version))
	if terr != nil {
		return terr
	}

	g.PipelineData.ReleaseCommit = tagCommit
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}

func (g *enginePython) DistStep() error {
	if !config.IsSet("pypi_username") || !config.IsSet("pypi_password") {
		return errors.EngineDistCredentialsMissing("Cannot deploy python package to pypi/warehouse, credentials missing")
	}


	pypircFile, _ := ioutil.TempFile("",".pypirc")
	defer os.Remove(pypircFile.Name())

	// write the .pypirc config jfile.
	var repository string
	if(config.IsSet("pypi_repository")){
		repository = config.GetString("pypi_repository")
	} else {
		repository = "https://upload.pypi.org/legacy/"
	}
	pypircContent := fmt.Sprintf(
		`[distutils]
		index-servers=pypi

		[pypi]
		repository = %s
		username = %s
		password = %s
		`,
		repository,
		config.GetString("pypi_username"),
		config.GetString("pypi_password"),
	)

	if _, perr := pypircFile.Write([]byte(pypircContent)); perr != nil{
		return perr
	}

	pythonDistCmd := "python setup.py sdist"
	if derr := utils.BashCmdExec(pythonDistCmd, g.PipelineData.GitLocalPath, ""); derr != nil {
		return errors.EngineDistPackageError("python setup.py sdist failed")
	}

	// using twine instead of setup.py (it supports HTTPS.)https://python-packaging-user-guide.readthedocs.org/en/latest/distributing/#uploading-your-project-to-pypi
	pypiUploadCmd := fmt.Sprint("twine upload --config-file %s  dist/*",
		pypircFile.Name(),
	)

	if uerr := utils.BashCmdExec(pypiUploadCmd, g.PipelineData.GitLocalPath, ""); uerr != nil {
		return errors.EngineDistPackageError("twine package upload failed. Check log for exact error")
	}
	return nil
}

//private Helpers

func (g *enginePython) retrieveCurrentMetadata(gitLocalPath string) error {
	//read metadata.json file.
	versionContent, rerr := ioutil.ReadFile(path.Join(gitLocalPath, "VERSION"))
	if rerr != nil {
		return rerr
	}
	g.CurrentMetadata.Version = strings.TrimSpace(versionContent)
	return nil
}

func (g *enginePython) populateNextMetadata() error {

	nextVersion, err := g.BumpVersion(g.CurrentMetadata.Version)
	if err != nil {
		return err
	}

	g.NextMetadata.Version = nextVersion
	return nil
}

func (g *enginePython) writeNextMetadata(gitLocalPath string) error {
	return ioutil.WriteFile(path.Join(gitLocalPath, "VERSION"), []byte(g.NextMetadata.Version), 0644)
}
